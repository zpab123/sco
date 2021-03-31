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
	"github.com/zpab123/sco/syncutil"
)

// /////////////////////////////////////////////////////////////////////////////
// TcpConn

// tcp 客户端
type TcpConn struct {
	addr      string               // 远端地址
	socket    *Socket              // socket
	state     *state.State         // 状态管理
	handler   IClientHandler       // 消息处理
	heartbeat time.Duration        // 心跳周期
	lastTime  syncutil.AtomicInt64 // 上次发送数据的时间
	chDie     chan struct{}        // 关闭通道
}

// 新建1个 tcp 连接
func NewTcpConn(addr string) *TcpConn {
	// 状态管理
	st := state.NewState()

	c := TcpConn{
		heartbeat: C_F_HEARTBEAT,
		addr:      addr,
		state:     st,
		chDie:     make(chan struct{}),
	}
	c.lastTime.Store(time.Now().Unix())

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

	// 发送线程
	go this.sendLoop()
	// 接收线程
	go this.recvLoop()
	// 心跳
	go this.heartbeatLoop()

	return nil
}

// 停止
func (this *TcpConn) Stop() {
	close(this.chDie)
	this.socket.Close()
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

// 设置处理器
func (this *TcpConn) SetHandler(h IClientHandler) {
	if nil != h {
		this.handler = h
	}
}

// 接收线程
func (this *TcpConn) recvLoop() {
	defer this.socket.SendPacket(nil) // 用于结束 sendLoop

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
		this.lastTime.Store(time.Now().Unix())
	}
}

// 心跳循环
func (this *TcpConn) heartbeatLoop() {
	if this.heartbeat <= 0 {
		return
	}

	ticker := time.NewTicker(this.heartbeat)

	defer func() {
		log.Logger.Debug(
			"[TcpConn] heartbeatLoop 结束",
		)

		ticker.Stop()
	}()

	hInt64 := int64(this.heartbeat / time.Second)

	for {
		select {
		case <-ticker.C:
			pass := time.Now().Unix() - this.lastTime.Load()
			if pass >= hInt64 {
				this.sendHeartbeat()
			}
		case <-this.chDie:
			return
		}
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

	pkt := NewPacket(protocol.C_MID_HANDSHAKE)
	pkt.AppendBytes(data)

	this.socket.SendBytes(pkt.Data())
}

// 发送握手ack
func (this *TcpConn) sendAck() {
	pkt := NewPacket(protocol.C_MID_HANDSHAKE_ACK)
	this.socket.SendBytes(pkt.Data())

	this.state.Set(C_CLI_ST_WORKING)
}

// 收到1个 pakcet
func (this *TcpConn) onPacket(pkt *Packet) {
	switch pkt.mid {
	case protocol.C_MID_HANDSHAKE: // 远端握手结果
		this.onHandshake(pkt.GetBody())
	case protocol.C_MID_HEARTBEAT: // 心跳
		//
	default:
		this.handle(pkt) // 处理
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
		this.sendAck()
	}
}

// 发送心跳数据
func (this *TcpConn) sendHeartbeat() error {
	// 发送心跳数据
	pkt := NewPacket(protocol.C_MID_HEARTBEAT)
	err := this.SendPacket(pkt)

	return err
}

// 需要处理的消息
func (this *TcpConn) handle(pkt *Packet) {
	if this.state.Get() != C_CLI_ST_WORKING {
		this.Stop()
		return
	}

	if nil != this.handler {
		this.handler.OnPacket(this, pkt)
	}
}
