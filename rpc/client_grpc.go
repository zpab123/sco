// /////////////////////////////////////////////////////////////////////////////
// grpc 客户端

package rpc

import (
	"context"
	"sync"
	"time"

	"github.com/zpab123/sco/discovery" // 服务发现
	"google.golang.org/grpc"           // grpc
)

// /////////////////////////////////////////////////////////////////////////////
// GrpcClient

// grpc 客户端
type GrpcClient struct {
	connMap     sync.Map      // rpc 连接集合
	reqTimeout  time.Duration // 请求超时
	dialTimeout time.Duration // 连接超时
}

func NewGrpcClient() {

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

	gconn := NewGrpcConn(conn)
	this.connMap.Store(desc.Mid, gconn)
}

// 移除 rpc 服务信息
func (this *GrpcClient) RemoveService(desc *discovery.ServiceDesc) {
	if _, ok := this.connMap.Load(desc.Mid); ok {
		this.connMap.Delete(desc.Mid)
	}
}
