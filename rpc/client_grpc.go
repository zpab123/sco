// /////////////////////////////////////////////////////////////////////////////
// grpc 客户端

package rpc

import (
	"context"
	"sync"
	"time"

	"github.com/zpab123/sco/discovery" // 服务发现
	"github.com/zpab123/sco/protocol"  // 消息协议
	"github.com/zpab123/zaplog"        // log
	//"google.golang.org/grpc"           // grpc
)

// /////////////////////////////////////////////////////////////////////////////
// GrpcClient

// grpc 客户端
type GrpcClient struct {
	connMap     sync.Map      // rpc 连接集合
	reqTimeout  time.Duration // 请求超时
	dialTimeout time.Duration // 连接超时
}

// 新建1个 GrpcClient
func NewGrpcClient() IClient {
	gc := &GrpcClient{
		reqTimeout:  5 * time.Second,
		dialTimeout: 5 * time.Second,
	}

	return gc
}

// 启动
func (this *GrpcClient) Run(ctx context.Context) {
	zaplog.Debugf("GrpcClient 启动成功")
}

// 停止
func (this *GrpcClient) Stop() {

}

// Handler 调用
func (this *GrpcClient) HandlerCall(mid uint16, data []byte) (bool, []byte) {
	req := protocol.HandlerReq{
		Data: data,
	}

	if mid != 201 {
		return true, nil
	}

	c, ok := this.connMap.Load("chat_1")
	if !ok {
		return true, nil
	}

	res, err := c.(*GrpcConn).HandlerCall(context.Background(), &req)
	if nil != err {
		zaplog.Debugf("[GrpcClient] HandlerCall_err=%s", err.Error())
		return true, nil
	}

	return res.Right, res.Data
}

// 远程调用
func (this *GrpcClient) RemoteCall(mid uint16, data []byte) []byte {
	req := protocol.RemoteReq{
		Data: data,
	}

	c, ok := this.connMap.Load("chat_1")
	if !ok {
		return nil
	}

	res, err := c.(*GrpcConn).RemoteCall(context.Background(), &req)
	if nil != err {
		zaplog.Debugf("[GrpcClient] RemoteCall_err=%s", err.Error())
		return nil
	}

	return res.Data
}

// 添加 rpc 服务信息
func (this *GrpcClient) AddService(desc *discovery.ServiceDesc) {
	addr := desc.Address()
	/*
		ctx, done := context.WithTimeout(context.Background(), this.dialTimeout)
		defer done()

		conn, err := grpc.DialContext(ctx, addr, grpc.WithInsecure())
		if nil != err {
			return
		}
	*/

	zaplog.Debugf("添加新的rpc连接，%s", addr)

	gconn := NewGrpcConn(addr)
	err := gconn.connect()
	if nil != err {
		zaplog.Debugf("GrpcClient")
		zaplog.Debugf(err.Error())
	}

	this.connMap.Store(desc.Name, gconn)
}

// 移除 rpc 服务信息
func (this *GrpcClient) RemoveService(desc *discovery.ServiceDesc) {
	if _, ok := this.connMap.Load(desc.Name); ok {
		this.connMap.Delete(desc.Name)
		// 销毁连接对象？
		zaplog.Debugf("移除rpc连接%s", desc.Address())
	}
}
