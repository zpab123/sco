// /////////////////////////////////////////////////////////////////////////////
// 常量-接口-types

package network

import (
	"net"

	"golang.org/x/net/websocket" // websocket 库
)

// /////////////////////////////////////////////////////////////////////////////
// 常量

// socket_buff 常量
const (
	C_BUFF_READ_SIZE  = 16384  // scoket 读取类 buff 长度
	C_BUFF_WRITE_SIZE = 16384  // scoket 写入类 buff 长度
	C_MAX_CONN        = 100000 // ClientAcceptor 默认最大连接数
)

// packet 常量
const (
	C_PKT_MID_LEN      uint16 = 2                              // packet 主 id 长度 (uin16)
	C_PKT_LEN_LEN      uint16 = 2                              // packet 长度信息长度 (uin16)
	C_PKT_HEAD_LEN     uint16 = C_PKT_MID_LEN + C_PKT_LEN_LEN  // 消息头大小:字节 main_id(2字节) + length(2字节)
	C_PKT_MAX_LEN      uint16 = 65535                          // 最大单个 packet 数据，= head + body = 64k
	C_PKT_BODY_MAX_LEN uint16 = C_PKT_MAX_LEN - C_PKT_HEAD_LEN // body 最大长度 （总长度-消息头）
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

// socket 组件
type ISocket interface {
	net.Conn // 接口继承： 符合 Conn 的对象
	Flush() error
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

// /////////////////////////////////////////////////////////////////////////////
// TClientAcceptorOpt 对象

// ClientAcceptor 组件配置参数
type TClientAcceptorOpt struct {
	Enable  bool   // 是否启用
	WsAddr  string // websocket 监听链接 格式 "192.168.1.222:8080"
	MaxConn uint32 // 最大连接数量，超过此数值后，不再接收新连接
}

// 创建1个新的 TNetServiceOpt
func NewTClientAcceptorOpt() *TClientAcceptorOpt {
	// 创建 TServerOpt
	opt := TClientAcceptorOpt{
		Enable:  true,
		MaxConn: C_MAX_CONN,
	}

	return &opt
}
