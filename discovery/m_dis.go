// /////////////////////////////////////////////////////////////////////////////
// 常量-接口-types

package discovery

import (
	"time"

	"github.com/pkg/errors"        // 异常库
	"github.com/zpab123/sco/model" // 全局模型
)

// /////////////////////////////////////////////////////////////////////////////
// 常量

// etcd 常量
const (
	C_DIAL_TIMEOUT    = 2 * time.Second // 连接注册中心超时时间
	C_ETCD_SERVER_DIR = "sco.servers/"  // etcd 数据库目录
)

//
const (
	C_SERVICE_ADD = iota // 服务添加
	C_SERVICE_DEL        // 服务删除
)

// 错误
const (
	C_ERROR_LEASE_TIMEOUT = errors.New("etcdDiscovery 重连服务器超时")
)

// /////////////////////////////////////////////////////////////////////////////
// 接口

// 服务发现接口
type IDiscovery interface {
	model.IComponent // 接口继承
}

// /////////////////////////////////////////////////////////////////////////////
// 接口

// etcdDiscovery 配置参数
type TEtcdDiscoveryOpt struct {
}
