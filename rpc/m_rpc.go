// /////////////////////////////////////////////////////////////////////////////
// 常量-接扣-types

package rpc

import (
	"github.com/zpab123/sco/discovery"
)

// /////////////////////////////////////////////////////////////////////////////
// 常量

// /////////////////////////////////////////////////////////////////////////////
// 接口

// rpc server 服务
type IServer interface {
	Run() error              // 启动服务器
	Stop() error             // 停止服务器
	SetService(svc IService) // 设置 rpc 服务
}

// rpc client 服务
type IClient interface {
	discovery.IListener                                 // 接口继承：服务发现侦听
	Run() error                                         // 启动
	Stop() error                                        // 停止
	HandlerCall(mid uint16, data []byte) (bool, []byte) // handler 调用
	RemoteCall(mid uint16, data []byte) []byte          // remote 调用
}

// rpc 服务
type IService interface {
	OnHandlerCall(data []byte) (bool, []byte) // Handler 调用
	OnRemoteCall(data []byte) []byte          // Remote 调用
}
