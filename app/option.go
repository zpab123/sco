// /////////////////////////////////////////////////////////////////////////////
// app 可选参数

package app

import (
	"github.com/zpab123/sco/network"
)

// /////////////////////////////////////////////////////////////////////////////
// Options 对象

// app 配置参数
type Options struct {
	AppType           byte                        // app 类型
	Cluster           bool                        // 是否开启集群服务
	ClientAcceptorOpt *network.TClientAcceptorOpt // 客户端接收器
}

// 新建1个默认 Options
func NewOptions() *Options {
	copt := network.NewTClientAcceptorOpt()

	o := Options{
		AppType:           C_APP_TYPE_FRONTEND,
		Cluster:           false,
		ClientAcceptorOpt: copt,
	}

	return &o
}
