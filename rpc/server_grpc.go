// /////////////////////////////////////////////////////////////////////////////
// grpc 服务端

package rpc

import (
	"context"
	"net"

	"github.com/zpab123/sco/protocol"
	"github.com/zpab123/zaplog"
	"google.golang.org/grpc"
)

// /////////////////////////////////////////////////////////////////////////////
// GrpcServer

// grpc 服务
type GrpcServer struct {
	options *GrpcServerOptions // 选项
	server  *grpc.Server       // grpc 服务器
	service *GrpcService       // grpc 消息服务
}

// 新建1个 GrpcServer
func NewGrpcServer(opt *GrpcServerOptions) IServer {
	if nil == opt {
		opt = &GrpcServerOptions{}
	}

	svc := NewGrpcService(opt.remoteService)

	gs := GrpcServer{
		options: opt,
		service: svc,
	}

	return &gs
}

// 启动 rpc 服务
func (this *GrpcServer) Run(ctx context.Context) {
	ln, err := net.Listen("tcp", this.options.Laddr)
	if nil != err {
		return
	}

	this.server = grpc.NewServer()
	this.registerService()

	go this.server.Serve(ln)

	zaplog.Infof("GrpcServer [%s] 启动成功", this.laddr)

	return
}

// graceful: stops the server from accepting new connections and RPCs and
// blocks until all the pending RPCs are finished.
// source: https://godoc.org/google.golang.org/grpc#Server.GracefulStop
func (this *GrpcServer) Stop() {
	this.server.GracefulStop()
}

// 设置引擎服务
func (this *GrpcServer) SetService(rs protocol.GrpcServer) {
	this.service = rs
}

// 注册服务
func (this *GrpcServer) registerService() {
	protocol.RegisterGrpcServer(this.server, this.service)
}
