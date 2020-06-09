// /////////////////////////////////////////////////////////////////////////////
// 代理对应于用户，用于存储原始连接信息

package network

import (
	"encoding/json"
	"errors"

	"github.com/zpab123/sco/protocol"
	"github.com/zpab123/sco/state"
	"github.com/zpab123/zaplog"
)

// /////////////////////////////////////////////////////////////////////////////
// 初始化

var (
	errState error = errors.New("状态错误")
)

// /////////////////////////////////////////////////////////////////////////////
// Agent

// 代理对应于用户，用于存储原始连接信息
type Agent struct {
	options  *TAgentOpt          // 配置参数
	socket   *Socket             // socket
	stateMgr *state.StateManager // 状态管理
	handler  IHandler            // 消息处理
	//checkHandshake interface // 握手检查函数（返回值为：握手数据）
}

// 新建1个 *Agent 对象
func NewAgent(socket *Socket, opt *TAgentOpt) *Agent {
	// 参数效验
	if nil == opt {
		opt = NewTAgentOpt()
	}

	// 状态管理
	st := state.NewStateManager()

	// 创建对象
	a := Agent{
		options:  opt,
		socket:   socket,
		stateMgr: st,
	}

	// 设置为初始化状态
	a.stateMgr.SetState(C_AGENT_ST_INIT)

	return &a
}

// 启动
func (this *Agent) Run() {
	// 发送线程
	go this.sendLoop()
	// 心跳？
	// 定时器？
	this.recvLoop() // 接收循环，这里不能 go this.recvLoop()，会导致 websocket 连接直接断开
}

// 停止
func (this *Agent) Stop() {
	this.socket.Close()
}

// 发送1个 packet 消息
func (this *Agent) SendPacket(pkt *Packet) error {
	// 状态效验
	if this.stateMgr.GetState() != C_AGENT_ST_WORKING {
		return errState
	}

	this.socket.SendPacket(pkt)

	return nil
}

// 发送 []byte
func (this *Agent) SendBytes(bytes []byte) error {
	// 状态效验
	if this.stateMgr.GetState() != C_AGENT_ST_WORKING {
		return errState
	}

	return this.socket.SendBytes(bytes)
}

// 打印信息
func (this *Agent) String() string {
	return this.socket.String()
}

// 设置处理器
func (this *Agent) SetHandler(h IHandler) {
	if nil != h {
		this.handler = h
	}
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

// 收到1个 pakcet
func (this *Agent) onPacket(pkt *Packet) {
	switch pkt.mid {
	case protocol.C_MID_HANDSHAKE: // 客户端握手请求
		this.onHandshake(pkt.GetBody())
	case protocol.C_MID_HANDSHAKE_ACK: // 客户端握手 ACK
		this.onAck()
	default:
		this.handler.OnPacket(this, pkt)
	}
}

// 握手消息
func (this *Agent) onHandshake(body []byte) {
	// 状态效验
	if this.stateMgr.GetState() != C_AGENT_ST_INIT {
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
		Heartbeat: 0,
	}
	data, err := json.Marshal(res)
	if nil != err {
		zaplog.Error("握手成功，但服务器未返回消息：编码握手消息出错")

		return
	}

	pkt := NewPacket(protocol.C_MID_HANDSHAKE)
	pkt.AppendBytes(data)
	this.socket.SendPacket(pkt) // 越过工作状态发送消息

	// 状态： 等待握手 ack
	this.stateMgr.SetState(C_AGENT_ST_WAIT_ACK)
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
	if !this.stateMgr.CompareAndSwap(C_AGENT_ST_WAIT_ACK, C_AGENT_ST_WORKING) {
		return
	}

	// 发送心跳数据
	this.sendHeartbeat()
}

//  发送心跳数据
func (this *Agent) sendHeartbeat() error {
	if this.stateMgr.GetState() != C_AGENT_ST_WORKING {
		return errState
	}

	// 发送心跳数据
	pkt := NewPacket(protocol.C_PKT_ID_HEARTBEAT)
	err := this.SendPacket(pkt)

	return err
}
