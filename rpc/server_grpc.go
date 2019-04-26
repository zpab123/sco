// /////////////////////////////////////////////////////////////////////////////
// grpc 服务端

package rpc

import (
	goContext "context"
	"net"

	"github.com/zpab123/sco/protocol" // 消息协议
	"github.com/zpab123/zaplog"       // log
	"google.golang.org/grpc"          // grpc
)

// /////////////////////////////////////////////////////////////////////////////
// GrpcServer

// grpc 服务
type GrpcServer struct {
	name       string                 // 组件名字
	laddr      string                 // 监听地址
	server     *grpc.Server           // grpc 服务
	rpcService protocol.ScoGrpcServer // sco rpc服务
}

// 新建1个 GrpcServer
func NewGrpcServer(laddr string) IServer {
	gs := &GrpcServer{
		laddr: laddr,
	}

	return gs
}

// 启动 rpc 服务
func (this *GrpcServer) Run(ctx goContext.Context) {
	ln, err := net.Listen("tcp", this.laddr)
	if nil != err {
		return
	}

	this.server = grpc.NewServer()
	this.registerRpcService()

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

func (this *GrpcServer) Name() string {
	return this.name
}

// 设置引擎服务
func (this *GrpcServer) SetRpcService(rs protocol.ScoGrpcServer) {
	this.rpcService = rs
}

// 注册服务
func (this *GrpcServer) registerRpcService() {
	protocol.RegisterScoGrpcServer(this.server, this.rpcService)
}
