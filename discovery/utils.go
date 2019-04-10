// /////////////////////////////////////////////////////////////////////////////
// 公用工具

package discovery

import (
	"encoding/json"

	"github.com/pkg/errors"     // 异常
	"github.com/zpab123/zaplog" // log
)

// 解析服务器信息
func parseServer(v []byte) (*ServiceDesc, error) {
	var sd *ServiceDesc
	var err error

	err = json.Unmarshal(v, &sd)
	if err != nil {
		return nil, err
	}

	return sd, nil
}
