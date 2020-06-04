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

// /////////////////////////////////////////////////////////////////////////////
// 变量

// 错误
var (
	V_ERR_BODY_LEN error = errors.New("packet中body长度错误") // packet中body长度错误
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
	IWsConnManager // 接口继承： websocket 连接管理
	Stop()         // 停止
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

// Process 接口
type IProcess interface {
	GetHandlerChan() chan *Packet // 获取 handler 通道
	GetRemoteChan() chan *Packet  // 获取 remote 通道
}

// packet 处理器
type IHandler interface {
	OnPacket(agent *Agent, pkt *Packet) // 收到1个 packet 消息
}

// /////////////////////////////////////////////////////////////////////////////
// TNetOptions 对象

// 网络配置
type TNetOptions struct {
	WsAddr    string        // websocket 监听链接 格式 "192.168.1.222:8080"
	MaxConn   uint32        // 最大连接数量，超过此数值后，不再接收新连接
	Heartbeat time.Duration // 心跳周期
	Handler   IHandler      // 消息处理器
}

// 新建1个默认的网络配置
func NewTNetOptions() *TNetOptions {
	opt := TNetOptions{
		MaxConn:   C_MAX_CONN,
		Heartbeat: C_HEARTBEAT,
	}

	return &opt
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
		ShakeKey:      "sco",
	}

	return &opt
}

// /////////////////////////////////////////////////////////////////////////////
// TAgentOpt 对象

// Agent 配置参数
type TAgentOpt struct {
	Heartbeat  time.Duration // 心跳周期
	Handler    IHandler      // 消息处理器
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
