// /////////////////////////////////////////////////////////////////////////////
// 常量-接口-types

package network

import (
	"net"
	"time"

	"github.com/pkg/errors"      // 异常库
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
	C_PKT_MID_LEN      uint32 = 2                             // packet 主 id 长度
	C_PKT_LEN_LEN      uint32 = 4                             // packet 长度信息长度
	C_PKT_HEAD_LEN     uint32 = C_PKT_MID_LEN + C_PKT_LEN_LEN // 消息头大小:字节 main_id + length
	C_PKT_BODY_MAX_LEN uint32 = 25 * 1024 * 1024              // body 最大长度 25M
)

// ScoConn 状态
const (
	C_CONN_STATE_INIT     uint32 = iota // 初始化状态
	C_CONN_STATE_SHAKE                  // 握手状态
	C_CONN_STATE_WAIT_ACK               // 等待远端握手ACK
	C_CONN_STATE_WORKING                // 工作中
	C_CONN_STATE_CLOSING                // 正在关闭
	C_CONN_STATE_CLOSED                 // 关闭状态
)

const (
	C_HEARTBEAT = 0 * time.Second // Agent 默认心跳周期
)

// 错误
const (
	C_ERR_BODY_LEN error = errors.New("packet中body长度错误") // packet中body长度错误
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
	Enable   bool       // 是否启用
	WsAddr   string     // websocket 监听链接 格式 "192.168.1.222:8080"
	MaxConn  uint32     // 最大连接数量，超过此数值后，不再接收新连接
	AgentOpt *TAgentOpt // Agent 配置参数
}

// 创建1个新的 TNetServiceOpt
func NewTClientAcceptorOpt() *TClientAcceptorOpt {
	ao := NewTAgentOpt()

	// 创建 TServerOpt
	opt := TClientAcceptorOpt{
		Enable:   true,
		MaxConn:  C_MAX_CONN,
		AgentOpt: ao,
	}

	return &opt
}

// /////////////////////////////////////////////////////////////////////////////
// TScoConnOpt 对象

// ScoConn 配置参数
type TScoConnOpt struct {
	ShakeKey      string            // 握手key
	Heartbeat     uint32            // 心跳间隔，单位：秒。0=不设置心跳
	BuffSocketOpt *TBufferSocketOpt // BufferSocket 配置参数
}

// 新建1个 WorldConnection 对象
func NewTScoConnOpt() *TScoConnOpt {
	bufopt := NewTBufferSocketOpt()

	opt := TScoConnOpt{
		BuffSocketOpt: bufopt,
	}

	return &opt
}

// /////////////////////////////////////////////////////////////////////////////
// TAgentOpt 对象

// Agent 配置参数
type TAgentOpt struct {
	Heartbeat  time.Duration // 心跳周期
	ScoConnOpt *TScoConnOpt  // ScoConn 配置参数
}

// 创建1个默认的 TAgentOpt
func NewTAgentOpt() *TAgentOpt {
	so := NewTScoConnOpt()

	ao := TAgentOpt{
		Heartbeat:  C_HEARTBEAT,
		ScoConnOpt: so,
	}

	return &ao
}
