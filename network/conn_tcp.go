// /////////////////////////////////////////////////////////////////////////////
// tcp 客户端

package network

import (
	"encoding/json"
	"net"
	"time"

	"github.com/zpab123/sco/log"
	"github.com/zpab123/sco/protocol"
	"github.com/zpab123/sco/state"
	"github.com/zpab123/sco/syncs"
)

// /////////////////////////////////////////////////////////////////////////////
// TcpConn

// tcp 客户端
type TcpConn struct {
	addr       string            // 远端地址
	socket     *Socket           // socket
	state      *state.State      // 状态管理
	heartbeat  time.Duration     // 心跳周期
	heartSend  int64             // 心跳-发送(纳秒)
	heartRecv  int64             // 心跳-接受(纳秒)
	lastRecv   syncs.AtomicInt64 // 上次收到数据的时间
	lastSend   syncs.AtomicInt64 // 上次发送数据时间
	chDie      chan struct{}     // 关闭通道
	packetChan chan *Packet      // 消息通道
}

// 新建1个 tcp 连接
func NewTcpConn(addr string) *TcpConn {
	// 状态管理
	st := state.NewState()

	c := TcpConn{
		addr:  addr,
		state: st,
		chDie: make(chan struct{}),
	}
	c.lastRecv.Store(time.Now().UnixNano())
	c.lastSend.Store(time.Now().UnixNano())

	// 设置为初始化状态
	c.state.Set(C_CLI_ST_INIT)

	return &c
}

// 启动
func (this *TcpConn) Run() error {
	conn, err := net.Dial("tcp", this.addr)
	if nil != err {
		return err
	}

	s, err := NewSocket(conn)
	if nil != err {
		return err
	}

	this.socket = s

	// 接收线程
	go this.recvLoop()

	// 发送线程
	go this.sendLoop()

	return nil
}

// 停止
func (this *TcpConn) Stop() {
	if this.state.Get() == C_AGENT_ST_CLOSING {
		return
	}

	if this.state.Get() == C_AGENT_ST_CLOSED {
		return
	}

	this.state.Set(C_AGENT_ST_CLOSING)

	err := this.socket.Close()
	if err != nil {
		return
	}

	close(this.chDie)

	this.state.Set(C_AGENT_ST_CLOSED)
}

// 发送1个 packet 消息
func (this *TcpConn) SendPacket(pkt *Packet) error {
	// 状态效验
	if this.state.Get() != C_CLI_ST_WORKING {
		return errState
	}

	this.socket.SendPacket(pkt)

	return nil
}

// 发送 []byte
func (this *TcpConn) SendBytes(bytes []byte) error {
	// 状态效验
	if this.state.Get() != C_CLI_ST_WORKING {
		return errState
	}

	this.socket.SendBytes(bytes)

	return nil
}

// 打印信息
func (this *TcpConn) String() string {
	return this.socket.String()
}

// 设置消息通道
func (this *TcpConn) SetPacketChan(ch chan *Packet) {
	if ch != nil {
		this.packetChan = ch
	}
}

// -----------------------------------------------------------------------------
// private

// 接收线程
func (this *TcpConn) recvLoop() {
	defer func() {
		// 用于结束 sendLoop
		this.socket.SendPacket(nil)
	}()

	for {
		pkt, err := this.socket.RecvPacket()
		if nil != err {
			log.Logger.Debug(
				"[TcpConn] recvLoop 结束",
				log.String("err", err.Error()),
			)

			return
		}

		if nil != pkt {
			this.onPacket(pkt)
			continue
		}
	}
}

// 发送线程
func (this *TcpConn) sendLoop() {
	defer func() {
		this.Stop()
	}()

	// 请求握手
	this.reqHandShake()

	for {
		err := this.socket.Flush()
		if nil != err {
			log.Logger.Debug(
				"[TcpConn] sendLoop 结束",
				log.String("err", err.Error()),
			)

			break
		}

		// 记录发送时间
		this.lastSend.Store(time.Now().UnixNano())
	}
}

// 发送握手请求
func (this *TcpConn) reqHandShake() {
	req := protocol.HandshakeReq{
		Key:      C_F_KEY,
		Acceptor: 1,
	}

	data, err := json.Marshal(&req)
	if nil != err {
		log.Logger.Debug(
			"[TcpConn] 编码握手消息失败",
		)

		this.Stop()
		return
	}

	pkt := NewPacket(protocol.C_MID_SCO, protocol.C_SID_HANDSHAKE_REQ)
	pkt.AppendBytes(data)

	this.socket.SendBytes(pkt.Data())
}

// 发送握手ack
func (this *TcpConn) sendAck() {
	pkt := NewPacket(protocol.C_MID_SCO, protocol.C_SID_ACK)
	this.socket.SendBytes(pkt.Data())

	this.state.Set(C_CLI_ST_WORKING)

	// 通知可以发送数据了
	if this.packetChan != nil {
		pkt := NewPacket(protocol.C_MID_SCO, protocol.C_SID_AGENT_WORKING)
		// 保存网络
		this.packetChan <- pkt
	}
}

// 收到1个 pakcet
func (this *TcpConn) onPacket(pkt *Packet) {
	this.lastRecv.Store(time.Now().UnixNano())
	switch pkt.mid {
	case protocol.C_MID_INVALID: // 无效
		log.Logger.Debug("[TcpConn] 无效 packet",
			log.Uint16("mid", pkt.mid),
		)

		this.Stop()
	case protocol.C_MID_SCO: // sco 内部消息
		this.onScoPacket(pkt)
	default:
		this.handle(pkt) // 处理
	}
}

// 框架内部消息
func (this *TcpConn) onScoPacket(pkt *Packet) {
	switch pkt.sid {
	case protocol.C_SID_HANDSHAKE_RES: // 握手请求
		this.onHandshake(pkt.GetBody())
	case protocol.C_SID_HEARTBEAT: // 心跳
	//log.Sugar.Debug("心跳")
	default:
		log.Logger.Debug("[TcpConn] 无效 packet",
			log.Uint16("sid", pkt.sid),
		)

		this.Stop()
	}
}

// 握手结果
func (this *TcpConn) onHandshake(data []byte) {
	res := protocol.HandshakeRes{}
	err := json.Unmarshal(data, &res)
	if nil != err {
		log.Logger.Debug(
			"[TcpConn] 解码握手结果失败",
		)

		this.Stop()
		return
	}

	if res.Code == protocol.C_CODE_OK {
		t := time.Duration(res.Heartbeat)
		this.heartbeat = t * time.Second
		this.heartSend = int64(this.heartbeat) / 2
		this.heartRecv = int64(this.heartbeat)

		this.sendAck()

		if this.heartbeat > 0 {
			go this.checkHeart()
		}
	} else {
		log.Logger.Debug(
			"[TcpConn] 握手失败",
			log.Uint32("code", res.Code),
		)

		this.Stop()
		return
	}
}

// 发送心跳数据
func (this *TcpConn) sendHeartbeat() error {
	// 发送心跳数据
	pkt := NewPacket(protocol.C_MID_SCO, protocol.C_SID_HEARTBEAT)
	err := this.SendPacket(pkt)

	return err
}

// 需要处理的消息
func (this *TcpConn) handle(pkt *Packet) {
	if this.state.Get() != C_CLI_ST_WORKING {
		this.Stop()

		return
	}

	if this.packetChan != nil {
		this.packetChan <- pkt
	}
}

// 检查心跳
func (this *TcpConn) checkHeart() {
	// 半程检测
	hb := this.heartbeat / 2
	ticker := time.NewTicker(hb)

	defer func() {
		ticker.Stop()

		log.Logger.Debug(
			"[TcpConn] checkHeart 结束",
		)
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

// 检查发送是否超时
func (this *TcpConn) checkSendTime(t time.Time) {
	pass := t.UnixNano() - this.lastSend.Load()
	if pass >= this.heartSend {
		this.sendHeartbeat()
	}
}

// 检查接收是否超时
func (this *TcpConn) checkRecvTime(t time.Time) {
	pass := t.UnixNano() - this.lastRecv.Load()
	if pass >= this.heartRecv {
		log.Logger.Debug(
			"[TcpConn] 远端心跳超时，关闭连接",
		)

		this.Stop()
	}
}
