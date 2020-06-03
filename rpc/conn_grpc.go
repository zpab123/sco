// /////////////////////////////////////////////////////////////////////////////
// grpc 连接对象

package rpc

import (
	"context"
	"sync"

	"github.com/zpab123/sco/protocol"
	//"github.com/zpab123/zaplog"
	"google.golang.org/grpc"
)

// /////////////////////////////////////////////////////////////////////////////
// GrpcConn

// grpc 连接对象
type GrpcConn struct {
	address   string              // 远端地址
	connected bool                // 是否连接
	lock      sync.Mutex          // 锁
	client    protocol.GrpcClient // 客户端
	conn      *grpc.ClientConn    // rpc conn
}

// 新建1个 GrpcConn 对象
func NewGrpcConn(address string) *GrpcConn {
	gc := &GrpcConn{
		address: address,
	}

	return gc
}

// 远程调用1个方法
func (this *GrpcConn) Call(ctx context.Context, req *protocol.GrpcRequest, opts ...grpc.CallOption) (*protocol.GrpcResponse, error) {
	/*
		res := new(protocol.GrpcResponse)
		method := fmt.Sprintf("/%s/%s", C_SVC_NAME, C_METHOD_CALL)
		err := this.clinetConn.Invoke(ctx, method, req, res, opts...)
		if nil != err {
			return nil, err
		}
	*/

	if !this.connected {
		if err := this.connect(); nil != err {
			return nil, err
		}
	}

	return this.client.Call(ctx, req)
}

// 连接远端
func (this *GrpcConn) connect() error {
	this.lock.Lock()
	defer this.lock.Unlock()

	if this.connected {
		return nil
	}

	conn, err := grpc.Dial(this.address, grpc.WithInsecure())
	if nil != err {
		return err
	}

	this.client = protocol.NewGrpcClient(conn)
	this.conn = conn
	this.connected = true
	return nil
}
