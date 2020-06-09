// /////////////////////////////////////////////////////////////////////////////
// 消息分发客户端

package dispatch

import (
	"sync"

	//"github.com/pkg/errors"
	"github.com/zpab123/sco/discovery" // 服务发现
	"github.com/zpab123/sco/network"
	"github.com/zpab123/zaplog"
)

// tcp Client
type TcpClient struct {
	clientMap sync.Map // client 连接集合
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
func (this *TcpClient) OnPacket(client *network.TcpClient, pkt *network.Packet) {
	// 需要发送给客户端
}

// 添加集群服务信息
func (this *TcpClient) AddService(desc *discovery.ServiceDesc) {
	addr := desc.Address()
	zaplog.Debugf("发现新服务，%s", addr)

	cli := network.NewTcpClient(addr)
	cli.SetHandler(this)
	cli.Run()

	this.clientMap.Store(desc.Name, cli)
}

// 移除集群服务信息
func (this *TcpClient) RemoveService(desc *discovery.ServiceDesc) {
	if c, ok := this.clientMap.Load(desc.Name); ok {
		this.clientMap.Delete(desc.Name)
		// 销毁连接对象？
		cli, rok := c.(network.TcpClient)
		if rok {
			cli.Stop()
		}

		zaplog.Debugf("移除服务%s", desc.Address())
	}
}
