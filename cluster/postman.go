// /////////////////////////////////////////////////////////////////////////////
// 消息分发

package cluster

import (
	"encoding/json"

	"github.com/zpab123/sco/log"
	"github.com/zpab123/sco/network"
	"github.com/zpab123/sco/protocol"
)

// -----------------------------------------------------------------------------
// Postman

// 消息转发
type Postman struct {
	svcId   uint16               // 服务id
	addrs   []string             // 网关地址
	gates   []*network.TcpConn   // 网关集合
	scoPkt  chan *network.Packet // 引擎消息
	gatePkt chan *network.Packet // 网关消息
}

// 创建1个 Postman

// svc=服务id gs=网关服务器地址列表
func NewPostman(svc uint16, as []string) *Postman {
	p := Postman{
		svcId:  svc,
		addrs:  as,
		scoPkt: make(chan *network.Packet, 100),
		gates:  make([]*network.TcpConn, 0),
	}

	return &p
}

// -----------------------------------------------------------------------------
// public

// 启动
func (this *Postman) Run() {
	for _, addr := range this.addrs {
		conn := network.NewTcpConn(addr)
		conn.SetScoPktChan(this.scoPkt)
		conn.SetNetPktChan(this.gatePkt)

		err := conn.Run()
		if err != nil {
			continue
		}

		this.gates = append(this.gates, conn)
	}

	go this.listen()
}

// 停止
func (this *Postman) Stop() {

}

// 推送消息
func (this *Postman) Post(pkt *network.Packet) {
	if len(this.gates) <= 0 {
		return
	}

	// 选择一个网关
	conn := this.gates[0]

	// 发送出去
	if conn != nil {
		conn.Send(pkt)
	}
}

// 设置网关消息通道
func (this *Postman) SetPktChan(ch chan *network.Packet) {
	if ch != nil {
		this.gatePkt = ch
	}
}

// -----------------------------------------------------------------------------
// private

// 处理消息
func (this *Postman) listen() {
	for {
		select {
		case pkt := <-this.scoPkt:
			this.onScoPkt(pkt)
		}
	}
}

// 引擎消息
func (this *Postman) onScoPkt(pkt *network.Packet) {
	if pkt.GetMid() != protocol.C_MID_SCO {
		return
	}

	switch pkt.GetSid() {
	case protocol.C_SID_CONN_WORKING:
		this.onConnWork(pkt.GetConn())
	}
}

// 开始工作
func (this *Postman) onConnWork(conn network.IConn) {
	// 创建协议
	req := protocol.ServiceRegReq{
		Id: this.svcId,
	}

	data, err := json.Marshal(&req)
	if nil != err {
		log.Logger.Debug(
			"[Postman] 编码服务注册消息失败",
		)

		return
	}

	// 发送请求
	pkt := network.NewPacket(protocol.C_MID_GATE, protocol.C_SID_SVCREG_REQ)
	pkt.AppendBytes(data)
	conn.Send(pkt)
}
