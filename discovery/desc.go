// /////////////////////////////////////////////////////////////////////////////
// 服务发现需要的服务器信息描述

package discovery

import (
	"encoding/json"
	"fmt"

	"github.com/zpab123/sco/log"
)

// 注册到服务发现的服务描述
type ServiceDesc struct {
	Name  string `json:"name"`  // 服务器名字，不同的服务器，名字不能相同
	Mid   uint16 `json:"mid"`   // 服务主Id，不同的服务器，mid 可能相同
	Laddr string `json:"laddr"` // 监听地址
}

// 获取 key
func (this *ServiceDesc) Key() string {
	return fmt.Sprintf("%s/%d/%s", C_ED_SERVICE_DIR, this.Mid, this.Name)
}

// 按照 json 格式化后的字符串
func (this *ServiceDesc) JsonString() string {
	defer log.Logger.Sync()

	bytes, err := json.Marshal(this)
	if nil != err {
		log.Logger.Error(
			"[ServiceDesc] 服务发现描述信息，转化为json失败，返回空字符串",
			log.Uint16("mid=", this.Mid),
			log.String("name=", this.Name),
		)

		return ""
	}

	return string(bytes)
}
