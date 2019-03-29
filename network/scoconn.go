// /////////////////////////////////////////////////////////////////////////////
// 对 PacketSocket 的封装

package network

import (
	"encoding/json"

	"github.com/pkg/errors"           // 异常
	"github.com/zpab123/sco/protocol" // world 内部通信协议
	"github.com/zpab123/sco/state"    // 状态管理
	"github.com/zpab123/zaplog"       // 日志
)

// /////////////////////////////////////////////////////////////////////////////
// ScoConn 对象

// sco 框架内部需要用到的一些常用网络消息
type ScoConn struct {
	stateMgr     *state.StateManager // 状态管理
	option       *TScoConnOpt        // 配置参数
	packetSocket *PacketSocket       // PacketSocket
}

// 新建1个 ScoConn 对象
func NewScoConn(socket ISocket, opt *TScoConnOpt) *ScoConn {
	// 参数效验
	if nil == opt {
		opt = NewTScoConnOpt()
	}

	// 创建状态管理
	st := state.NewStateManager()

	// 创建 packetSocket
	bufSocket := NewBufferSocket(socket, opt.BuffSocketOpt)
	pktSocket := NewPacketSocket(bufSocket)

	// 创建对象
	wc := &ScoConn{
		stateMgr:     st,
		packetSocket: pktSocket,
		option:       opt,
	}

	// 设置为初始化状态
	wc.stateMgr.SetState(C_CONN_STATE_INIT)

	return wc
}

// 接收1个 Packet 消息
func (this *ScoConn) RecvPacket() (*Packet, error) {
	// 接收 packet
	pkt, err := this.packetSocket.RecvPacket()
	if nil == pkt || nil != err {
		return nil, err
	}

	// 内部 packet
	if pkt.pktId < protocol.C_PKT_ID_SCO {
		this.handlePacket(pkt)

		return nil, err
	}

	// 状态效验
	if this.stateMgr.GetState() != C_CONN_STATE_WORKING {
		this.Close()

		err = errors.Errorf("ScoConn %s 收到数据，但是状态错误。当前状态=%d，正确状态=%s", this, this.stateMgr.GetState(), C_CONN_STATE_WORKING)

		return nil, err
	}

	return pkt, nil
}

// 关闭 ScoConn
func (this *ScoConn) Close() error {
	var err error

	// 关闭中
	if this.stateMgr.GetState() == C_CONN_STATE_CLOSING {
		err = errors.New("ScoConn 关闭失败：它正在关闭")

		return err
	} else {
		this.stateMgr.SetState(C_CONN_STATE_CLOSING)
	}

	err = this.packetSocket.Close()

	this.stateMgr.SetState(C_CONN_STATE_CLOSED)

	return err
}

//  发送心跳数据
func (this *ScoConn) SendHeartbeat() error {
	var err error
	if this.stateMgr.GetState() != C_CONN_STATE_WORKING {
		err = errors.New("发送心跳失败，ScoConn 状态不在工作中")

		return err
	}

	zaplog.Debugf("ScoConn %s 发送心跳", this)

	// 发送心跳数据
	pkt := NewPacket(protocol.C_PKT_ID_HEARTBEAT)
	err = this.sendPacket(pkt)

	return err
}

// 发送通用数据
func (this *ScoConn) SendData(data []byte) {
	pkt := NewPacket(protocol.C_PKT_ID_DATA)
	pkt.AppendBytes(data)

	this.sendPacket(pkt)
}

// 刷新缓冲区
func (this *ScoConn) Flush() error {
	return this.packetSocket.Flush()
}

// 打印信息
func (this *ScoConn) String() string {
	return this.packetSocket.String()
}

// 发送1个 packet 消息
func (this *ScoConn) sendPacket(pkt *Packet) error {
	var err error

	// 状态效验
	if this.stateMgr.GetState() != C_CONN_STATE_WORKING {
		err = errors.Errorf("ScoConn %s 发送 Packet 数据失败：状态不在 working 中", this)

		return err
	}

	return this.packetSocket.SendPacket(pkt)
}

// 处理 Packet 消息
func (this *ScoConn) handlePacket(pkt *Packet) {
	defer pkt.Release()

	// 根据类型处理数据
	switch pkt.pktId {
	case protocol.C_PKT_ID_HANDSHAKE: // 客户端握手请求
		this.handleHandshake(pkt.GetBody())
	case protocol.C_PKT_ID_HANDSHAKE_ACK: // 客户端握手 ACK
		this.handleHandshakeAck()
	default:
		zaplog.Errorf("ScoConn &s 收到无效消息类型=%d，关闭连接", this, pkt.pktId)

		this.Close()
	}
}

//  处理握手消息
func (this *ScoConn) handleHandshake(body []byte) {
	var err error
	// 状态效验
	if this.stateMgr.GetState() != C_CONN_STATE_INIT {
		return
	}

	// 解码消息
	req := &protocol.HandshakeReq{}
	err = json.Unmarshal(body, req)
	if nil != err {
		this.Close()

		return
	}

	// 版本验证
	if this.option.ShakeKey != "" && req.Key != this.option.ShakeKey {
		this.handshakeFail(protocol.C_CODE_SHAKE_KEY_ERROR)

		return
	}

	// 通信方式验证,后续添加

	// 握手成功
	this.handshakeOk()
}

//  返回握手消息
func (this *ScoConn) handshakeOk() {
	// 状态效验
	if this.stateMgr.GetState() != C_CONN_STATE_INIT {
		return
	}

	// 返回数据
	res := &protocol.HandshakeOk{
		Code:      protocol.C_CODE_OK,
		Heartbeat: this.option.Heartbeat,
	}
	data, err := json.Marshal(res)
	if nil != err {
		zaplog.Error("握手成功，但服务器未返回消息：编码握手消息出错")

		return
	}

	pkt := NewPacket(protocol.C_PKT_ID_HANDSHAKE)
	pkt.AppendBytes(data)
	this.packetSocket.SendPacket(pkt) // 越过工作状态发送消息

	// 状态： 等待握手 ack
	this.stateMgr.SetState(C_CONN_STATE_WAIT_ACK)
}

//  握手失败
func (this *ScoConn) handshakeFail(code uint32) {
	defer this.Close()

	// 状态效验
	if this.stateMgr.GetState() != C_CONN_STATE_INIT {
		return
	}

	// 返回数据
	res := &protocol.HandshakeFail{
		Code: code,
	}
	data, err := json.Marshal(res)
	if nil != err {
		zaplog.Error("握手失败，但服务器未返回消息：编码握手消息出错")

		return
	}

	pkt := NewPacket(protocol.C_PKT_ID_HANDSHAKE)
	pkt.AppendBytes(data)
	this.packetSocket.SendPacket(pkt) // 越过工作状态发送消息
}

//  处理握手ACK
func (this *ScoConn) handleHandshakeAck() {
	// 状态：工作中
	if !this.stateMgr.SwapState(C_CONN_STATE_WAIT_ACK, C_CONN_STATE_WORKING) {

		return
	}

	// 发送心跳数据
	this.SendHeartbeat()
}
