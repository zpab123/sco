// /////////////////////////////////////////////////////////////////////////////
// 常量-接口-types

package network

import (
	"golang.org/x/net/websocket" // websocket 库
)

// /////////////////////////////////////////////////////////////////////////////
// 常量

// acceptor 类型
const (
	C_ACCEPTOR_NAME_TCP = "tcpAcceptor"  // 支持 tcp
	C_ACCEPTOR_NAME_WS  = "wsAcceptor"   // 支持 websocket
	C_ACCEPTOR_NAME_MUL = "multiformity" // tcpAcceptor + wsAcceptor 组合
	C_ACCEPTOR_NAME_COM = "composite"    // 同时支持 tcp 和 websocket
)

// server 常量
const (
	C_NET_SERVER_CMPT_NAME = "network.NetServer" // 组件名字
	C_NET_SERVER_MAX_CONN  = 100000              // server 默认最大连接数
)

// /////////////////////////////////////////////////////////////////////////////
// 接口

// acceptor 接口
type IAcceptor interface {
	Run() error  // 组件开始运行
	Stop() error // 组件停止运行
}

// websocket 连接管理
type IWsConnManager interface {
	OnNewWsConn(wsconn *websocket.Conn) // 收到1个新的 websocket 连接对象
}

// 连接管理
type IConnManager interface {
	IWsConnManager // websocket 连接管理
}

// /////////////////////////////////////////////////////////////////////////////
// Laddr 对象

// 监听地址集合
type TLaddr struct {
	TcpAddr string // Tcp 监听地址：格式 192.168.1.1:8600
	WsAddr  string // websocket 监听地址: 格式 192.168.1.1:8600
	UdpAddr string // udp 监听地址: 格式 192.168.1.1:8600
	KcpAddr string // kcp 监听地址: 格式 192.168.1.1:8600
}

// /////////////////////////////////////////////////////////////////////////////
// TNetServerOpt 对象

// NetServer 组件配置参数
type TNetServerOpt struct {
	Enable       bool   // 是否启动 connector
	AcceptorName string // 接收器名字
	MaxConn      uint32 // 最大连接数量，超过此数值后，不再接收新连接
	ForClient    bool   // 是否面向客户端
}

// 创建1个新的 TServerOpt
func NewTNetServerOpt() *TNetServerOpt {
	// 创建对象

	// 创建 TServerOpt
	opt := &TNetServerOpt{
		Enable:       true,
		AcceptorName: C_ACCEPTOR_NAME_WS,
		MaxConn:      C_NET_SERVER_MAX_CONN,
		ForClient:    true,
	}

	return opt
}
