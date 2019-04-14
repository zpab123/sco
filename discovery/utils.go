// /////////////////////////////////////////////////////////////////////////////
// 公用工具

package discovery

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/pkg/errors" // 异常
)

// 解析服务信息
func parseService(value []byte) (*ServiceDesc, error) {
	var svcDesc *ServiceDesc
	err := json.Unmarshal(value, svcDesc)
	if nil != err {
		return nil, err
	}

	return svcDesc, nil
}

// 解析服务 key 信息
func parseServiceKey(key string) (string, string, error) {
	strs := strings.Split(key, "/")

	if len(strs) != 3 {
		err := errors.Errorf("解析服务信息[%s]出错", key)
		return "", "", err
	}

	svcType := strs[1]
	name := strs[2]

	return svcType, name, nil
}

// 根据服务类型和名字获取服务器信息
func getKey(stype, name string) string {
	return fmt.Sprintf("%s/%s/%s", C_ETCD_SERVER_DIR, stype, name)
}
