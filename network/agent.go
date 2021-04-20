// /////////////////////////////////////////////////////////////////////////////
// 代理对应于用户，用于存储原始连接信息

package network

import (
	"encoding/json"
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
	id             int32             // id 标识
	socket         *Socket           // socket
	key            string            // 握手key
	heartbeat      time.Duration     // 心跳周期
	heartbeatInt64 int64             // 心跳周期(毫秒)
	handler        IHandler          // 消息处理
	state          *state.State      // 状态管理
	mgr            IAgentManager     // 连接管理
	lastRecv       syncs.AtomicInt64 // 上次收到数据的时间
	lastSend       syncs.AtomicInt64 // 上次发送数据时间
	chDie          chan struct{}     // 关闭通道
	packetChan     chan *Packet      // 消息通道
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
		key:            C_F_KEY,
		heartbeat:      C_F_HEARTBEAT,
		heartbeatInt64: int64(C_F_HEARTBEAT),
		socket:         socket,
		state:          st,
		chDie:          make(chan struct{}),
	}
	a.lastRecv.Store(time.Now().UnixNano())
	a.lastSend.Store(time.Now().UnixNano())

	// 设置为初始化状态
	a.state.Set(C_AGENT_ST_INIT)

	return &a, nil
}

// 启动
func (this *Agent) Run() {
	// 发送线程
	go this.sendLoop()

	// 心跳
	if this.heartbeat > 0 {
		go this.heartbeatLoop()
	}

	// 接收循环，这里不能 go this.recvLoop()，会导致 websocket 连接直接断开
	this.recvLoop()
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

	close(this.chDie)

	this.socket.Close()
	if nil != this.mgr {
		this.mgr.OnAgentStop(this)
	}

	this.state.Set(C_AGENT_ST_CLOSED)
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
func (this *Agent) SetConnMgr(mgr IAgentManager) {
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

// 设置 Handler 处理器
func (this *Agent) SetHandler(h IHandler) {
	if nil != h {
		this.handler = h
	}
}

// 打印信息
func (this *Agent) String() string {
	return this.socket.String()
}

// 接收线程
func (this *Agent) recvLoop() {
	defer func() {
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
			this.onPacket(pkt)
			continue
		}
	}
}

// 发送线程
func (this *Agent) sendLoop() {
	defer func() {
		log.Logger.Debug(
			"[Agent] sendLoop 结束",
		)

		this.Stop()
	}()

	for {
		err := this.socket.Flush()
		if nil != err {
			break
		}

		this.lastSend.Store(time.Now().UnixNano())
	}
}

// 心跳线程
func (this *Agent) heartbeatLoop() {
	// 半程检测
	hb := this.heartbeat / 2
	ticker := time.NewTicker(hb)

	defer func() {
		log.Logger.Debug(
			"[Agent] heartbeatLoop 结束",
		)

		ticker.Stop()
	}()

	for {
		select {
		case <-ticker.C:
			t := time.Now()
			this.checkRecvTime(t)
			this.checkSendTime(t)
		case <-this.chDie:
			return
		}
	}
}

// 收到1个 pakcet
func (this *Agent) onPacket(pkt *Packet) {
	this.lastRecv.Store(time.Now().UnixNano())
	switch pkt.mid {
	case protocol.C_MID_INVALID: // 无效
		this.Stop()
	case protocol.C_MID_HANDSHAKE: // 客户端握手请求
		this.onHandshake(pkt.GetBody())
	case protocol.C_MID_HANDSHAKE_ACK: // 客户端握手 ACK
		this.onAck()
	case protocol.C_MID_HEARTBEAT: // 心跳
	default:
		this.handle(pkt)
	}
}

// 握手消息
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

//  握手成功
func (this *Agent) handshakeOk() {
	defer log.Logger.Sync()

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

	pkt := NewPacket(protocol.C_MID_HANDSHAKE, 0)
	pkt.AppendBytes(data)
	this.socket.SendPacket(pkt) // 越过工作状态发送消息

	// 状态： 等待握手 ack
	this.state.Set(C_AGENT_ST_WAIT_ACK)
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

	pkt := NewPacket(protocol.C_MID_HANDSHAKE, 0)
	pkt.AppendBytes(data)
	this.socket.SendPacket(pkt) // 越过工作状态发送消息
}

//  握手ACK
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
	pkt := NewPacket(protocol.C_MID_HEARTBEAT, 0)
	err := this.SendPacket(pkt)

	return err
}

// 处理 pkcket
func (this *Agent) handle(pkt *Packet) {
	if pkt.mid <= protocol.C_MID_SCO {
		this.Stop()
		return
	}

	if this.packetChan != nil {
		this.packetChan <- pkt
	}
}

// 检查发送是否超时
func (this *Agent) checkSendTime(t time.Time) {
	pass := t.UnixNano() - this.lastSend.Load()
	if pass >= this.heartbeatInt64/2 {
		this.sendHeartbeat()
	}
}

// 检查接收是否超时
func (this *Agent) checkRecvTime(t time.Time) {
	pass := t.UnixNano() - this.lastRecv.Load()
	if pass >= this.heartbeatInt64 {
		log.Logger.Debug(
			"[Agent] 心跳超时，关闭连接",
		)

		this.Stop()
	}
}
