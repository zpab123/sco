// /////////////////////////////////////////////////////////////////////////////
// app 可选参数

package app

import (
	"github.com/zpab123/sco/discovery"
	"github.com/zpab123/sco/network"
	"github.com/zpab123/sco/rpc"
)

// /////////////////////////////////////////////////////////////////////////////
// Options 对象

// app 配置参数
type Options struct {
	AppType   byte               // app 类型
	Id        string             // app 唯一标识
	Mid       uint16             // 服务ID
	Cluster   bool               // 是否开启集群服务
	Frontend  *network.Frontend  // 前端网络配置ss
	RpcServer *rpc.ServerOptions // rpc 服务选项
	Discovery *discovery.Options // 服务发现选项
}

// 新建1个默认 Options
func NewOptions() *Options {
	fo := network.NewFrontend()
	rso := rpc.NewServerOptions()
	dis := discovery.NewOptions()

	o := Options{
		AppType:   C_APP_TYPE_FRONTEND,
		Cluster:   false,
		Frontend:  fo,
		RpcServer: rso,
		Discovery: dis,
	}

	return &o
}
