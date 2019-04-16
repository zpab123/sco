// /////////////////////////////////////////////////////////////////////////////
// 服务发现需要的服务器信息描述

package discovery

import (
	"encoding/json"
	"fmt"

	"github.com/zpab123/zaplog" // log
)

// 注册到服务发现的服务描述
type ServiceDesc struct {
	Type string            `json:"type"` // 服务类型
	Name string            `json:"name"` // 服务名字
	Mid  uint16            `json:"mid"`  // 服务主Id，与类型对应
	Host string            `json:"host"` // 监听地址
	Port uint              `json:"port"` // 监听端口
	Tags []string          `json:"tags"` // 分类标签
	Meta map[string]string `json:"meta"` // 细节配置
}

// 获取监听地址，格式: 127.0.0.1:8080 形式
func (this *ServiceDesc) Address() string {
	return fmt.Sprintf("%s:%d", this.Host, this.Port)
}

// 获取 key
func (this *ServiceDesc) Key() string {
	return fmt.Sprintf("%s/%s/%s", C_ETCD_SERVER_DIR, this.Type, this.Name)
}

// 按照 json 格式化后的字符串
func (this *ServiceDesc) JsonString() string {
	bytes, err := json.Marshal(this)
	if nil != err {
		zaplog.Errorf("服务器[type=%s，name=%s]服务发现描述信息，转化为json失败，返回空字符串", this.Type, this.Name)

		return ""
	}

	return string(bytes)
}
