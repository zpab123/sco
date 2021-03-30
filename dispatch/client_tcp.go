// /////////////////////////////////////////////////////////////////////////////
// 消息分发客户端

package dispatch

import (
	"sync"

	"github.com/zpab123/sco/discovery"
	"github.com/zpab123/sco/log"
	"github.com/zpab123/sco/network"
)

// tcp Client
type TcpClient struct {
	connMap sync.Map // conn 连接集合
}

// 新建1个 TcpServer
// 成功： 返回 *TcpServer, nil
// 失败： 返回 nil, error
func NewTcpClient() (*TcpClient, error) {
	//var err error

	c := TcpClient{}

	return &c, nil
}

// 启动
func (this *TcpClient) Run() {

}

// 停止
func (this *TcpClient) Stop() {

}

// 收到消息 [network.IClientHandler]
func (this *TcpClient) OnPacket(c *network.TcpConn, pkt *network.Packet) {
	// 需要发送给客户端
}

// 添加集群服务信息
func (this *TcpClient) AddService(desc *discovery.ServiceDesc) {
	addr := desc.Address()

	log.Logger.Debug(
		"发现新服务",
		log.String("addr=", addr),
	)

	c := network.NewTcpConn(addr)
	c.SetHandler(this)
	c.Run()

	this.connMap.Store(desc.Name, c)
}

// 移除集群服务信息
func (this *TcpClient) RemoveService(desc *discovery.ServiceDesc) {
	if c, ok := this.connMap.Load(desc.Name); ok {
		this.connMap.Delete(desc.Name)
		// 销毁连接对象？
		tc, rok := c.(*network.TcpConn)
		if rok {
			tc.Stop()
		}

		log.Logger.Debug(
			"移除服务",
			log.String("addr=", desc.Address()),
		)
	}
}
