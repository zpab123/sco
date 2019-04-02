// /////////////////////////////////////////////////////////////////////////////
// 常量-接口-types

package netservice

import (
	"github.com/zpab123/sco/model"   // 全局模型
	"github.com/zpab123/sco/network" // 网络
	"github.com/zpab123/sco/session" // session 组件
)

// /////////////////////////////////////////////////////////////////////////////
// 常量

// netserver 常量
const (
	C_CMPT_NAME = "netservice" // 组件名字
	C_MAX_CONN  = 100000       // server 默认最大连接数
)

// /////////////////////////////////////////////////////////////////////////////
// 接口

// 接收器接口
type INetService interface {
	model.IComponent // 接口继承：组件接口
}

// /////////////////////////////////////////////////////////////////////////////
// TNetServiceOpt 对象

// NetServer 组件配置参数
type TNetServiceOpt struct {
	Enable       bool                       // 是否启动 connector
	AcceptorName string                     // 接收器名字
	Acceptor     network.IAcceptor          // 接收器
	MaxConn      uint32                     // 最大连接数量，超过此数值后，不再接收新连接
	ForClient    bool                       // 是否面向客户端
	TcpConnOpt   *model.TTcpConnOpt         // tcpSocket 配置参数
	ClientSesOpt *session.TClientSessionOpt // ClientSession 配置参数
	ServerSesOpt *session.TServerSessionOpt // ServerSession 配置参数
}

// 创建1个新的 TNetServiceOpt
func NewTNetServiceOpt(handler session.IMsgHandler) *TNetServiceOpt {
	// 创建对象
	tcpOpt := model.NewTTcpConnOpt()

	csOpt := session.NewTClientSessionOpt(handler)
	ssOpt := session.NewTServerSessionOpt(handler)

	// 创建 TServerOpt
	opt := &TNetServiceOpt{
		Enable:       true,
		AcceptorName: network.C_ACCEPTOR_NAME_WS,
		MaxConn:      C_MAX_CONN,
		ForClient:    true,
		TcpConnOpt:   tcpOpt,
		ClientSesOpt: csOpt,
		ServerSesOpt: ssOpt,
	}

	return opt
}
