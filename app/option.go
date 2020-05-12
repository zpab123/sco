// /////////////////////////////////////////////////////////////////////////////
// app 可选参数

package app

// /////////////////////////////////////////////////////////////////////////////
// Options 对象

// app 配置参数
type Options struct {
	AppType byte // 服务器类型 1=前端，2=后端
	Cluster bool // 是否开启集群服务
}

// 新建1个默认 Options
func NewOptions() *Options {
	o := Options{
		AppType: C_APP_TYPE_FRONTEND,
		Cluster: false,
	}

	return &o
}
