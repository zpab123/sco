// /////////////////////////////////////////////////////////////////////////////
// 代理对应于用户，用于存储原始连接信息

package network

import (
	"encoding/json"
	"time"

	"github.com/pkg/errors"
	"github.com/zpab123/sco/protocol"
	"github.com/zpab123/sco/state"
	"github.com/zpab123/syncutil"
	"github.com/zpab123/zaplog"
)

// /////////////////////////////////////////////////////////////////////////////
// 初始化

// 常量
const (
	heartbeatUint16 uint16 = uint16(C_AGENT_HEARTBEAT / time.Second) // 心跳时间(秒)（uint16）
	heartbeatInt64  int64  = int64(C_AGENT_HEARTBEAT / time.Second)  // 心跳时间(秒)（uint16）
)

// 变量
var (
	errState error = errors.New("状态错误")
)

// /////////////////////////////////////////////////////////////////////////////
// Agent

// 代理对应于用户，用于存储原始连接信息
type Agent struct {
	options  *AgentOpt            // 配置参数
	id       int32                // id 标识
	socket   *Socket              // socket
	state    *state.State         // 状态管理
	handler  IHandler             // 消息处理
	connMgr  IConnManager         // 连接管理
	lastTime syncutil.AtomicInt64 // 上次收到客户端消息的时间
	chDie    chan struct{}        // 关闭通道
	//checkHandshake interface // 握手检查函数（返回值为：握手数据）
}

// 新建1个 *Agent 对象
// 成功：返回 *Agent nil
// 失败：返回 nil error
func NewAgent(socket *Socket, opt *AgentOpt) (*Agent, error) {
	// 参数效验
	if nil == socket {
		err := errors.New("参数 socket 为空")
		return nil, err
	}

	if nil == opt {
		opt = NewAgentOpt()
	}

	// 状态管理
	st := state.NewState()

	// 创建对象
	a := Agent{
		options: opt,
		socket:  socket,
		state:   st,
		chDie:   make(chan struct{}),
	}
	a.lastTime.Store(time.Now().Unix())

	// 设置为初始化状态
	a.state.Set(C_AGENT_ST_INIT)

	return &a, nil
}

// 启动
func (this *Agent) Run() {
	// 发送线程
	go this.sendLoop()
	// 心跳
	go this.heartbeatLoop()
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
	if nil != this.connMgr {
		this.connMgr.OnAgentStop(this)
	}

	this.state.Set(C_AGENT_ST_CLOSED)
}

// 设置连接管理
func (this *Agent) SetConnMgr(mgr IConnManager) {
	if nil != mgr {
		this.connMgr = mgr
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
		zaplog.Debugf("[Agent] recvLoop 结束")
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
		zaplog.Debugf("[Agent] sendLoop 结束")
		this.Stop()
	}()

	for {
		err := this.socket.Flush()
		if nil != err {
			break
		}
	}
}

// 心跳线程
func (this *Agent) heartbeatLoop() {
	if this.options.Heartbeat <= 0 {
		return
	}

	ticker := time.NewTicker(this.options.Heartbeat)

	defer func() {
		zaplog.Debugf("[Agent] heartbeatLoop 结束")
		ticker.Stop()
		this.Stop()
	}()

	for {
		select {
		case <-ticker.C:
			pass := time.Now().Unix() - this.lastTime.Load()
			if pass > heartbeatInt64*2 {
				zaplog.Debugf("[Agent] %s 心跳超时，关闭连接", this)
				return
			} else if pass >= heartbeatInt64 {
				this.sendHeartbeat()
			}
		case <-this.chDie:
			return
		}
	}
}

// 收到1个 pakcet
func (this *Agent) onPacket(pkt *Packet) {
	this.lastTime.Store(time.Now().Unix())
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
		return
	}

	// 握手协议检查

	// 握手成功
	this.handshakeOk()
}

//  握手成功
func (this *Agent) handshakeOk() {
	// 返回数据
	res := &protocol.HandshakeRes{
		Code:      protocol.C_CODE_OK,
		Heartbeat: heartbeatUint16,
	}
	data, err := json.Marshal(res)
	if nil != err {
		zaplog.Error("[Agent] 握手成功，但服务器未返回消息：编码握手消息出错")
		return
	}

	pkt := NewPacket(protocol.C_MID_HANDSHAKE)
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
		zaplog.Error("握手失败，但服务器未返回消息：编码握手消息出错")

		return
	}

	pkt := NewPacket(protocol.C_MID_HANDSHAKE)
	pkt.AppendBytes(data)
	this.socket.SendPacket(pkt) // 越过工作状态发送消息
}

//  握手ACK
func (this *Agent) onAck() {
	// 状态：工作中
	if !this.state.CompareAndSwap(C_AGENT_ST_WAIT_ACK, C_AGENT_ST_WORKING) {
		return
	}

	// 发送心跳数据
	this.sendHeartbeat()
}

//  发送心跳数据
func (this *Agent) sendHeartbeat() error {
	// 发送心跳数据
	pkt := NewPacket(protocol.C_MID_HEARTBEAT)
	err := this.SendPacket(pkt)

	return err
}

// 处理 pkcket
func (this *Agent) handle(pkt *Packet) {
	if pkt.GetMid() <= protocol.C_MID_SCO {
		this.Stop()
		return
	}

	if nil != this.handler {
		this.handler.OnPacket(this, pkt)
	}
}

// /////////////////////////////////////////////////////////////////////////////
// AgentOpt 对象

// Agent 配置参数
type AgentOpt struct {
	Heartbeat time.Duration // 心跳周期
}

// 创建1个默认的 TAgentOpt
func NewAgentOpt() *AgentOpt {
	o := AgentOpt{
		Heartbeat: C_AGENT_HEARTBEAT,
	}

	return &o
}
