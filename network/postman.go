// /////////////////////////////////////////////////////////////////////////////
// 消息分发

package network

import (
	"encoding/json"

	"github.com/zpab123/sco/log"
	"github.com/zpab123/sco/protocol"
)

// -----------------------------------------------------------------------------
// Postman

// 消息转发
type Postman struct {
	appid     uint16          // app 唯一编号
	svcId     uint16          // 服务id
	addrs     []string        // 集群地址
	clusters  []*TcpConn      // 集群集合
	connEvt   chan *ConnEvent // 连接事件
	clientPkt chan *Packet    // client -> server 消息
	serverPkt chan *Packet    // server -> server 消息
	stcPkt    chan *Packet    // server -> client 消息
}

// 创建1个 Postman

// svc=服务id gs=网关服务器地址列表
func NewPostman(aid, svc uint16, as []string) *Postman {
	p := Postman{
		appid:    aid,
		svcId:    svc,
		addrs:    as,
		connEvt:  make(chan *ConnEvent, 10),
		clusters: make([]*TcpConn, 0),
	}

	return &p
}

// -----------------------------------------------------------------------------
// public

// 启动
func (this *Postman) Run() {
	for _, addr := range this.addrs {
		conn := NewTcpConn(addr)
		conn.SetPostman(this)
		conn.SetEventChan(this.connEvt)
		conn.SetClientPacketChan(this.clientPkt)
		conn.SetServerPacketChan(this.serverPkt)
		conn.SetStcPacketChan(this.stcPkt)

		err := conn.Run()
		if err != nil {
			continue
		}

		this.clusters = append(this.clusters, conn)
	}

	go this.listen()
}

// 停止
func (this *Postman) Stop() {

}

// 将消息推送给某个服务类型
func (this *Postman) Post(pkt *Packet) {
	if len(this.clusters) <= 0 {
		return
	}

	// 选择一个中转服务器
	conn := this.clusters[0]

	// 发送出去
	if conn != nil {
		conn.Send(pkt)
	}
}

// 将消息推送给某个服务器
//
// id=服务器id
func (this *Postman) PostTo(id uint16, pkt *Packet) {
	// 此时 pkt 的 sid 就是 serverid
}

// 设置 客户端->服务器 消息通道
func (this *Postman) SetClientPacketChan(ch chan *Packet) {
	if ch != nil {
		this.clientPkt = ch
	}
}

// 设置 服务器->服务器 消息通道
func (this *Postman) SetServerPacketChan(ch chan *Packet) {
	if ch != nil {
		this.serverPkt = ch
	}
}

// 设置 服务器 -> 客户端 消息通道
func (this *Postman) SetStcPacketChan(ch chan *Packet) {
	if ch != nil {
		this.stcPkt = ch
	}
}

// -----------------------------------------------------------------------------
// private

// 处理消息
func (this *Postman) listen() {
	for {
		select {
		case evt := <-this.connEvt:
			this.onConnEvt(evt)
		}
	}
}

// 引擎消息
func (this *Postman) onConnEvt(evt *ConnEvent) {
	switch evt.id {
	case C_EVT_WORKING:
		this.onConnWork(evt.conn)
	}
}

// 开始工作
func (this *Postman) onConnWork(conn *TcpConn) {
	// 创建协议
	req := protocol.ServiceRegReq{
		AppId: this.appid,
		SvcId: this.svcId,
	}

	data, err := json.Marshal(&req)
	if nil != err {
		log.Logger.Debug(
			"[Postman] 编码服务注册消息失败",
		)

		return
	}

	// 发送请求
	pkt := NewPacket(C_PKT_KIND_SER_SER, 0, this.appid, protocol.C_SID_CLUSTER, protocol.C_MID_SVCREG_REQ)
	pkt.AppendBytes(data)
	conn.Send(pkt)
}