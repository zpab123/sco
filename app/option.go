// /////////////////////////////////////////////////////////////////////////////
// app 可选参数

package app

// /////////////////////////////////////////////////////////////////////////////
// Option 对象

// app 配置参数
type Options struct {
	Id          string   // app 唯一标识
	Mid         uint16   // 服务ID
	Cluster     bool     // 是否开启集群服务
	Dispatchers []string // 分发器地址集合 ["192.168.0.1:66", "192.168.0.1:88", ...]
}

// 新建1个默认 Options
func NewOptions() *Options {
	ds := make([]string, 0)

	o := Options{
		Cluster:     false,
		Dispatchers: ds,
	}

	return &o
}
