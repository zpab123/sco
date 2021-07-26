// /////////////////////////////////////////////////////////////////////////////
// 消息分发

package dispatch

import (
	"github.com/zpab123/sco/network"
)

// -----------------------------------------------------------------------------
// public

type Dispatcher struct {
	addrs      []string
	conns      []*network.TcpConn
	packetChan chan *network.Packet // 数据包
}

func NewDispatcher(a []string) *Dispatcher {
	d := Dispatcher{
		addrs: a,
		conns: make([]*network.TcpConn, 0),
	}

	return &d
}

func (this *Dispatcher) Run() {
	// 连接陈宫后，想分发器注册服务
	// 分发器回复注册结果
	// 注册成功后

	for _, addr := range this.addrs {
		conn := network.NewTcpConn(addr)
		conn.SetPacketChan(this.packetChan)

		err := conn.Run()
		if err != nil {
			continue
		}

		this.conns = append(this.conns, conn)
	}
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
