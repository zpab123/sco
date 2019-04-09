// /////////////////////////////////////////////////////////////////////////////
// 服务发现需要的服务器信息描述

package discovery

import (
	"fmt"
)

// 注册到服务发现的服务描述
type ServiceDesc struct {
	Name string            // 服务名字
	Mid  string            // 服务主Id，整个集群唯一
	Host string            // 监听地址
	Port int               // 监听端口
	Tags []string          // 分类标签
	Meta map[string]string // 细节配置
}

// 获取监听地址，格式: 127.0.0.1:8080 形式
func (this *ServiceDesc) Address() string {
	return fmt.Sprintf("%s:%d", self.Host, self.Port)
}
