// /////////////////////////////////////////////////////////////////////////////
// 常量-接口-types

package network

import (
	"errors"
	"net"
	"time"

	"golang.org/x/net/websocket" // websocket 库
)

// /////////////////////////////////////////////////////////////////////////////
// 常量

// packet 常量
const (
	C_PKT_MID_LEN      uint32 = 2                             // packet 主 id 长度
	C_PKT_LEN_LEN      uint32 = 4                             // packet 长度信息长度
	C_PKT_HEAD_LEN     uint32 = C_PKT_MID_LEN + C_PKT_LEN_LEN // 消息头大小:字节 main_id + length
	C_PKT_BODY_MAX_LEN uint32 = 25 * 1024 * 1024              // body 最大长度 25M
)

// client 状态
const (
	C_CLI_ST_INIT     uint32 = iota // 初始化状态
	C_CLI_ST_SHAKE                  // 握手状态
	C_CLI_ST_WAIT_ACK               // 等待远端握手ACK
	C_CLI_ST_WORKING                // 工作中
	C_CLI_ST_CLOSING                // 正在关闭
	C_CLI_ST_CLOSED                 // 关闭状态
)

// agent 常量
const (
	C_AGENT_HEARTBEAT          = 0 * time.Second // Agent 默认心跳周期
	C_AGENT_ST_INIT     uint32 = iota            // 初始化状态
	C_AGENT_ST_SHAKE                             // 握手状态
	C_AGENT_ST_WAIT_ACK                          // 等待远端握手ACK
	C_AGENT_ST_WORKING                           // 工作中
	C_AGENT_ST_CLOSING                           // 正在关闭
	C_AGENT_ST_CLOSED                            // 关闭状态
)

// conn 常量
const (
	C_CONN_MAX = 10000 // 默认最大连接数
)

// /////////////////////////////////////////////////////////////////////////////
// 变量

// 错误
var (
	V_ERR_BODY_LEN error = errors.New("packet 中 body 长度错误") // packet中body长度错误
)

// /////////////////////////////////////////////////////////////////////////////
// 接口

// acceptor 接口
type IAcceptor interface {
	Run() error                  // 组件开始运行
	Stop() error                 // 组件停止运行
	SetConnMgr(mgr IConnManager) // 设置连接管理
}

// 连接管理
type IConnManager interface {
	ITcpConnManager        // 接口继承： tcp 连接管理
	IWsConnManager         // 接口继承： websocket 连接管理
	Stop()                 // 停止
	SetHandler(h IHandler) // 设置消息处理器
}

// tcp 连接管理
type ITcpConnManager interface {
	OnTcpConn(conn net.Conn) // 收到1个新的 tcp 连接对象
}

// websocket 连接管理
type IWsConnManager interface {
	OnWsConn(wsconn *websocket.Conn) // 收到1个新的 websocket 连接对象
}

// socket 组件
type ISocket interface {
	net.Conn // 接口继承： 符合 Conn 的对象
	Flush() error
}

// Process 接口
type IProcess interface {
	GetHandlerChan() chan *Packet // 获取 handler 通道
	GetRemoteChan() chan *Packet  // 获取 remote 通道
}

// packet 处理器
type IHandler interface {
	OnPacket(a *Agent, pkt *Packet) // 收到1个 packet 消息
}

// Client packet 处理器
type IClientHandler interface {
	OnPacket(c *TcpConn, pkt *Packet) // 收到1个 packet 消息
}

// /////////////////////////////////////////////////////////////////////////////
// TNetOptions 对象

// 网络配置
type NetOptions struct {
	WsAddr    string        // websocket 监听链接 格式 "192.168.1.222:8080"
	MaxConn   uint32        // 最大连接数量，超过此数值后，不再接收新连接
	Heartbeat time.Duration // 心跳周期
}

// 新建1个默认的网络配置
func NewNetOptions() *NetOptions {
	opt := NetOptions{
		MaxConn:   C_CONN_MAX,
		Heartbeat: C_AGENT_HEARTBEAT,
	}

	return &opt
}
