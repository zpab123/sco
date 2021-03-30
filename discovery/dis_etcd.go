// /////////////////////////////////////////////////////////////////////////////
// ectd 服务发现

package discovery

import (
	"context"
	"sync"
	"time"

	"github.com/pkg/errors"      // 异常库
	"github.com/zpab123/sco/log" // log
	"go.etcd.io/etcd/clientv3"   // etcd 客户端
)

var (
	LeaseTimeout = errors.New("EtcdDiscovery 重连服务器超时")
)

// /////////////////////////////////////////////////////////////////////////////
// EtcdDiscovery 对象

// ectd 服务发现
type EtcdDiscovery struct {
	options      *EtcdDiscoveryOpt                  // 配置参数
	client       *clientv3.Client                   // etcd 客户端
	endpoints    []string                           // 注册中心地址集合
	leaseID      clientv3.LeaseID                   // 租约id
	svcMapByName sync.Map                           // 服务器集群信息集合
	svcMapByMid  map[uint16]map[string]*ServiceDesc // 服务器集群信息集合
	rwMutex      sync.RWMutex                       // 读写锁
	serviceDesc  *ServiceDesc                       // 自身服务器信息
	listeners    []IListener                        // 服务发现事件侦听对象
}

// 新建1个 EtcdDiscovery 对象

// 成功，返回 *EtcdDiscovery 和 nil
// 失败，返回 nil 和 error
func NewEtcdDiscovery(endpoints []string, opts ...*EtcdDiscoveryOpt) (IDiscovery, error) {
	// 参数效验
	if len(endpoints) <= 0 {
		err := errors.New("创建 EtcdDiscovery 失败。参数 endpoints 为空")

		return nil, err
	}

	// 选项
	var opt *EtcdDiscoveryOpt
	if len(opts) <= 0 {
		opt = NewEtcdDiscoveryOpt()
	} else {
		opt = opts[0]
	}

	ed := EtcdDiscovery{
		endpoints:   endpoints,
		options:     opt,
		svcMapByMid: make(map[uint16]map[string]*ServiceDesc),
		listeners:   make([]IListener, 0),
	}

	return &ed, nil
}

// 启动服务发现
func (this *EtcdDiscovery) Run() error {
	defer log.Logger.Sync()

	var err error
	// 建立连接
	if this.client == nil {
		conf := clientv3.Config{
			Endpoints:   this.endpoints,
			DialTimeout: this.options.DialTimeout,
		}

		c, err := clientv3.New(conf)
		if nil != err {
			return err
		} else {
			this.client = c
		}
	}

	// 建立租约
	err = this.grantLease()
	if nil != err {
		return err
	}

	// 同步服务
	err = this.syncService()
	if nil != err {
		return err
	}

	// 服务更新
	go this.update()

	// 侦听etcd事件
	go this.watchEtcdChanges()

	log.Logger.Info(
		"[EtcdDiscovery] 启动成功",
	)

	return nil
}

// 停止服务发现
func (this *EtcdDiscovery) Stop() error {
	return this.revoke()
}

// 获取 EtcdDiscovery 可选参数
func (this *EtcdDiscovery) Options() *EtcdDiscoveryOpt {
	return this.options
}

// 设置 EtcdDiscovery 自身服务信息
func (this *EtcdDiscovery) SetService(svcDesc *ServiceDesc) {
	this.serviceDesc = svcDesc
}

// 从 etcd 更新所有服务信息
func (this *EtcdDiscovery) UpdateService() error {
	keys, err := this.client.Get(context.TODO(), C_ED_SERVICE_DIR, clientv3.WithPrefix(), clientv3.WithKeysOnly())
	if nil != err {
		return err
	}

	// 保存有效
	validName := make([]string, 0)
	for _, kv := range keys.Kvs {
		mid, name, err := parseServiceKey(string(kv.Key))
		if nil != err {
			// return err
			continue
		}

		validName = append(validName, name)

		if _, ok := this.svcMapByName.Load(name); !ok {
			svcDesc, err := this.getServiceFromEtcd(mid, name)
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

// 添加1个服务发现侦听对象
func (this *EtcdDiscovery) AddListener(ln IListener) {
	if nil != ln {
		this.listeners = append(this.listeners, ln)
	}
}

// 建立租约
func (this *EtcdDiscovery) grantLease() error {
	// 设置租约时间：如果这里没开通服务器的话，会一直卡住
	resp, err := this.client.Grant(context.TODO(), int64(this.options.HeartbeatTTL.Seconds()))
	if err != nil {
		return err
	}
	this.leaseID = resp.ID

	// 设置续租
	leaseChan, err := this.client.KeepAlive(context.TODO(), this.leaseID)
	if nil != err {
		return err
	}

	<-leaseChan
	go this.watchLeaseChan(leaseChan)

	return nil
}

// 监听租约通道
func (this *EtcdDiscovery) watchLeaseChan(leaseChan <-chan *clientv3.LeaseKeepAliveResponse) {
	defer log.Logger.Sync()

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
					if err == LeaseTimeout {
						return
					}

					// 达到最大重连次数
					if renewCount >= this.options.RenewLeaseMaxCount {
						return
					}

					log.Logger.Warn(
						"[EtcdDiscovery] 签约失败",
						log.Uint64("下次签约时间(秒):", uint64(this.options.RenewLeaseInterval.Seconds())),
					)

					time.Sleep(this.options.RenewLeaseInterval)
					continue
				}

				return
			}
		}
	}
}

// 重新续约
func (this *EtcdDiscovery) renewLease() error {
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
	case <-time.After(this.options.RenewLeaseTimeout):
		return LeaseTimeout
	}
}

// 获取服务
func (this *EtcdDiscovery) getServiceFromEtcd(mid, name string) (*ServiceDesc, error) {
	k := getKey(mid, name)
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
func (this *EtcdDiscovery) addService(sd *ServiceDesc) {
	if _, loaded := this.svcMapByName.LoadOrStore(sd.Name, sd); !loaded {
		// 保存
		this.writeLockScope(func() {
			nMap, ok := this.svcMapByMid[sd.Mid]
			if !ok {
				nMap = make(map[string]*ServiceDesc)
				this.svcMapByMid[sd.Mid] = nMap
			}

			nMap[sd.Name] = sd
		})

		// 通知
		if sd.Name != this.serviceDesc.Name {
			this.notifyListeners(C_SERVICE_ADD, sd)
		}
	}
}

// 带锁写入数据
func (this *EtcdDiscovery) writeLockScope(f func()) {
	this.rwMutex.Lock()
	defer this.rwMutex.Unlock()

	f()
}

// 通知
func (this *EtcdDiscovery) notifyListeners(act int, sd *ServiceDesc) {
	log.Logger.Debug(
		"[EtcdDiscovery] 服务发生变化",
		log.String("name=", sd.Name),
		log.Int64("act=", int64(act)),
	)

	if C_SERVICE_ADD == act {
		for _, ln := range this.listeners {
			ln.AddService(sd)
		}
	} else if C_SERVICE_DEL == act {
		for _, ln := range this.listeners {
			ln.RemoveService(sd)
		}
	}
}

// 删除无效
func (this *EtcdDiscovery) deleteInvalid(validName []string) {
	rangeFunc := func(key interface{}, value interface{}) bool {
		name := key.(string)

		// 是否存在
		var have bool = false
		for _, v := range validName {
			if v == name {
				have = true
				break
			} else {
				have = false
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
func (this *EtcdDiscovery) deleteService(name string) {
	if v, ok := this.svcMapByName.Load(name); ok {
		// 删除
		this.svcMapByName.Delete(name)

		svc, r := v.(*ServiceDesc)
		if !r {
			return
		}

		// 删除
		this.writeLockScope(func() {
			if nMap, ok := this.svcMapByMid[svc.Mid]; ok {
				delete(nMap, svc.Name)
			}
		})

		// 通知
		this.notifyListeners(C_SERVICE_DEL, svc)
	}
}

// 同步服务
func (this *EtcdDiscovery) syncService() error {
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
func (this *EtcdDiscovery) putService(svcDesc *ServiceDesc) error {
	// 使用租约id建立临时 key-value 存储
	_, err := this.client.Put(context.TODO(), svcDesc.Key(), svcDesc.JsonString(), clientv3.WithLease(this.leaseID))

	return err
}

// 定时更新
func (this *EtcdDiscovery) update() {
	ticker := time.NewTicker(this.options.UpdateInterval)
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
func (this *EtcdDiscovery) watchEtcdChanges() {
	wChan := this.client.Watch(context.Background(), C_ED_SERVICE_DIR, clientv3.WithPrefix())

	for {
		select {
		case wRes := <-wChan:
			for _, evt := range wRes.Events {
				switch evt.Type {
				case clientv3.EventTypePut: //增加服务
					var sd *ServiceDesc
					var err error
					if sd, err = parseService(evt.Kv.Value); err != nil {
						log.Logger.Warn(
							"[EtcdDiscovery] 发现新服务，但是解析服务发现json信息失败",
							log.String("err=", err.Error()),
						)
						log.Logger.Sync()

						continue
					}

					this.addService(sd)

					log.Logger.Debug(
						"[EtcdDiscovery] 发现新服务",
						log.String("name=", sd.Name),
						log.Uint16("mid=", sd.Mid),
					)
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
func (this *EtcdDiscovery) revoke() error {
	c := make(chan error)
	defer close(c)

	go func() {
		_, err := this.client.Revoke(context.TODO(), this.leaseID)
		c <- err
	}()

	select {
	case err := <-c:
		return err
	case <-time.After(this.options.RevokeTimeout):
		return nil
	}
}

// /////////////////////////////////////////////////////////////////////////////
// EtcdDiscovery

// EtcdDiscovery 配置参数
type EtcdDiscoveryOpt struct {
	DialTimeout        time.Duration // 连接注册中心超时时间
	HeartbeatTTL       time.Duration // 租约时间
	UpdateInterval     time.Duration // 服务更新周期
	RenewLeaseTimeout  time.Duration // 重新续约超时时间
	RenewLeaseMaxCount int           // 重新续约最大次数
	RenewLeaseInterval time.Duration // 重新续约间隔
	RevokeTimeout      time.Duration // 废除超时时间
}

// 新建1个 EtcdDiscovery 对象
func NewEtcdDiscoveryOpt() *EtcdDiscoveryOpt {
	opt := EtcdDiscoveryOpt{
		DialTimeout:        C_ED_DT,
		HeartbeatTTL:       C_ED_HEARTBEAT,
		UpdateInterval:     C_ED_UI,
		RenewLeaseTimeout:  C_ED_RLT,
		RenewLeaseMaxCount: C_ED_RLC,
		RenewLeaseInterval: C_ED_RLI,
		RevokeTimeout:      C_ED_RT,
	}

	return &opt
}
