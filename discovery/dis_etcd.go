// /////////////////////////////////////////////////////////////////////////////
// ectd 服务发现

package discovery

import (
	"go.etcd.io/etcd/clientv3"
)

// /////////////////////////////////////////////////////////////////////////////
// etcdDiscovery 对象

// ectd 服务发现
type etcdDiscovery struct {
	client  *clientv3.Client // etcd 客户端
	leaseID clientv3.LeaseID
}

// 新建1个 etcdDiscovery 对象
func NewEtcdDiscovery() (IDiscovery, error) {
	ed := &etcdDiscovery{}

	return ed, nil
}
