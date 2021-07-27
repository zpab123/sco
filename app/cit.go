// /////////////////////////////////////////////////////////////////////////////
// 常量-接口-types

package app

import (
	"time"

	"github.com/zpab123/sco/network"
)

// /////////////////////////////////////////////////////////////////////////////
// 常量

const (
	C_STOP_TIME_OUT = 10 * time.Second // 关闭app的时候，超过此时间，就会强制关闭
)

// /////////////////////////////////////////////////////////////////////////////
// 接口

// App 代理
type IDelegate interface {
	Init(app *Application)        // app 初始化
	Working()                     // 启动
	OnPacket(pkt *network.Packet) // 收到1个客户端消息
}

// /////////////////////////////////////////////////////////////////////////////
// types
