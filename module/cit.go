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
	// 订阅消息
	// mod=订阅者 msgId=消息id
	Subscribe(mod IModule, msgId uint32, ch chan Messge)
	// 取消订阅
	// mod=订阅者 msgId=消息id
	Unsubscribe(mod IModule, msgId uint32)
	// 广播消息
	// mod=发送者 msgId=消息id data=携带数据
	Broadcast(mod IModule, msgId uint32, data interface{})
	// 向某个模块发送消息
	// mod=发送者 recver=接收者id msgId=消息id data=携带数据
	Post(mod IModule, recver uint32, msgId uint32, data interface{})
}
