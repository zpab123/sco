// /////////////////////////////////////////////////////////////////////////////
// 代理对应于用户，用于存储原始连接信息

package network

import (
	"encoding/json"
	"sync"
	"time"

	"github.com/pkg/errors"
	"github.com/zpab123/sco/log"
	"github.com/zpab123/sco/protocol"
	"github.com/zpab123/sco/state"
	"github.com/zpab123/sco/syncs"
)

// /////////////////////////////////////////////////////////////////////////////
// 初始化

// 变量
var (
	errState error = errors.New("状态错误")
)

// /////////////////////////////////////////////////////////////////////////////
// Agent

// 代理对应于用户，用于存储原始连接信息
type Agent struct {
	id         int32             // id 标识
	socket     *Socket           // socket
	key        string            // 握手key
	heartbeat  time.Duration     // 心跳周期
	state      *state.State      // 状态管理
	mgr        IAgentManager     // 连接管理
	lastRecv   syncs.AtomicInt64 // 上次收到数据的时间
	lastSend   syncs.AtomicInt64 // 上次发送数据时间
	packetChan chan *Packet      // 消息通道
	stopGroup  sync.WaitGroup    // 停止等待组
	scoPacket  chan *Packet      // 引擎消息通道
}

// 新建1个 *Agent 对象
// 成功：返回 *Agent nil
// 失败：返回 nil error
func NewAgent(socket *Socket) (*Agent, error) {
	// 参数效验
	if nil == socket {
		err := errors.New("参数 socket 为空")
		return nil, err
	}

	// 状态管理
	st := state.NewState()

	// 创建对象
	a := Agent{
		key:       C_F_KEY,
		heartbeat: C_F_HEARTBEAT,
		socket:    socket,
		state:     st,
	}
	a.lastRecv.Store(time.Now().UnixNano())
	a.lastSend.Store(time.Now().UnixNano())

	// 设置为初始化状态
	a.state.Set(C_AGENT_ST_INIT)

	return &a, nil
}

// -----------------------------------------------------------------------------
// 类似 fmt.Sprintf 中的打印接口

// 打印信息
func (this *Agent) String() string {
	return this.socket.String()
}

// -----------------------------------------------------------------------------
// public

// 启动
func (this *Agent) Run() {
	this.stopGroup.Add(2)
	// 发送线程
	go this.sendLoop()

	// 接收循环，这里不能 go this.recvLoop()，会导致 websocket 连接直接断开
	this.recvLoop()

	this.stopGroup.Wait()

	log.Logger.Debug("[Agent] 停止",
		log.String("ip", this.String()),
	)

	if nil != this.mgr {
		this.mgr.OnAgentStop(this)
	}

	this.state.Set(C_AGENT_ST_CLOSED)
}

// 停止
func (this *Agent) Stop() {
	if this.state.Get() == C_AGENT_ST_CLOSING {
		return
	}

	if this.state.Get() == C_AGENT_ST_CLOSED {
		return
	}

	this.state.Set(C_AGENT_ST_CLOSING)

	this.socket.Close()
}

// 设置握手 key
func (this *Agent) SetKey(k string) {
	if "" != k {
		this.key = k
	}
}

// 设置心跳 key
func (this *Agent) SetHeartbeat(h time.Duration) {
	this.heartbeat = h
}

// 设置连接管理
func (this *Agent) SetMgr(mgr IAgentManager) {
	if nil != mgr {
		this.mgr = mgr
	}
}

// 设置 id
func (this *Agent) SetId(id int32) {
	if id >= 0 {
		this.id = id
	}
}

// 获取 id
func (this *Agent) GetId() int32 {
	return this.id
}

// 设置消息通道
func (this *Agent) SetPacketChan(ch chan *Packet) {
	if ch != nil {
		this.packetChan = ch
	}
}

// 发送1个 packet 消息
func (this *Agent) SendPacket(pkt *Packet) error {
	// 状态效验
	if this.state.Get() != C_AGENT_ST_WORKING {
		return errState
	}

	this.socket.SendPacket(pkt)

	return nil
}

// 发送 []byte
func (this *Agent) SendBytes(bytes []byte) error {
	// 状态效验
	if this.state.Get() != C_AGENT_ST_WORKING {
		return errState
	}

	this.socket.SendBytes(bytes)

	return nil
}

// -----------------------------------------------------------------------------
// private

// 接收线程
func (this *Agent) recvLoop() {
	defer func() {
		this.stopGroup.Done()

		log.Logger.Debug(
			"[Agent] recvLoop 结束",
		)

		this.socket.SendPacket(nil) // 用于结束 sendLoop
	}()

	for {
		pkt, err := this.socket.RecvPacket()
		if nil != err {
			return
		}

		if nil != pkt {
			pkt.conn = this
			this.onPacket(pkt)
			continue
		}
	}
}

// 发送线程
func (this *Agent) sendLoop() {
	defer func() {
		this.stopGroup.Done()

		log.Logger.Debug(
			"[Agent] sendLoop 结束",
		)
	}()

	for {
		err := this.socket.Flush()
		if nil != err {
			break
		}

		this.lastSend.Store(time.Now().UnixNano())
	}
}

// 收到1个 pakcet
func (this *Agent) onPacket(pkt *Packet) {
	this.lastRecv.Store(time.Now().UnixNano())
	switch pkt.mid {
	case protocol.C_MID_INVALID: // 无效
		log.Logger.Debug("[Agent] 无效 packet",
			log.Uint16("mid", pkt.mid),
		)

		this.Stop()
	case protocol.C_MID_SCO: // 框架内部消息
		this.onScoPacket(pkt)
	default:
		this.handle(pkt)
	}
}

// 框架内部消息
func (this *Agent) onScoPacket(pkt *Packet) {
	switch pkt.GetSid() {
	case protocol.C_SID_HANDSHAKE_REQ: // 握手请求
		this.onHandshake(pkt.GetBody())
	case protocol.C_SID_ACK: // 握手 ACK
		this.onAck()
	case protocol.C_SID_HEARTBEAT: // 心跳
	//log.Sugar.Debug("心跳")
	default:
		// 其他消息
		log.Logger.Debug("[Agent] 无效 packet",
			log.Uint16("sid", pkt.GetSid()),
		)

		this.Stop()
	}
}

// 握手请求
func (this *Agent) onHandshake(body []byte) {
	// 状态效验
	if this.state.Get() != C_AGENT_ST_INIT {
		this.Stop()
		return
	}

	// 解码消息
	req := &protocol.HandshakeReq{}
	err := json.Unmarshal(body, req)
	if nil != err {
		this.Stop()
		return
	}

	// 握手key检查
	if this.key != req.Key {
		this.handshakeFail(protocol.C_CODE_KEY_ERR)
		return
	}

	// 其他验证

	// 握手成功
	this.handshakeOk()
}

//  握手失败
func (this *Agent) handshakeFail(code uint32) {
	defer this.Stop()
	// 返回数据
	res := &protocol.HandshakeRes{
		Code: code,
	}
	data, err := json.Marshal(res)
	if nil != err {
		log.Logger.Error(
			"[Agent] 握手失败，但服务器未返回消息：编码握手消息出错",
		)

		return
	}

	pkt := NewPacket(protocol.C_MID_SCO, protocol.C_SID_HANDSHAKE_RES)
	pkt.AppendBytes(data)
	this.socket.SendPacket(pkt) // 越过工作状态发送消息
}

//  握手成功
func (this *Agent) handshakeOk() {
	// defer log.Logger.Sync()

	hUint16 := uint16(this.heartbeat / time.Second)

	// 返回数据
	res := &protocol.HandshakeRes{
		Code:      protocol.C_CODE_OK,
		Heartbeat: hUint16,
	}
	data, err := json.Marshal(res)
	if nil != err {
		log.Logger.Error(
			"[Agent] 握手成功，但服务器未返回消息：编码握手消息出错",
		)

		return
	}

	pkt := NewPacket(protocol.C_MID_SCO, protocol.C_SID_HANDSHAKE_RES)
	pkt.AppendBytes(data)
	this.socket.SendPacket(pkt) // 越过工作状态发送消息

	// 状态： 等待握手 ack
	this.state.Set(C_AGENT_ST_WAIT_ACK)
}

//  收到握手 ACK
func (this *Agent) onAck() {
	// 状态：工作中
	if !this.state.CompareAndSwap(C_AGENT_ST_WAIT_ACK, C_AGENT_ST_WORKING) {
		this.Stop()
		return
	}

	// 发送心跳数据
	this.sendHeartbeat()
}

//  发送心跳数据
func (this *Agent) sendHeartbeat() error {
	// 发送心跳数据
	pkt := NewPacket(protocol.C_MID_SCO, protocol.C_SID_HEARTBEAT)
	err := this.SendPacket(pkt)

	return err
}

// 处理 pkcket
func (this *Agent) handle(pkt *Packet) {
	if this.state.Get() != state.C_WORKING {
		log.Logger.Debug("[Agent] 非工作状态，关闭改连接")

		this.Stop()
		return
	}

	if this.packetChan != nil {
		this.packetChan <- pkt
	}
}

// 停止成功
func (this *Agent) onStop() {
	// 将信息打包
	// 发送到主线程一个 chan 中
	// 主线程将消息发送给服务方
}
