// /////////////////////////////////////////////////////////////////////////////
// app 可选参数

package app

import (
	"github.com/zpab123/sco/network"
	"github.com/zpab123/sco/rpc"
)

// /////////////////////////////////////////////////////////////////////////////
// Options 对象

// app 配置参数
type Options struct {
	AppType   byte               // app 类型
	Name      string             // app名字，不同的app名字不要相同
	Mid       uint16             // 服务ID
	Cluster   bool               // 是否开启集群服务
	Frontend  *network.Frontend  // 前端网络配置
	RpcServer *rpc.ServerOptions // rpc 服务选项
}

// 新建1个默认 Options
func NewOptions() *Options {
	fo := network.NewFrontend()
	rso := rpc.NewServerOptions()

	o := Options{
		AppType:   C_APP_TYPE_FRONTEND,
		Cluster:   false,
		Frontend:  fo,
		RpcServer: rso,
	}

	return &o
}
