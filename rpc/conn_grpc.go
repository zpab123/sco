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
	address   string             // 远端地址
	connected bool               // 是否连接
	lock      sync.Mutex         // 锁
	client    protocol.ScoClient // 客户端
	conn      *grpc.ClientConn   // rpc conn
}

// 新建1个 GrpcConn 对象
func NewGrpcConn(addr string) *GrpcConn {
	gc := &GrpcConn{
		address: addr,
	}

	return gc
}

// Handler 调用
func (this *GrpcConn) HandlerCall(ctx context.Context, req *protocol.HandlerReq) (*protocol.HandlerRes, error) {
	if !this.connected {
		if err := this.connect(); nil != err {
			return nil, err
		}
	}

	return this.client.HandlerCall(ctx, req)
}

// Remote 调用
func (this *GrpcConn) RemoteCall(ctx context.Context, req *protocol.RemoteReq) (*protocol.RemoteRes, error) {
	if !this.connected {
		if err := this.connect(); nil != err {
			return nil, err
		}
	}

	return this.client.RemoteCall(ctx, req)
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

	this.client = protocol.NewScoClient(conn)
	this.conn = conn
	this.connected = true
	return nil
}
