// /////////////////////////////////////////////////////////////////////////////
// ectd 服务发现

package discovery

import (
	"context"
	"sync"

	"github.com/zpab123/zaplog" // log
	"go.etcd.io/etcd/clientv3"  // etcd 客户端
)

// /////////////////////////////////////////////////////////////////////////////
// etcdDiscovery 对象

// ectd 服务发现
type etcdDiscovery struct {
	client          *clientv3.Client                   // etcd 客户端
	leaseID         clientv3.LeaseID                   // 未知
	endpoints       []string                           // 注册中心地址集合
	serverMapByID   sync.Map                           // 服务器集群信息集合
	serverMapByType map[string]map[uint16]*ServiceDesc // 服务器集群信息集合
	rwMutex         sync.RWMutex                       // 读写锁
	serverDesc      *ServiceDesc                       // 自身服务器信息
}

// 新建1个 etcdDiscovery 对象
func NewEtcdDiscovery(endpoints []string) (IDiscovery, error) {
	ed := &etcdDiscovery{
		serverMapByType: make(map[string]map[uint16]*ServiceDesc),
	}

	return ed, nil
}

// 启动服务发现
func (this *etcdDiscovery) Run() {
	var err error
	if this.client == nil {
		conf := clientv3.Config{
			Endpoints:   this.endpoints,
			DialTimeout: C_DIAL_TIMEOUT,
		}

		this.client, err = clientv3.New(conf)
		if nil != err {
			return
		}
	}

	ch := this.client.Watch(context.Background(), C_ETCD_SERVER_DIR, clientv3.WithPrefix())
	go this.watchEtcdChanges(ch)
}

// 观察信息变化
func (this *etcdDiscovery) watchEtcdChanges(wChan clientv3.WatchChan) {
	for {
		select {
		case wRes := <-wChan:
			for _, evt := range wRes.Events {
				switch evt.Type {
				case clientv3.EventTypePut: //增加服务
					if sd, err := parseServer(evt.Kv.Value); err != nil {
						zaplog.Warnf("etcdDiscovery 发现新服务，但是解析服务发现json信息失败。err=%s", err.Error())

						continue
					}

					this.addServer(sd)
					zaplog.Debugf("etcdDiscovery 发现新服务，name=%s", evt.Kv.Key)
				case clientv3.EventTypeDelete: // 删除服务
				}
			}
		}
	}
}

// 增加1个服务信息
func (this *etcdDiscovery) addServer(sd *ServiceDesc) {
	if _, loaded := this.serverMapByID.LoadOrStore(sd.Mid, sd); !loaded {
		// 保存
		this.writeLockScope(func() {
			tMap, ok := this.serverMapByType[sd.Name]
			if !ok {
				tMap := make(map[uint16]*ServiceDesc)
				this.serverMapByType[sd.Name] = tMap
			}

			tMap[sd.Mid] = sd
		})

		// 通知
		if sd.Mid != this.serverDesc.Mid {
			this.notifyListeners(C_SERVICE_ADD, sd)
		}
	}
}

// 带锁写入数据
func (this *etcdDiscovery) writeLockScope(f func()) {
	this.rwMutex.Lock()
	defer this.rwMutex.Unlock()

	f()
}

// 通知
func (this *etcdDiscovery) notifyListeners(act int, sd *ServiceDesc) {

}
