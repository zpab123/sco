// /////////////////////////////////////////////////////////////////////////////
// app 可选参数

package app

import (
	"github.com/zpab123/sco/discovery"  // 服务发现
	"github.com/zpab123/sco/netservice" // 网络服务
)

// /////////////////////////////////////////////////////////////////////////////
// Option 对象

// app 配置参数
type Option struct {
	NetServiceOpt     *netservice.TNetServiceOpt   // 网络服务参数
	ClentMsgChanSize  int                          // 客户端消息通道长度
	ServerMsgChanSize int                          // 服务器消息长度
	Cluster           bool                         // 是否启动集群模式
	DiscoveryOpt      *discovery.TEtcdDiscoveryOpt // 服务发现参数
}

// 设置 app 的默认参数
func setdDfaultOpt(app *Application) {
	// 网络服务
	nsOpt := netservice.NewTNetServiceOpt()
	dcOpt := discovery.NewTEtcdDiscoveryOpt()

	opt := &Option{
		NetServiceOpt:     nsOpt,
		ClentMsgChanSize:  C_CLIENT_MSG_CHAN_SIZE,
		ServerMsgChanSize: C_SERVER_MSG_CHAN_SIZE,
		DiscoveryOpt:      dcOpt,
	}

	app.Option = opt
}
