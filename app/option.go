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
	AppType   byte                 // app 类型
	ServiceId uint16               // 服务ID
	Cluster   bool                 // 是否开启集群服务
	NetOpt    *network.TNetOptions // 网络配置（客户端）
}

// 新建1个默认 Options
func NewOptions() *Options {
	nopt := network.NewTNetOptions()

	o := Options{
		AppType: C_APP_TYPE_FRONTEND,
		Cluster: false,
		NetOpt:  nopt,
	}

	return &o
}
