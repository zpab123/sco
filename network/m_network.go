// /////////////////////////////////////////////////////////////////////////////
// 常量-接口-types

package network

import (
	"net"

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

// socket_buff 常量
const (
	C_BUFF_READ_SIZE  = 16384 // scoket 读取类 buff 长度
	C_BUFF_WRITE_SIZE = 16384 // scoket 写入类 buff 长度
)

// packet 常量
const (
	C_PKT_HEAD_LEN = 6                // 消息头大小:字节 type(2字节) + length(4字节)
	C_PKT_MAX_LEN  = 25 * 1024 * 1024 // 最大单个 packet 数据，= head + body = 25M
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

// socket 组件
type ISocket interface {
	net.Conn // 接口继承： 符合 Conn 的对象
	Flush() error
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
// TBufferSocketOpt 对象

// BufferSocket 配置参数
type TBufferSocketOpt struct {
	ReadBufferSize  int // 读取 buffer 字节大小
	WriteBufferSize int // 写入 buffer 字节大小
}

// 新建1个 TBufferSocketOpt 对象
func NewTBufferSocketOpt() *TBufferSocketOpt {
	bs := &TBufferSocketOpt{
		ReadBufferSize:  C_BUFF_READ_SIZE,
		WriteBufferSize: C_BUFF_WRITE_SIZE,
	}

	return bs
}
