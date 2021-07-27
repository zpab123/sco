// /////////////////////////////////////////////////////////////////////////////
// 消息分发

package dispatch

import (
	"encoding/json"

	"github.com/zpab123/sco/log"
	"github.com/zpab123/sco/network"
	"github.com/zpab123/sco/protocol"
)

// -----------------------------------------------------------------------------
// public

type Dispatcher struct {
	svcId      uint16               // 服务id
	addrs      []string             // 连接地址
	conns      []*network.TcpConn   // 连接集合
	scoPacket  chan *network.Packet // 引擎消息通道
	packetChan chan *network.Packet // 数据包
}

func NewDispatcher(id uint16, a []string) *Dispatcher {
	d := Dispatcher{
		svcId:     id,
		addrs:     a,
		scoPacket: make(chan *network.Packet, 100),
		conns:     make([]*network.TcpConn, 0),
	}

	return &d
}

// 启动
func (this *Dispatcher) Run() {
	// 连接陈宫后，想分发器注册服务
	// 分发器回复注册结果
	// 注册成功后

	for _, addr := range this.addrs {
		conn := network.NewTcpConn(addr)
		conn.SetScoPacket(this.scoPacket)
		conn.SetPacketChan(this.packetChan)

		err := conn.Run()
		if err != nil {
			continue
		}

		this.conns = append(this.conns, conn)
	}

	go this.start()
}

func (this *Dispatcher) Send(pkt *network.Packet) {
	if len(this.conns) <= 0 {
		return
	}

	// 选择一个分发器
	conn := this.conns[0]

	// 发送出去
	if conn != nil {
		conn.SendPacket(pkt)
	}
}

func (this *Dispatcher) SetPacketChan(ch chan *network.Packet) {
	if ch != nil {
		this.packetChan = ch
	}
}

// -----------------------------------------------------------------------------
// private

// 处理内部消息
func (this *Dispatcher) start() {
	select {
	case pkt := <-this.scoPacket:
		this.onScoPkt(pkt)
	}
}

// 引擎消息
func (this *Dispatcher) onScoPkt(pkt *network.Packet) {
	if pkt.GetMid() != protocol.C_MID_SCO {
		return
	}

	switch pkt.GetSid() {
	case protocol.C_SID_AGENT_WORKING:
		this.onConnWorking(pkt.GetConn())
	}
}

// 开始工作
func (this *Dispatcher) onConnWorking(conn network.IConn) {
	// 创建协议
	req := protocol.ServiceRegReq{
		Id: this.svcId,
	}

	data, err := json.Marshal(&req)
	if nil != err {
		log.Logger.Debug(
			"[Dispatcher] 编码服务注册消息失败",
		)

		return
	}

	// 发送请求
	pkt := network.NewPacket(protocol.C_MID_DISPATCH, protocol.C_SID_SVCREG_REQ)
	pkt.AppendBytes(data)
	conn.SendPacket(pkt)
}
