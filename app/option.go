// /////////////////////////////////////////////////////////////////////////////
<<<<<<< HEAD
// app 配置选项
=======
// app 可选参数
>>>>>>> develop

package app

import (
	"github.com/zpab123/sco/netservice" // 网络服务
)

// /////////////////////////////////////////////////////////////////////////////
// Option 对象

<<<<<<< HEAD
// app 自定义参数
type Option struct {
	App           *Application               // app 对象
	NetServiceOpt *netservice.TNetServiceOpt // 网络服务配置
}

// 新建1个默认配置
func NewOption(app *Application) *Option {
	opt := &Option{
		App: app,
	}

	opt.NetServiceOpt = newNetServiceOpt(opt.App)

	return opt
}

// 新建网络服务参数
func newNetServiceOpt(app *Application) *netservice.TNetServiceOpt {
	ns := netservice.NewTNetServiceOpt(app.appDelegate)

	return ns
=======
// app 配置参数
type Option struct {
	NetServiceOpt     *netservice.TNetServiceOpt // 网络服务参数
	ClentMsgChanSize  int                        // 客户端消息通道长度
	ServerMsgChanSize int                        // 服务器消息长度
}

// 设置 app 的默认参数
func setdDfaultOpt(app *Application) {
	// 网络服务
	nsOpt := netservice.NewTNetServiceOpt()

	opt := &Option{
		NetServiceOpt:     nsOpt,
		ClentMsgChanSize:  C_CLIENT_MSG_CHAN_SIZE,
		ServerMsgChanSize: C_SERVER_MSG_CHAN_SIZE,
	}

	app.Option = opt
>>>>>>> develop
}
