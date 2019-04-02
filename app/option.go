// /////////////////////////////////////////////////////////////////////////////
// app 配置选项

package app

import (
	"github.com/zpab123/sco/netservice" // 网络服务
)

// /////////////////////////////////////////////////////////////////////////////
// Option 对象

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
}
