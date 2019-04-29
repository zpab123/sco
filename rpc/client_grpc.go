// /////////////////////////////////////////////////////////////////////////////
// grpc 客户端

package rpc

import (
	"context"
	"sync"
	"time"

	"github.com/zpab123/sco/discovery" // 服务发现
	"github.com/zpab123/zaplog"        // log
	"google.golang.org/grpc"           // grpc
)

// /////////////////////////////////////////////////////////////////////////////
// GrpcClient

// grpc 客户端
type GrpcClient struct {
	name        string        // 名字
	connMap     sync.Map      // rpc 连接集合
	reqTimeout  time.Duration // 请求超时
	dialTimeout time.Duration // 连接超时
}

// 新建1个 GrpcClient
func NewGrpcClient() *GrpcClient {
	gc := &GrpcClient{
		name:        C_RPC_GC,
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

// 名字
func (this *GrpcClient) Name() string {
	return this.name
}

// 远程调用
func (this *GrpcClient) Call() {
	// 数据编码
	// 发送数据
}

// 添加 rpc 服务信息
func (this *GrpcClient) AddService(desc *discovery.ServiceDesc) {
	addr := desc.Address()
	ctx, done := context.WithTimeout(context.Background(), this.dialTimeout)
	defer done()

	conn, err := grpc.DialContext(ctx, addr, grpc.WithInsecure())
	if nil != err {
		return
	}

	zaplog.Debugf("添加新的rpc连接，%s", addr)

	gconn := NewGrpcConn(conn)
	this.connMap.Store(desc.Mid, gconn)
}

// 移除 rpc 服务信息
func (this *GrpcClient) RemoveService(desc *discovery.ServiceDesc) {
	if _, ok := this.connMap.Load(desc.Mid); ok {
		this.connMap.Delete(desc.Mid)
		// 销毁连接对象？
		zaplog.Debugf("移除rpc连接%s", desc.Address())
	}
}
