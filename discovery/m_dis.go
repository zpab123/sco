// /////////////////////////////////////////////////////////////////////////////
// 常量-接口-types

package discovery

import (
	"context"
	"time"
)

// /////////////////////////////////////////////////////////////////////////////
// 常量

// etcdDiscovery 常量
const (
	C_ED_SERVICE_DIR = "sco.service"     // 数据库目录
	C_ED_DT          = 2 * time.Second   // 连接注册中心超时时间
	C_ED_HEARTBEAT   = 60 * time.Second  // 租约时间
	C_ED_UI          = 120 * time.Second // 服务更新周期
	C_ED_RLT         = 60 * time.Second  // 服务更新周期
	C_ED_RLC         = 15                // 重新续约最大次数
	C_ED_RLI         = 5 * time.Second   // 重新续约间隔
	C_ED_RT          = 5 * time.Second   // 废除超时时间
)

//
const (
	C_SERVICE_ADD = iota // 服务添加
	C_SERVICE_DEL        // 服务删除
)

// /////////////////////////////////////////////////////////////////////////////
// 接口

// 服务发现接口
type IDiscovery interface {
	SetService(svcDesc *ServiceDesc) // 设置服务
	Run(ctx context.Context)         // 启动服务发现
	AddListener(listener IListener)  // 添加侦听者
}

// 服务发现事件侦听
type IListener interface {
	AddService(svcDec *ServiceDesc)    // 添加1个服务
	RemoveService(svcDec *ServiceDesc) // 移除1个服务
}

// /////////////////////////////////////////////////////////////////////////////
// etcdDiscovery

// etcdDiscovery 配置参数
type TEtcdDiscoveryOpt struct {
	Enable             bool          // 是否启用
	DialTimeout        time.Duration // 连接注册中心超时时间
	HeartbeatTTL       time.Duration // 租约时间
	UpdateInterval     time.Duration // 服务更新周期
	RenewLeaseTimeout  time.Duration // 重新续约超时时间
	RenewLeaseMacCount int           // 重新续约最大次数
	RenewLeaseInterval time.Duration // 重新续约间隔
	RevokeTimeout      time.Duration // 废除超时时间
}

// 新建1个 etcdDiscovery 对象
func NewTEtcdDiscoveryOpt() *TEtcdDiscoveryOpt {
	opt := TEtcdDiscoveryOpt{
		Enable:             true,
		DialTimeout:        C_ED_DT,
		HeartbeatTTL:       C_ED_HEARTBEAT,
		UpdateInterval:     C_ED_UI,
		RenewLeaseTimeout:  C_ED_RLT,
		RenewLeaseMacCount: C_ED_RLC,
		RenewLeaseInterval: C_ED_RLI,
		RevokeTimeout:      C_ED_RT,
	}

	return &opt
}
