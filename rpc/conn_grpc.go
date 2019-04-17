// /////////////////////////////////////////////////////////////////////////////
// grpc 连接对象

package rpc

import (
	"context"
	"fmt"

	"github.com/zpab123/sco/protocol" // 消息协议
	"google.golang.org/grpc"          // grpc
)

// /////////////////////////////////////////////////////////////////////////////
// GrpcConn

// grpc 连接对象
type GrpcConn struct {
	clinetConn *grpc.ClientConn // grpc 连接对象
}

// 新建1个 GrpcConn 对象
func NewGrpcConn(conn *grpc.ClientConn) IConn {
	gc := &GrpcConn{
		clinetConn: conn,
	}

	return gc
}

// 远程调用1个方法
func (this *GrpcConn) Call(ctx context.Context, req *protocol.RpcRequest, opts ...grpc.CallOption) (*protocol.RpcResponse, error) {
	res := new(protocol.RpcResponse)
	method := fmt.Sprintf("/%s/%s", C_SVC_NAME, C_METHOD_CALL)
	err := this.clinetConn.Invoke(ctx, method, req, res, opts...)
	if nil != err {
		return nil, err
	}

	return res, nil
}
