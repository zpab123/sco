// /////////////////////////////////////////////////////////////////////////////
// 1个通用服务器对象

package app

import (
	"math/rand"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/zpab123/sco/discovery"
	"github.com/zpab123/sco/log"
	"github.com/zpab123/sco/network"
	"github.com/zpab123/sco/rpc"
)

// /////////////////////////////////////////////////////////////////////////////
// Application

// 1个通用服务器对象
type Application struct {
	Options           *Options             // 配置选项
	acceptors         []network.IAcceptor  // 接收器切片
	connMgr           network.IConnManager // 连接管理
	handler           network.IHandler     // handler 服务
	rpcServer         rpc.IServer          // rpc 服务端
	rpcClient         rpc.IClient          // rpc 客户端
	discovery         discovery.IDiscovery // 服务发现
	remote            IRemote              // remote 服务
	signalChan        chan os.Signal       // 操作系统信号
	stopGroup         sync.WaitGroup       // 停止等待组
	clientPacet       chan *network.Packet // 来自客户端的消息
	serverPacket      chan *network.Packet // 来自服务器的消息
	remotePacket      chan *network.Packet // 远端消息
	acceptorForServer network.IAcceptor    // 用于服务器之间的相互连接
	serverConnMgr     network.IConnManager // 用于服务器连接管理
}

// 创建1个新的 Application 对象
func NewApplication(opts ...*Options) *Application {
	var opt *Options
	if len(opts) <= 0 {
		opt = NewOptions()
	} else {
		opt = opts[0]
	}

	// 创建对象
	sig := make(chan os.Signal, 1)
	cp := make(chan *network.Packet, 1000)
	rp := make(chan *network.Packet, 1000)
	sp := make(chan *network.Packet, 1000)

	// 创建 app
	a := Application{
		signalChan:   sig,
		Options:      opt,
		acceptors:    []network.IAcceptor{},
		clientPacet:  cp,
		serverPacket: sp,
		remotePacket: rp,
	}
	a.init()

	return &a
}

// 启动 app
func (this *Application) Run() {
	defer log.Logger.Sync()

	// 设置随机种子
	rand.Seed(time.Now().UnixNano())

	// 前端
	if C_APP_TYPE_FRONTEND == this.Options.AppType {
		this.runFrontend()
	}

	// 集群服务
	if this.Options.Cluster {
		this.runCluster()
	}

	log.Logger.Info(
		"[Application] 启动成功",
		log.String("id=", this.Options.Id),
	)

	// 侦听信号
	this.listenSignal()

	// 主循环
	this.mainLoop()
}

// 停止 app
func (this *Application) Stop() {
	defer log.Logger.Sync()

	log.Logger.Info(
		"[Application] 正在结束",
		log.String("id=", this.Options.Id),
	)

	// 停止前端
	this.stopFrontend()

	// 停止集群
	this.stopCluster()

	this.stopGroup.Wait()

	log.Logger.Info(
		"[Application] 优雅退出",
		log.String("id=", this.Options.Id),
	)

	os.Exit(0)
}

// 设置连接管理
func (this *Application) SetConnMgr(mgr network.IConnManager) {
	if nil != mgr {
		this.connMgr = mgr
	}
}

// 添加接收器
func (this *Application) AddAcceptor(acc network.IAcceptor) {
	if nil != acc {
		this.acceptors = append(this.acceptors, acc)
	}
}

// 设置 handler
func (this *Application) SetHandler(h network.IHandler) {
	if nil != h {
		this.handler = h
	}
}

// 设置 remote 服务
func (this *Application) SetRemote(r IRemote) {
	if nil != r {
		this.remote = r
	}
}

// 以 rpc 的方式获取远程数据
func (this *Application) Call(mid uint16, data []byte) []byte {
	if nil == this.rpcClient {
		return nil
	}

	return this.rpcClient.RemoteCall(mid, data)
}

// -----------------------------------------------------------------------------
// Agent -> IHandler 接口

// 收到1个 pakcet
func (this *Application) OnPacket(pkt *network.Packet) {
	// 集群
	if this.Options.Cluster {
		if pkt.GetMid() != this.Options.Mid {
			this.remotePacket <- pkt
			return
		}
	}

	this.clientPacet <- pkt
}

// -----------------------------------------------------------------------------
//

// 收到1个远端 pakcet
func (this *Application) onRemotePacket(a *network.Agent, pkt *network.Packet) {
	if nil == this.rpcClient {
		return
	}

	r, data := this.rpcClient.HandlerCall(pkt.GetMid(), pkt.GetBody())
	if nil != data {
		a.SendBytes(data)
	}

	if !r {
		a.Stop()
	}
}

// -----------------------------------------------------------------------------
// rpc -> IService 接口

// 收到 Handler 请求
func (this *Application) OnHandlerCall(data []byte) (bool, []byte) {
	if nil == this.handler {
		return true, nil
	}

	// return this.handler.OnData(data)
	return true, nil
}

// 收到 Remote 请求
func (this *Application) OnRemoteCall(data []byte) []byte {
	if nil == data {
		return nil
	}

	if nil == this.remote {
		return nil
	}

	return this.remote.OnData(data)
}

// -----------------------------------------------------------------------------
//

// 初始化
func (this *Application) init() {

}

// 侦听信号
func (this *Application) listenSignal() {
	// 排除信号
	signal.Ignore(syscall.Signal(10), syscall.Signal(12), syscall.SIGPIPE, syscall.SIGHUP)
	signal.Notify(this.signalChan, syscall.SIGINT, syscall.SIGTERM)
}

// 主循环
func (this *Application) mainLoop() {
	for {
		select {
		case cp := <-this.clientPacet:
			this.onClientPacket(cp)
		case rp := <-this.remotePacket:
			this.doRemotePacket(rp)
		case sp := <-this.serverPacket:
			this.onServerPacket(sp)
		case sig := <-this.signalChan:
			this.onSignal(sig)
		}
	}
}

// 启动前端
func (this *Application) runFrontend() {
	defer log.Logger.Sync()

	if nil == this.connMgr {
		this.newConnMgr()
	}

	if len(this.acceptors) <= 0 {
		this.newAcceptor()
	}

	if len(this.acceptors) <= 0 {
		log.Logger.Warn(
			"[Application] 为前端app，但无接收器",
			log.String("id=", this.Options.Id),
		)

		return
	}

	for _, acc := range this.acceptors {
		go acc.Run()
	}
}

// 停止前端
func (this *Application) stopFrontend() {
	// 停止接收器
	if len(this.acceptors) > 0 {
		for _, acc := range this.acceptors {
			acc.Stop()
		}
	}

	// 停止连接管理
	if nil != this.connMgr {
		this.connMgr.Stop()
	}
}

// 创建默认连接管理
func (this *Application) newConnMgr() {
	this.connMgr = network.NewConnMgr(this.Options.Frontend.MaxConn)
	this.connMgr.SetKey(this.Options.Frontend.Key)
	this.connMgr.SetHeartbeat(this.Options.Frontend.Heartbeat)
	this.connMgr.SetPacketChan(this.clientPacet)
}

// 创建默认接收器
func (this *Application) newAcceptor() {
	this.newTcpAcceptor()
	this.newWsAcceptor()
}

// 创建 tcp 接收器
func (this *Application) newTcpAcceptor() {
	if "" == this.Options.Frontend.TcpAddr {
		return
	}

	a, err := network.NewTcpAcceptor(this.Options.Frontend.TcpAddr)
	if nil != err {
		return
	}
	a.SetConnMgr(this.connMgr)
	this.acceptors = append(this.acceptors, a)
}

// 创建 websocket 接收器
func (this *Application) newWsAcceptor() {
	if "" == this.Options.Frontend.WsAddr {
		return
	}

	a, err := network.NewWsAcceptor(this.Options.Frontend.WsAddr)
	if nil != err {
		return
	}
	a.SetConnMgr(this.connMgr)
	this.acceptors = append(this.acceptors, a)
}

// 启动集群
func (this *Application) runCluster() {
	/*
		if nil == this.rpcServer {
			this.newRpcServer()
		}
		if nil != this.rpcServer {
			go this.rpcServer.Run()
		}

		if nil == this.rpcClient {
			this.newRpcClient()
		}
		if nil != this.rpcClient {
			go this.rpcClient.Run()
		}

		if nil == this.discovery {
			this.newDiscovery()
		}
		if nil != this.discovery {
			go this.discovery.Run()
		}
	*/
	if this.acceptorForServer == nil {
		this.newAcceptorForServer()
	}

	if this.acceptorForServer != nil {
		go this.acceptorForServer.Run()
	}
}

// 停止集群
func (this *Application) stopCluster() {
	if nil != this.rpcServer {
		this.rpcServer.Stop()
	}

	if nil != this.rpcClient {
		this.rpcClient.Stop()
	}

	if nil != this.discovery {
		this.discovery.Stop()
	}
}

// 创建 rpcserver
func (this *Application) newRpcServer() {
	defer log.Logger.Sync()

	s, err := rpc.NewGrpcServer(this.Options.RpcServer.Laddr)
	if nil != err {
		log.Logger.Warn(
			"[Application] 创建 GrpcServer 失败",
			log.String("err=", err.Error()),
		)

		return
	}

	s.SetService(this)
	this.rpcServer = s

}

func (this *Application) newAcceptorForServer() {
	this.serverConnMgr = network.NewConnMgr(this.Options.Frontend.MaxConn)
	this.serverConnMgr.SetKey(this.Options.Frontend.Key)
	this.serverConnMgr.SetHeartbeat(this.Options.Frontend.Heartbeat)
	this.serverConnMgr.SetPacketChan(this.serverPacket)
	if "" == this.Options.RpcServer.Laddr {
		return
	}

	a, err := network.NewTcpAcceptor(this.Options.RpcServer.Laddr)
	if nil != err {
		return
	}
	a.SetConnMgr(this.serverConnMgr)
	this.acceptorForServer = a
}

// 创建 rpcClient
func (this *Application) newRpcClient() {
	this.rpcClient = rpc.NewGrpcClient()
}

// 创建服务发现
func (this *Application) newDiscovery() {
	defer log.Logger.Sync()

	e := this.Options.Discovery.Endpoints
	if len(e) <= 0 {
		return
	}

	o := this.Options.Discovery.Etcd
	d, err := discovery.NewEtcdDiscovery(e, o)
	if nil != err {
		log.Logger.Warn(
			"[Application] 创建 EtcdDiscovery 失败",
			log.String("err=", err.Error()),
		)

		return
	}
	this.discovery = d

	// 服务描述
	desc := discovery.ServiceDesc{
		Name:  this.Options.Id,
		Mid:   this.Options.Mid,
		Laddr: this.Options.RpcServer.Laddr,
	}
	this.discovery.SetService(&desc)

	if nil != this.rpcClient {
		this.discovery.AddListener(this.rpcClient)
	}
}

//  本地消息
func (this *Application) onClientPacket(pkt *network.Packet) {
	this.handler.OnPacket(pkt)
}

//  远端消息
func (this *Application) doRemotePacket(pkt *network.Packet) {

}

// 服务器消息
func (this *Application) onServerPacket(pkt *network.Packet) {
	// 来自客户的 onClientPacket()

	// 来自其他服务器的
}

// 操作系统信号
func (this *Application) onSignal(sig os.Signal) {
	defer log.Logger.Sync()

	if syscall.SIGINT == sig || syscall.SIGTERM == sig {
		go this.Stop()
		time.Sleep(C_STOP_OUT_TIME)
		log.Logger.Warn(
			"[Application] 关闭超时，强制关闭",
			log.Uint16("超时时间(秒)=", uint16(C_STOP_OUT_TIME/time.Second)),
		)

		os.Exit(1)
	} else {
		log.Logger.Warn(
			"[Application] 异常的操作系统信号",
			log.String("sid=", sig.String()),
		)
	}
}
