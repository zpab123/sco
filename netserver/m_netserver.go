// /////////////////////////////////////////////////////////////////////////////
// 常量-接口-types

package netserver

import (
	"github.com/zpab123/sco/network" // 网络
)

// /////////////////////////////////////////////////////////////////////////////
// 常量

// netserver 常量
const (
	C_CMPT_NAME = "netserver.netServer" // 组件名字
	C_MAX_CONN  = 100000                // server 默认最大连接数
)

// /////////////////////////////////////////////////////////////////////////////
// TNetServerOpt 对象

// NetServer 组件配置参数
type TNetServerOpt struct {
	Enable       bool   // 是否启动 connector
	AcceptorName string // 接收器名字
	MaxConn      uint32 // 最大连接数量，超过此数值后，不再接收新连接
	ForClient    bool   // 是否面向客户端
}

// 创建1个新的 TNetServerOpt
func NewTNetServerOpt() *TNetServerOpt {
	// 创建对象

	// 创建 TServerOpt
	opt := &TNetServerOpt{
		Enable:       true,
		AcceptorName: network.C_ACCEPTOR_NAME_WS,
		MaxConn:      C_MAX_CONN,
		ForClient:    true,
	}

	return opt
}
