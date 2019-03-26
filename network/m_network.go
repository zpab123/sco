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
