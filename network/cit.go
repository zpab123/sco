// /////////////////////////////////////////////////////////////////////////////
// 常量-接口-types

package network

import (
	"errors"
	"net"
	"time"

	"golang.org/x/net/websocket"
)

// /////////////////////////////////////////////////////////////////////////////
// 常量

// packet 常量
const (
	C_PKT_KIND_LEN     uint32 = 1                                   // packet kind 长度
	C_PKT_KIND_END     uint32 = C_PKT_KIND_LEN                      // packet kind 结束位置 1
	C_PKT_CLIENT_LEN   uint32 = 4                                   // packet client 长度
	C_PKT_CLIENT_END   uint32 = C_PKT_KIND_END + C_PKT_CLIENT_LEN   // packet client 结束位置 1 + 4
	C_PKT_SENDER_LEN   uint32 = 2                                   // packet sender 长度
	C_PKT_SENDER_END   uint32 = C_PKT_CLIENT_END + C_PKT_SENDER_LEN // packet client 结束位置 5 + 2
	C_PKT_SID_LEN      uint32 = 2                                   // packet sid 长度
	C_PKT_SID_END      uint32 = C_PKT_SENDER_END + C_PKT_SID_LEN    // packet sid 结束位置 7 + 2
	C_PKT_MID_LEN      uint32 = 2                                   // packet mid 长度
	C_PKT_MID_END      uint32 = C_PKT_SID_END + C_PKT_MID_LEN       // packet mid 长度 9 +2
	C_PKT_LEN_LEN      uint32 = 4                                   // packet 长度信息长度
	C_PKT_BODY_MAX_LEN uint32 = 25 * 1024 * 1024                    // body 最大长度 25M
	C_PKT_HEAD_LEN     uint32 = C_PKT_MID_END + C_PKT_LEN_LEN       // 消息头大小
)

// packet 常量
const (
	C_PKT_KIND_CONN    byte = 1   // 连接消息
	C_PKT_KIND_SCO     byte = 2   // 引擎消息
	C_PKT_KIND_CLI_SER byte = 100 // client -> server
	C_PKT_KIND_SER_SER byte = 101 // server -> server
	C_PKT_KIND_SER_CLI byte = 102 // server -> client
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
	C_AGENT_ST_INIT     uint32 = iota // 初始化状态
	C_AGENT_ST_SHAKE                  // 握手状态
	C_AGENT_ST_WAIT_ACK               // 等待远端握手ACK
	C_AGENT_ST_WORKING                // 工作中
	C_AGENT_ST_CLOSING                // 正在关闭
	C_AGENT_ST_CLOSED                 // 关闭状态
)

// 前端常量
const (
	C_F_MAX_CONN  int32         = 10000              // 默认最大连接数
	C_F_KEY       string        = "scob9kxH6FdqOKnA" // 握手 key
	C_F_HEARTBEAT time.Duration = 3 * time.Second    //  心跳周期
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
	Run() error  // 开始运行
	Stop() error // 停止运行
}

// agent 管理
type IAgentManager interface {
	ITcpConnManager               // 接口继承： tcp 连接管理
	IWsConnManager                // 接口继承： websocket 连接管理
	Run()                         // 启动
	Stop()                        // 停止
	SetKey(k string)              // 设置握手key
	SetHeartbeat(h time.Duration) // 设置心跳
	SetHandler(h IHandler)        // 设置消息处理器
	SetPacketChan(chan *Packet)   // 设置消息通道
	GetConnNum() int32            // 获取当前连接数
	OnAgentStop(a *Agent)         // 某个 Agent 停止
}

// 连接管理
type IConnManager interface {
	ITcpConnManager               // 接口继承： tcp 连接管理
	IWsConnManager                // 接口继承： websocket 连接管理
	Run()                         // 启动
	Stop()                        // 停止
	SetKey(k string)              // 设置握手key
	SetHeartbeat(h time.Duration) // 设置心跳
	SetHandler(h IHandler)        // 设置消息处理器
	SetPacketChan(chan *Packet)   // 设置消息通道
	GetConnNum() int32            // 获取当前连接数
}

// tcp 连接管理
type ITcpConnManager interface {
	OnTcpConn(conn net.Conn) // 收到1个新的 tcp 连接对象
}

// websocket 连接管理
type IWsConnManager interface {
	OnWsConn(wsconn *websocket.Conn) // 收到1个新的 websocket 连接对象
}

// packet 处理器
type IHandler interface {
	OnPacket(pkt *Packet) // 收到1个 packet 消息
}

// Client packet 处理器
type IClientHandler interface {
	OnPacket(c *TcpConn, pkt *Packet) // 收到1个 packet 消息
}

// 连接对象接口
type IConn interface {
	Stop()                  // 停止
	Send(pkt *Packet) error // 发送 packet
}

// /////////////////////////////////////////////////////////////////////////////
// Frontend 对象

// 前端网络配置
type Frontend struct {
	TcpAddr   string        // tcp 监听链接 格式 "192.168.1.222:8080"
	WsAddr    string        // websocket 监听链接 格式 "192.168.1.222:8080"
	MaxConn   int32         // 最大连接数量，超过此数值后，不再接收新连接
	Key       string        // 握手key
	Heartbeat time.Duration // 心跳周期
}

// 新建1个默认的网络配置
func NewFrontend() *Frontend {
	opt := Frontend{
		MaxConn:   C_F_MAX_CONN,
		Key:       C_F_KEY,
		Heartbeat: C_F_HEARTBEAT,
	}

	return &opt
}
