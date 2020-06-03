// /////////////////////////////////////////////////////////////////////////////
// 配置选项

package rpc

// /////////////////////////////////////////////////////////////////////////////
// ServerOptions 对象

// 服务器配置选项
type ServerOptions struct {
	Laddr         string         // 服务器监听地址
	remoteService IRemoteService // remote 服务
}

// 新建1个 ServerOptions
func NewServerOptions() *ServerOptions {
	so := ServerOptions{}
	return &so
}

// /////////////////////////////////////////////////////////////////////////////
// Options 对象

// grpc 服务选项
type GrpcServerOptions struct {
	Laddr         string         // 服务器监听地址
	RemoteService IRemoteService // remote 服务
}
