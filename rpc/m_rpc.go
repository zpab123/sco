// /////////////////////////////////////////////////////////////////////////////
// 常量-接扣-types

package rpc

import (
	"context"

	"github.com/zpab123/sco/discovery"
	//"github.com/zpab123/sco/protocol"
	//"google.golang.org/grpc"
)

// /////////////////////////////////////////////////////////////////////////////
// 常量
const (
	C_SVC_NAME    = "sco.rpcService" // sco rpc服务名称
	C_METHOD_CALL = "Call"           // sco Call 方法
)

// /////////////////////////////////////////////////////////////////////////////
// 接口

// rpc server 服务
type IServer interface {
	Run(ctx context.Context) // 启动服务器
	SetService(svc IService) // 设置 rpc 服务
}

// rpc client 服务
type IClient interface {
	discovery.IListener                                 // 接口继承：服务发现侦听
	Run(ctx context.Context)                            // 启动 client
	HandlerCall(mid uint16, data []byte) (bool, []byte) // handler 调用
	RemoteCall(mid uint16, data []byte) []byte          // remote 调用
}

// handler 服务
type IHandler interface {
	OnData(data []byte) (bool, []byte) // 收到 handler 数据
}

// remote 服务
type IRemoteService interface {
	OnData(data []byte) []byte // 收到 remote 数据
}

// rpc 服务
type IService interface {
	OnHandlerCall(data []byte) (bool, []byte) // Handler 调用
	OnRemoteCall(data []byte) (bool, []byte)  // Remote 调用
}
