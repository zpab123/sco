// /////////////////////////////////////////////////////////////////////////////
// grpc 服务端

package rpc

import (
	"net"

	"google.golang.org/grpc"
)

var (
	// sco 底层 rpc 方法
	methods = []grpc.MethodDesc{
		{
			MethodName: C_SVC_CALL,
			Handler:    _ScoCallHandler(),
		},
	}

	// sco 引擎底层 rpc 服务描述
	scoServiceDesc = grpc.ServiceDesc{
		ServiceName: C_SCO_SERVICE,
		HandlerType: (*IScoService)(nil),
		Methods:     methods,
		Streams:     []grpc.StreamDesc{},
		Metadata:    "",
	}
)

// /////////////////////////////////////////////////////////////////////////////
// GrpcServer

// grpc 服务
type GrpcServer struct {
	laddr      string       // 监听地址
	server     *grpc.Server // grpc 服务
	scoService IScoService  // sco 引擎服务
}

// 启动 rpc 服务
func (this *GrpcServer) Run() error {
	ln, err := net.Listen("tcp", this.laddr)
	if nil != err {
		return err
	}

	this.server = grpc.NewServer()
	this.registerScoService()

	go this.server.Serve(ln)

	return nil
}

// 设置引擎服务
func (this *GrpcServer) SetScoService(ss IScoService) {
	this.scoService = ss
}

// 注册服务
func (this *GrpcServer) registerScoService() {
	this.server.RegisterService(scoServiceDesc, this.scoService)
}

// rpc 方法调用请求
func _ScoCallHandler() {

}
