// /////////////////////////////////////////////////////////////////////////////
// scorpio 轻度游戏服务器框架

package sco

import (
	"github.com/zpab123/sco/app" // 1个通用服务器库
)

// /////////////////////////////////////////////////////////////////////////////
// 对外 api

// 创建1个新的 Application 对象
//
// appType=server.json 中配置的类型
func CreateApp(appType string, delegate app.IDelegate) *app.Application {
	// 创建 app
	app := app.NewApplication(appType, delegate)
	app.Init()

	return app
}
