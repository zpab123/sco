// /////////////////////////////////////////////////////////////////////////////
// ectd 服务发现

package discovery

import (
	"context"
	"sync"
	"time"

	"github.com/pkg/errors"     // 异常库
	"github.com/zpab123/zaplog" // log
	"go.etcd.io/etcd/clientv3"  // etcd 客户端
)

var (
	ERROR_LEASE_TIMEOUT = errors.New("etcdDiscovery 重连服务器超时")
)

// /////////////////////////////////////////////////////////////////////////////
// etcdDiscovery 对象

// ectd 服务发现
type etcdDiscovery struct {
	name         string                             // 组件名字
	option       *TEtcdDiscoveryOpt                 // 配置参数
	client       *clientv3.Client                   // etcd 客户端
	endpoints    []string                           // 注册中心地址集合
	leaseID      clientv3.LeaseID                   // 租约id
	svcMapByName sync.Map                           // 服务器集群信息集合
	svcMapByType map[string]map[string]*ServiceDesc // 服务器集群信息集合
	rwMutex      sync.RWMutex                       // 读写锁
	serviceDesc  *ServiceDesc                       // 自身服务器信息
}

// 新建1个 etcdDiscovery 对象
func NewEtcdDiscovery(endpoints []string) (*etcdDiscovery, error) {
	var err error

	// 参数效验
	if len(endpoints) <= 0 {
		err = errors.New("创建 etcdDiscovery 失败。参数 endpoints 长度为0")

		return nil, err
	}

	// 基础对象
	opt := NewTEtcdDiscoveryOpt()

	ed := &etcdDiscovery{
		name:         C_ED_NAME,
		endpoints:    endpoints,
		option:       opt,
		svcMapByType: make(map[string]map[string]*ServiceDesc),
	}

	return ed, nil
}

// 启动服务发现
func (this *etcdDiscovery) Run(ctx context.Context) {
	var err error
	// 建立连接
	if this.client == nil {
		conf := clientv3.Config{
			Endpoints:   this.endpoints,
			DialTimeout: this.option.DialTimeout,
		}

		this.client, err = clientv3.New(conf)
		if nil != err {
			// err = errors.Errorf("etcdDiscovery 创建 clientv3 出错：%s", err.Error())
			return
		}
	}

	// 建立租约
	err = this.grantLease()
	if nil != err {
		return
	}

	// 同步服务
	err = this.syncService()
	if nil != err {
		return
	}

	// 服务更新
	go this.update()

	// 侦听etcd事件
	go this.watchEtcdChanges()

	return
}

// 停止服务发现
func (this *etcdDiscovery) Stop() {

}

// 获取组件名字
func (this *etcdDiscovery) Name() string {
	return this.name
}

// 获取 etcdDiscovery 可选参数
func (this *etcdDiscovery) Option() *TEtcdDiscoveryOpt {
	return this.option
}

// 设置 etcdDiscovery 自身服务信息
func (this *etcdDiscovery) SetService(svcDesc *ServiceDesc) {
	this.serviceDesc = svcDesc
}

// 从 etcd 更新所有服务信息
func (this *etcdDiscovery) UpdateService() error {
	keys, err := this.client.Get(context.TODO(), C_ETCD_SERVER_DIR, clientv3.WithPrefix(), clientv3.WithKeysOnly())
	if nil != err {
		return err
	}

	// 保存有效
	validName := make([]string, 0)
	for _, kv := range keys.Kvs {
		stype, name, err := parseServiceKey(string(kv.Key))
		if nil != err {
			// return err
		}

		validName = append(validName, name)

		if _, ok := this.svcMapByName.Load(name); !ok {
			svcDesc, err := this.getServiceFromEtcd(stype, name)
			if nil != err || nil == svcDesc {
				continue
			}

			this.addService(svcDesc)
		}
	}

	// 删除无效
	this.deleteInvalid(validName)

	return nil
}

// 建立租约
func (this *etcdDiscovery) grantLease() error {
	// 设置租约时间：如果这里没开通服务器的话，会一直卡住
	resp, err := this.client.Grant(context.TODO(), int64(this.option.HeartbeatTTL.Seconds()))
	if err != nil {
		return err
	}
	this.leaseID = resp.ID

	// 设置续租
	leaseChan, err := this.client.KeepAlive(context.TODO(), this.leaseID)
	if err != nil {
		return err
	}

	<-leaseChan
	go this.watchLeaseChan(leaseChan)

	return nil
}

// 监听租约通道
func (this *etcdDiscovery) watchLeaseChan(leaseChan <-chan *clientv3.LeaseKeepAliveResponse) {
	// 重新续约次数
	renewCount := 0

	for {
		select {
		case leaseResp := <-leaseChan:
			// 续约成功
			if leaseResp != nil {
				renewCount = 0
				continue
			}

			// 重新签约
			for {
				err := this.renewLease()

				if err != nil {
					renewCount += 1

					// 超时
					if err == ERROR_LEASE_TIMEOUT {
						return
					}

					// 达到最大重连次数
					if renewCount >= this.option.RenewLeaseMacCount {
						return
					}

					zaplog.Warnf("etcdDiscovery 签约失败，%d秒后重新签约", uint64(this.option.RenewLeaseInterval.Seconds()))

					time.Sleep(this.option.RenewLeaseInterval)
					continue
				}

				return
			}
		}
	}
}

// 重新续约
func (this *etcdDiscovery) renewLease() error {
	ch := make(chan error)

	// 异步连接
	go func() {
		defer close(ch)

		err := this.grantLease()
		if err != nil {
			ch <- err
			return
		}

		err = this.syncService()
		ch <- err
	}()

	select {
	case err := <-ch:
		return err
	case <-time.After(this.option.RenewLeaseTimeout):
		return ERROR_LEASE_TIMEOUT
	}
}

// 获取服务
func (this *etcdDiscovery) getServiceFromEtcd(stype, name string) (*ServiceDesc, error) {
	k := getKey(stype, name)
	v, err := this.client.Get(context.TODO(), k)
	if nil != err {
		return nil, err
	}

	if len(v.Kvs) == 0 {
		return nil, err
	}

	return parseService(v.Kvs[0].Value)
}

// 增加1个服务信息
func (this *etcdDiscovery) addService(sd *ServiceDesc) {
	if _, loaded := this.svcMapByName.LoadOrStore(sd.Name, sd); !loaded {
		// 保存
		this.writeLockScope(func() {
			nameMap, ok := this.svcMapByType[sd.Type]
			if !ok {
				nameMap = make(map[string]*ServiceDesc)
				this.svcMapByType[sd.Type] = nameMap
			}

			nameMap[sd.Name] = sd
		})

		// 通知
		if sd.Name != this.serviceDesc.Name {
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

// 删除无效
func (this *etcdDiscovery) deleteInvalid(validName []string) {
	rangeFunc := func(key interface{}, value interface{}) bool {
		name := key.(string)

		// 是否存在
		var have bool = false
		for _, v := range validName {
			if v == name {
				have = true
				break
			}
		}

		if !have {
			this.deleteService(name)
		}

		return true
	}

	this.svcMapByName.Range(rangeFunc)
}

// 删除某个服务
func (this *etcdDiscovery) deleteService(name string) {
	if v, ok := this.svcMapByName.Load(name); ok {
		svc := v.(*ServiceDesc)

		// 删除
		this.svcMapByName.Delete(name)

		// 删除
		this.writeLockScope(func() {
			if nMap, ok := this.svcMapByType[svc.Type]; ok {
				delete(nMap, svc.Name)
			}
		})

		// 通知
		this.notifyListeners(C_SERVICE_DEL, svc)
	}
}

// 同步服务
func (this *etcdDiscovery) syncService() error {
	// 推送自己服务
	if nil != this.serviceDesc {
		err := this.putService(this.serviceDesc)
		if nil != err {
			return err
		}
	}

	// 更新服务
	this.UpdateService()

	return nil
}

// 将服务加入 etcd
func (this *etcdDiscovery) putService(svcDesc *ServiceDesc) error {
	// 使用租约id建立临时 key-value 存储
	_, err := this.client.Put(context.TODO(), svcDesc.Key(), svcDesc.JsonString(), clientv3.WithLease(this.leaseID))

	return err
}

// 定时更新
func (this *etcdDiscovery) update() {
	ticker := time.NewTicker(this.option.UpdateInterval)
	for {
		select {
		case <-ticker.C:
			err := this.UpdateService()
			if nil != err {

			}
		}
	}
}

// 观察信息变化
func (this *etcdDiscovery) watchEtcdChanges() {
	wChan := this.client.Watch(context.Background(), C_ETCD_SERVER_DIR, clientv3.WithPrefix())

	for {
		select {
		case wRes := <-wChan:
			for _, evt := range wRes.Events {
				switch evt.Type {
				case clientv3.EventTypePut: //增加服务
					var sd *ServiceDesc
					var err error
					if sd, err = parseService(evt.Kv.Value); err != nil {
						zaplog.Warnf("etcdDiscovery 发现新服务，但是解析服务发现json信息失败。err=%s", err.Error())

						continue
					}

					this.addService(sd)
					zaplog.Debugf("etcdDiscovery 发现新服务，name=%s", evt.Kv.Key)
				case clientv3.EventTypeDelete: // 删除服务
					_, name, err := parseServiceKey(string(evt.Kv.Key))
					if nil != err {
						continue
					}

					this.deleteService(name)
				}
			}
		}
	}
}

// 废除服务
func (this *etcdDiscovery) revoke() error {
	c := make(chan error)
	defer close(c)

	go func() {
		_, err := this.client.Revoke(context.TODO(), this.leaseID)
		c <- err
	}()

	select {
	case err := <-c:
		return err
	case <-time.After(this.option.RevokeTimeout):
		return nil
	}
}
