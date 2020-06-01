// /////////////////////////////////////////////////////////////////////////////
// 常量-接扣-types

package rpc

import (
	"context"

	"github.com/zpab123/sco/discovery"
	"github.com/zpab123/sco/protocol"
)

// /////////////////////////////////////////////////////////////////////////////
// 接口

// rpc server 服务
type IServer interface {
	Run(ctx context.Context)
	SetRpcService(rs protocol.ScoGrpcServer)
}

// rpc client 服务
type IClient interface {
	discovery.IListener // 接口继承：服务发现侦听
}
