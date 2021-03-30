// /////////////////////////////////////////////////////////////////////////////
// grpc 客户端

package rpc

import (
	"context"
	"math/rand"
	"sync"
	"time"

	"github.com/zpab123/sco/discovery"
	"github.com/zpab123/sco/log"
	"github.com/zpab123/sco/protocol"
)

// /////////////////////////////////////////////////////////////////////////////
// GrpcClient

// grpc 客户端
type GrpcClient struct {
	connMapByMid map[uint16]map[string]*GrpcConn // 服务器集群信息集合
	reqTimeout   time.Duration                   // 请求超时
	dialTimeout  time.Duration                   // 连接超时
	mutex        sync.Mutex                      // 锁
}

// 新建1个 GrpcClient
func NewGrpcClient() IClient {
	gc := GrpcClient{
		connMapByMid: make(map[uint16]map[string]*GrpcConn),
		reqTimeout:   5 * time.Second,
		dialTimeout:  5 * time.Second,
	}

	return &gc
}

// 启动
func (this *GrpcClient) Run() error {
	defer log.Logger.Sync()

	log.Logger.Info(
		"[GrpcClient] 启动成功",
	)

	return nil
}

// 停止
func (this *GrpcClient) Stop() error {
	defer log.Logger.Sync()

	for _, nMap := range this.connMapByMid {
		for _, gc := range nMap {
			gc.close()
		}
	}

	log.Logger.Info(
		"[GrpcClient] 停止成功",
	)

	return nil
}

// Handler 调用
func (this *GrpcClient) HandlerCall(mid uint16, data []byte) (bool, []byte) {
	req := protocol.HandlerReq{
		Data: data,
	}

	gc := this.randConn(mid)
	if nil == gc {
		return true, nil
	}

	res, err := gc.HandlerCall(context.Background(), &req)
	if nil != err {
		log.Logger.Debug(
			"[GrpcClient] HandlerCall 错误",
			log.String("err=", err.Error()),
		)

		return true, nil
	}

	return res.Right, res.Data
}

// 远程调用
func (this *GrpcClient) RemoteCall(mid uint16, data []byte) []byte {
	if nil == data {
		return nil
	}

	req := protocol.RemoteReq{
		Data: data,
	}

	gc := this.randConn(mid)
	if nil == gc {
		return nil
	}

	res, err := gc.RemoteCall(context.Background(), &req)
	if nil != err {
		log.Logger.Debug(
			"[GrpcClient] RemoteCall 错误",
			log.String("err=", err.Error()),
		)

		return nil
	}

	return res.Data
}

// 添加 rpc 服务信息
func (this *GrpcClient) AddService(desc *discovery.ServiceDesc) {
	if nil == desc {
		return
	}

	gc := NewGrpcConn(desc.Laddr)

	/*
		err := gc.connect()
		if nil != err {
			zaplog.Debugf("[GrpcClient] 连接 rpcServer=%s 失败。err=%s", desc.Laddr, err.Error())
			return
		}
	*/

	this.mutex.Lock()
	nMap, ok := this.connMapByMid[desc.Mid]
	if !ok {
		nMap = make(map[string]*GrpcConn)
		this.connMapByMid[desc.Mid] = nMap
	}
	nMap[desc.Name] = gc
	this.mutex.Unlock()
}

// 移除 rpc 服务信息
func (this *GrpcClient) RemoveService(desc *discovery.ServiceDesc) {
	if nil == desc {
		return
	}

	this.mutex.Lock()
	defer this.mutex.Unlock()

	nMap, ok := this.connMapByMid[desc.Mid]
	if !ok {
		return
	}
	gc, ok := nMap[desc.Name]
	if !ok {
		return
	}
	gc.close()
	delete(nMap, desc.Name)
}

// 根据 mid 随机1个 GrpcConn
func (this *GrpcClient) randConn(mid uint16) *GrpcConn {
	this.mutex.Lock()
	defer this.mutex.Unlock()

	nMap, ok := this.connMapByMid[mid]
	if !ok {
		return nil
	}

	lis := make([]*GrpcConn, 0)
	for _, gc := range nMap {
		lis = append(lis, gc)
	}

	n := len(lis)
	if n <= 0 {
		return nil
	}

	return lis[rand.Intn(n)]
}
