// /////////////////////////////////////////////////////////////////////////////
// 常量-接口-types

package app

import (
	"time"

	"github.com/zpab123/sco/session" // session 管理
)

// /////////////////////////////////////////////////////////////////////////////
// 常量

const (
	C_STOP_OUT_TIME        = 30 * time.Second // 关闭app的时候，超过此时间，就会强制关闭
	C_CLIENT_MSG_CHAN_SIZE = 10000            // 消息 chan 默认长度
	C_SERVER_MSG_CHAN_SIZE = 10000            // 消息 chan 默认长度
)

// /////////////////////////////////////////////////////////////////////////////
// 接口

// App 代理
type IDelegate interface {
	Init(app *Application)        // app 初始化
	OnClentMsg(session.ClientMsg) // 收到1个客户端消息
}

// /////////////////////////////////////////////////////////////////////////////
// TBaseInfo 对象

// app 启动信息
type TBaseInfo struct {
	AppType  string    // App 类型
	MainPath string    // main 程序所在路径
	Env      string    // 运行环境 production= 开发环境 development = 运营环境
	Name     string    // App 名字
	RunTime  time.Time // 启动时间
}
