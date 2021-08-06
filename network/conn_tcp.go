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
	addr      string            // 远端地址
	socket    *Socket           // socket
	session   *Session          // 会话
	state     *state.State      // 状态管理
	heartbeat time.Duration     // 心跳周期
	heartSend int64             // 心跳-发送(纳秒)
	heartRecv int64             // 心跳-接受(纳秒)
	lastRecv  syncs.AtomicInt64 // 上次收到数据的时间
	lastSend  syncs.AtomicInt64 // 上次发送数据时间
	chDie     chan struct{}     // 关闭通道
	evtChan   chan *ConnEvent   // 连接事件
	clientPkt chan *Packet      // client -> server 消息
	serverPkt chan *Packet      // server -> server 消息
	stcPkt    chan *Packet      // server -> client
	postMan   *Postman          // 转发对象
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

// /////////////////////////////////////////////////////////////////////////////
// 打印接口

// 打印信息
func (this *TcpConn) String() string {
	return this.socket.String()
}

// /////////////////////////////////////////////////////////////////////////////
// public

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

	ses := NewSession()
	ses.conn = this
	ses.postMan = this.postMan

	this.session = ses

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

	this.socket.Close()

	close(this.chDie)

	this.state.Set(C_AGENT_ST_CLOSED)
}

// 发送1个 packet 消息
func (this *TcpConn) Send(pkt *Packet) error {
	if this.state.Get() != C_CLI_ST_WORKING {
		return errState
	}

	this.socket.Send(pkt)

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

// 设置转发
func (this *TcpConn) SetPostman(man *Postman) {
	if man != nil {
		this.postMan = man
	}
}

// 设置连接事件通道
func (this *TcpConn) SetEventChan(ch chan *ConnEvent) {
	if ch != nil {
		this.evtChan = ch
	}
}

// 设置 客户端->服务器 消息通道
func (this *TcpConn) SetClientPacketChan(ch chan *Packet) {
	if ch != nil {
		this.clientPkt = ch
	}
}

// 设置 服务器->服务器 消息通道
func (this *TcpConn) SetServerPacketChan(ch chan *Packet) {
	if ch != nil {
		this.serverPkt = ch
	}
}

// 设置 服务器 -> 客户端 消息通道
func (this *TcpConn) SetStcPacketChan(ch chan *Packet) {
	if ch != nil {
		this.stcPkt = ch
	}
}

// -----------------------------------------------------------------------------
// private

// 接收线程
func (this *TcpConn) recvLoop() {
	defer func() {
		// 用于结束 sendLoop
		this.socket.Send(nil)
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
			pkt.session = this.session
			this.onPacket(pkt)
			continue
		}
	}
}

// 发送线程
func (this *TcpConn) sendLoop() {
	defer this.Stop()

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

	pkt := NewPacket(C_PKT_KIND_CONN, 0, 0, 0, protocol.C_MID_HANDSHAKE_REQ)
	pkt.AppendBytes(data)

	this.socket.SendBytes(pkt.Data())
}

// 发送握手ack
func (this *TcpConn) sendAck() {
	pkt := NewPacket(C_PKT_KIND_CONN, 0, 0, 0, protocol.C_MID_ACK)
	this.socket.SendBytes(pkt.Data())

	this.state.Set(C_CLI_ST_WORKING)

	// 通知连接状态
	if this.evtChan != nil {
		evt := ConnEvent{
			id:   C_EVT_WORKING,
			conn: this,
		}

		this.evtChan <- &evt
	}
}

// 收到1个 pakcet
func (this *TcpConn) onPacket(pkt *Packet) {
	this.lastRecv.Store(time.Now().UnixNano())

	switch pkt.kind {
	case C_PKT_KIND_CONN: // 连接消息
		this.onConnPkt(pkt)
	case C_PKT_KIND_CLI_SER: // client -> server
		this.onClientPkt(pkt)
	case C_PKT_KIND_SER_SER: // server -> server
		this.onServerPkt(pkt)
	case C_PKT_KIND_SER_CLI: // server -> client
		this.onStcPkt(pkt)
	default:
		log.Logger.Debug("[Agent] 无效 kind 断开连接",
			log.Int8("kind", int8(pkt.kind)),
		)

		this.Stop()
	}
}

// 连接消息
func (this *TcpConn) onConnPkt(pkt *Packet) {
	switch pkt.mid {
	case protocol.C_MID_HANDSHAKE_RES: // 握手结果
		this.onHandshake(pkt.GetBody())
	case protocol.C_MID_HEARTBEAT: // 心跳
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
func (this *TcpConn) sendHeartbeat() {
	pkt := NewPacket(C_PKT_KIND_CONN, 0, 0, 0, protocol.C_MID_HEARTBEAT)
	this.socket.Send(pkt)
}

// 需要处理的消息
func (this *TcpConn) handle(pkt *Packet) {
	if this.state.Get() != C_CLI_ST_WORKING {
		this.Stop()

		return
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

// 收到客户端消息
func (this *TcpConn) onClientPkt(pkt *Packet) {
	if this.clientPkt != nil {
		this.clientPkt <- pkt
	}
}

// 收到服务器消息
func (this *TcpConn) onServerPkt(pkt *Packet) {
	if this.serverPkt != nil {
		this.serverPkt <- pkt
	}
}

// server -> client
func (this *TcpConn) onStcPkt(pkt *Packet) {
	if this.stcPkt != nil {
		this.stcPkt <- pkt
	}
}
