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
	appid     uint16               // app 唯一编号
	svcId     uint16               // 服务id
	addrs     []string             // 集群地址
	clusters  []*network.TcpConn   // 集群集合
	scoPkt    chan *network.Packet // 引擎消息
	clientPkt chan *network.Packet // client -> server 消息
	serverPkt chan *network.Packet // server -> server 消息
	stcPkt    chan *network.Packet // server -> client 消息
}

// 创建1个 Postman

// svc=服务id gs=网关服务器地址列表
func NewPostman(aid, svc uint16, as []string) *Postman {
	p := Postman{
		appid:    aid,
		svcId:    svc,
		addrs:    as,
		scoPkt:   make(chan *network.Packet, 100),
		clusters: make([]*network.TcpConn, 0),
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

// 推送消息
func (this *Postman) Post(pkt *network.Packet) {
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

// 将消息发送给指定服务器
//
// id=服务器id
func (this *Postman) PostTo(id uint16, pkt *network.Packet) {
	// 此时 pkt 的 sid 就是 serverid
}

// 设置客户端消息通道
func (this *Postman) SetClientPacketChan(ch chan *network.Packet) {
	if ch != nil {
		this.clientPkt = ch
	}
}

// 设置集群消息通道
func (this *Postman) SetServerPacketChan(ch chan *network.Packet) {
	if ch != nil {
		this.serverPkt = ch
	}
}

// 设置集群消息通道
func (this *Postman) SetStcPacketChan(ch chan *network.Packet) {
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
		case pkt := <-this.scoPkt:
			this.onScoPkt(pkt)
		}
	}
}

// 引擎消息
func (this *Postman) onScoPkt(pkt *network.Packet) {
	if pkt.Kind() != network.C_PKT_KIND_SCO {
		return
	}

	switch pkt.Sid() {
	case protocol.C_SID_NET:
		this.onConnWork(pkt.GetConn())
	}
}

// 网络消息
func (this *Postman) onNetPkt(pkt *network.Packet) {
	switch pkt.Mid() {
	case protocol.C_MID_NET_WORK:
		this.onConnWork(pkt.GetConn())
	}
}

// 开始工作
func (this *Postman) onConnWork(conn network.IConn) {
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
	pkt := network.NewPacket(network.C_PKT_KIND_SER_SER, 0, this.appid, protocol.C_SID_CLUSTER, protocol.C_MID_SVCREG_REQ)
	pkt.AppendBytes(data)
	conn.Send(pkt)
}
