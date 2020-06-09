// /////////////////////////////////////////////////////////////////////////////
// 消息分发服务

package dispatch

import (
	"github.com/pkg/errors"
	"github.com/zpab123/sco/network"
)

// tcp 服务器
type TcpServer struct {
	laddr    string               // 监听地址
	acceptor *network.TcpAcceptor // tcp 接收器
	connMgr  *network.ConnMgr     // 连接管理
}

// 新建1个 TcpServer
// 成功： 返回 *TcpServer, nil
// 失败： 返回 nil, error
func NewTcpServer(laddr string) (*TcpServer, error) {
	var err error
	// 参数效验
	if "" == laddr {
		err = errors.New("参数 laddr 为空")
		return nil, err
	}

	a, err := network.NewTcpAcceptor(laddr)
	if nil == err {
		return nil, err
	}

	mgr := network.NewConnMgr(network.C_MAX_CONN)
	a.SetConnMgr(mgr)

	s := TcpServer{
		laddr:    laddr,
		acceptor: a,
		connMgr:  mgr,
	}
	mgr.SetHandler(&s)

	return &s, nil
}

// 启动
func (this *TcpServer) Run() {
	go this.acceptor.Run()
}

// 停止
func (this *TcpServer) Stop() {
	this.acceptor.Stop()
}

// 收到消息 [network.IHandler]
func (this *TcpServer) OnPacket(a *network.Agent, pkt *network.Packet) {

}
