// /////////////////////////////////////////////////////////////////////////////
// 常量-接扣-types

package rpc

import (
	"github.com/zpab123/sco/protocol" // 消息协议
	"golang.org/x/net/context"        // golang 上下文
)

// /////////////////////////////////////////////////////////////////////////////
// 常量
const (
	C_SCO_SERVICE = "rpc.scoService" // sco rpc服务名称
	C_SVC_CALL    = "Call"           // sco Call 方法
)

// /////////////////////////////////////////////////////////////////////////////
// 接口

// rpc server 服务
type IServer interface {
	SetScoService(ss IScoService) // 设置引擎服务
}

// sco 引擎服务
type IScoService interface {
	Call(ctx context.Context, req *protocol.RpcRequest) (*protocol.RpcResponse, error) // 方法调用
}
