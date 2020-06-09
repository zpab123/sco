// /////////////////////////////////////////////////////////////////////////////
// 消息分发客户端

package dispatch

import (
	"sync"

	//"github.com/pkg/errors"
	"github.com/zpab123/sco/discovery" // 服务发现
	"github.com/zpab123/sco/network"
)

// tcp Client
type TcpClient struct {
	connMap sync.Map // rpc 连接集合
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

// 收到消息 [network.IHandler]
func (this *TcpClient) OnPacket(a *network.Agent, pkt *network.Packet) {

}

// 添加集群服务信息
func (this *TcpClient) AddService(desc *discovery.ServiceDesc) {

}

// 移除集群服务信息
func (this *TcpClient) RemoveService(desc *discovery.ServiceDesc) {

}
