// /////////////////////////////////////////////////////////////////////////////
// grpc 服务端

package rpc

import (
	"context"
	"net"

	"github.com/pkg/errors"
	"github.com/zpab123/sco/protocol"
	"github.com/zpab123/zaplog"
	"google.golang.org/grpc"
)

// /////////////////////////////////////////////////////////////////////////////
// GrpcServer

// grpc 服务
type GrpcServer struct {
	laddr   string         // 监听地址
	server  *grpc.Server   // grpc 服务器
	remote  IRemoteService // remote 服务
	service IService       // rpc 服务
}

// 新建1个 GrpcServer
// 成功：返回 *GrpcServer nil
// 失败：返回 nil error
func NewGrpcServer(laddr string) (*GrpcServer, error) {
	if "" == laddr {
		err := errors.New("参数 laddr 为空")
		return nil, err
	}

	gs := GrpcServer{
		laddr: laddr,
	}

	return &gs, nil
}

// 启动 rpc 服务
func (this *GrpcServer) Run() error {
	ln, err := net.Listen("tcp", this.laddr)
	if nil != err {
		return err
	}

	this.server = grpc.NewServer()
	this.registerService()

	go this.server.Serve(ln)

	zaplog.Infof("[GrpcServer] [%s] 启动成功", this.laddr)

	return nil
}

// 停止 grpc
func (this *GrpcServer) Stop() error {
	this.server.GracefulStop()
	zaplog.Debugf("[GrpcServer] 停止")
	return nil
}

// 设置 rpc 服务
func (this *GrpcServer) SetService(svc IService) {
	if nil != svc {
		this.service = svc
	}
}

// Handler 调用
func (this *GrpcServer) HandlerCall(ctx context.Context, req *protocol.HandlerReq) (*protocol.HandlerRes, error) {
	res := protocol.HandlerRes{}
	if nil == this.service {
		res.Right = true
		return &res, nil
	}

	res.Right, res.Data = this.service.OnHandlerCall(req.Data)

	return &res, nil
}

// Remote 调用
func (this *GrpcServer) RemoteCall(ctx context.Context, req *protocol.RemoteReq) (*protocol.RemoteRes, error) {
	return nil, nil
}

// 注册服务
func (this *GrpcServer) registerService() {
	protocol.RegisterScoServer(this.server, this)
}
