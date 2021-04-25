// /////////////////////////////////////////////////////////////////////////////
// 常量-接口-types

package module

import (
	"context"
)

// /////////////////////////////////////////////////////////////////////////////
// 常量

const (
	C_MSG_TYPE_BROAD  uint8 = 0 // 广播类消息
	C_MSG_TYPE_DIRECT uint8 = 1 // 定向类消息
)

// /////////////////////////////////////////////////////////////////////////////
// 接口

// 模块接口
type IModule interface {
	// 获取 module id
	GetId() uint32
	// 设置模块消息管理者
	SetMsgMgr(mgr IMessgeMgr)
	// 初始化 module
	OnInit()
	// 销毁 module
	OnDestroy()
	// 启动 module
	Run(ctx context.Context)
	// 接收模块消息的通道
	GetMsgChan() chan Messge
}

// 模块消息管理者
type IMessgeMgr interface {
	// 添加消息订阅者
	AddSubscriber(suber *Subscriber)
	// 删除消息订阅者
	DelSubscriber(suber *Subscriber)
	// 广播消息
	// sender=发送者id msgId=消息id data=携带数据
	Broadcast(sender uint32, msgId uint32, data interface{})
	// 向某个模块发送消息
	// sender=发送者id recver=接收者id msgId=消息id data=携带数据
	Post(sender uint32, recver uint32, msgId uint32, data interface{})
}
