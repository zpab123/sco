// /////////////////////////////////////////////////////////////////////////////
// 服务器选项

package dispatch

// 服务器选项
type ServerOptions struct {
	Laddr string // 监听地址
}

// 新建1个 *ServerOptions
func NewServerOptions() *ServerOptions {
	o := ServerOptions{}

	return &o
}
