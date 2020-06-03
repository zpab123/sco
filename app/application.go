// /////////////////////////////////////////////////////////////////////////////
// 1个通用服务器对象

package app

import (
	"context"
	"math/rand"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/zpab123/sco/discovery"
	"github.com/zpab123/sco/network"
	"github.com/zpab123/sco/rpc"
	"github.com/zpab123/zaplog"
)

// /////////////////////////////////////////////////////////////////////////////
// Application

// 1个通用服务器对象
type Application struct {
	Options        *Options                // 配置选项
	acceptors      []network.IAcceptor     // 接收器切片
	connMgr        network.IConnManager    // 连接管理
	clientAcceptor *network.ClientAcceptor // 客户端接收器
	discovery      discovery.IDiscovery    // 服务发现
	rpcServer      rpc.IServer             // rpc 服务端
	rpcClient      rpc.IClient             // rpc 客户端
	signalChan     chan os.Signal          // 操作系统信号
	stopGroup      sync.WaitGroup          // 停止等待组
	ctx            context.Context         // 上下文
	cancel         context.CancelFunc      // 退出通知函数
	remoteChan     chan *network.Packet    // remote 消息
	handleChan     chan *network.Packet    // 本地消息
	packetChan     chan *network.Packet    // 消息处理器
	localPacket    chan *network.Packet    // 本地消息
	remotePacket   chan *network.Packet    // 远程消息
	handler        network.IHandler        // 消息处理
	remoteService  rpc.IRemoteService      // remote 服务
}

// 创建1个新的 Application 对象
func NewApplication() *Application {
	// 创建对象
	sig := make(chan os.Signal, 1)
	opt := NewOptions()
	rc := make(chan *network.Packet, 1000)
	hc := make(chan *network.Packet, 1000)
	pc := make(chan *network.Packet, 1000)
	lp := make(chan *network.Packet, 1000)
	rp := make(chan *network.Packet, 1000)
	ctx, cancel := context.WithCancel(context.Background())

	// 创建 app
	a := Application{
		signalChan:   sig,
		Options:      opt,
		acceptors:    []network.IAcceptor{},
		remoteChan:   rc,
		handleChan:   hc,
		packetChan:   pc,
		localPacket:  lp,
		remotePacket: rp,
		ctx:          ctx,
		cancel:       cancel,
	}
	a.init()

	return &a
}

// 启动 app
func (this *Application) Run() {
	// 设置随机种子
	rand.Seed(time.Now().UnixNano())

	// 客户端接收器
	if C_APP_TYPE_FRONTEND == this.Options.AppType {
		if nil == this.connMgr {
			this.newConnMgr()
		}

		if len(this.acceptors) <= 0 {
			this.newClientAcceptor()
		}

		for _, acc := range this.acceptors {
			acc.SetConnMgr(this.connMgr)
			err := acc.Run()
			if nil == err {
				this.stopGroup.Add(1)
			}
		}
		this.stopGroup.Add(1)
		go this.clientAcceptor.Run()
	}

	// rpc服务
	if this.Options.Cluster {
		if nil == this.rpcServer {
			this.newRpcServer()
		}
		go this.rpcServer.Run(this.ctx)

		if nil == this.rpcClient {
			this.newRpcClient()
		}
		go this.rpcClient.Run(this.ctx)

		if nil == this.discovery {
			this.newDiscovery()
		}
		go this.discovery.Run(this.ctx)
	}

	// 消息分发
	go this.dispatch()

	zaplog.Infof("服务器，启动成功...")

	// 侦听结束信号
	this.waitStopSignal()
}

// 停止 app
func (this *Application) Stop() {
	zaplog.Infof("正在结束...")
	var err error

	// 客户端接收
	if this.clientAcceptor != nil {
		err = this.clientAcceptor.Stop()
		if nil == err {
			zaplog.Warnf("clientAcceptor 结束异常")
		}

		this.stopGroup.Done()
	}

	this.stopGroup.Wait()
	zaplog.Infof("服务器，优雅退出")
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

// 注册 handler
func (this *Application) RegisterHandler(handler network.IHandler) {
	if nil != handler {
		this.handler = handler
	}
}

// 设置 remote 服务
func (this *Application) SetRemoteService(rs rpc.IRemoteService) {
	if nil != rs {
		this.remoteService = rs
	}
}

// 初始化
func (this *Application) init() {
	// 默认设置
	this.defaultConfig()
	// 解析参数
	this.parseArgs()
}

// 侦听结束信号
func (this *Application) waitStopSignal() {
	// 排除信号
	signal.Ignore(syscall.Signal(10), syscall.Signal(12), syscall.SIGPIPE, syscall.SIGHUP)
	signal.Notify(this.signalChan, syscall.SIGINT, syscall.SIGTERM)

	for {
		sig := <-this.signalChan
		if syscall.SIGINT == sig || syscall.SIGTERM == sig {
			go this.Stop()
			time.Sleep(C_STOP_OUT_TIME)
			zaplog.Warnf("服务器，超过 %v 秒未优雅关闭，强制关闭", C_STOP_OUT_TIME)
			os.Exit(1)
		} else {
			zaplog.Warnf("异常的操作系统信号=%s", sig)
		}
	}
}

// 设置默认
func (this *Application) defaultConfig() {

}

// 解析命令行参数
func (this *Application) parseArgs() {

}

// 创建默认连接管理
func (this *Application) newConnMgr() {
	this.connMgr = network.NewConnMgr(this.Options.Net.MaxConn)
}

// 创建默认接收器
func (this *Application) newAcceptor() {
	this.newTcpAcceptor()
	this.newWsAcceptor()
}

// 创建 tcp 接收器
func (this *Application) newTcpAcceptor() {

}

// 创建 websocket 接收器
func (this *Application) newWsAcceptor() {

}

// 创建 clientAcceptor
func (this *Application) newClientAcceptor() {
	opt := network.NewTClientAcceptorOpt()
	opt.Handler = this
	opt.WsAddr = this.Options.Net.WsAddr

	this.clientAcceptor = network.NewClientAcceptor(opt)
}

// 创建 rpcserver
func (this *Application) newRpcServer() {
	opt := rpc.GrpcServerOptions{
		Laddr:         this.Options.RpcServer.Laddr,
		RemoteService: this.remoteService,
	}

	this.rpcServer = rpc.NewGrpcServer(&opt)
}

// 创建 rpcClient
func (this *Application) newRpcClient() {
	this.rpcClient = rpc.NewGrpcClient()
}

// 创建服务发现
func (this *Application) newDiscovery() {
	endpoints := []string{
		"http://192.168.1.69:2379",
		"http://192.168.1.69:2479",
		"http://192.168.1.69:2579",
	}

	this.discovery, _ = discovery.NewEtcdDiscovery(endpoints)

	// 服务描述
	desc := discovery.ServiceDesc{
		Name:  this.Options.Name,
		Mid:   this.Options.ServiceId,
		Laddr: this.Options.RpcServer.Laddr,
	}
	this.discovery.SetService(&desc)

	if nil != this.rpcClient {
		this.discovery.AddListener(this.rpcClient)
	}
}

// 收到1个 pakcet
func (this *Application) OnPacket(agent *network.Agent, pkt *network.Packet) {
	if pkt.GetMid() != this.Options.ServiceId {
		this.dispatchPacket(agent, pkt)
		return
	}

	// 本地
	if nil != this.handler {
		this.handler.OnPacket(agent, pkt)
	}
}

// 分发消息
func (this *Application) dispatchPacket(agent *network.Agent, pkt *network.Packet) {
	if nil != this.rpcClient {
		res := this.rpcClient.Call(pkt.GetMid(), pkt.GetBody())
		zaplog.Debugf("dispatchPacket=%v", res)
		if nil != res {
			agent.SendData(res)
		}
	}
}

// 分发消息
func (this *Application) dispatch() {
	for {
		select {
		case pkt := <-this.handleChan: // 本地消息
			this.handle(pkt)
		case pkt := <-this.remoteChan: // rpc 消息
			this.remote(pkt)
		case pkt := <-this.remotePacket: // 远程packet
			this.sendToServer(pkt)
		}
	}
}

// 处理本地消息
func (this *Application) handle(pkt *network.Packet) {

}

// 处理 rpc 消息
func (this *Application) remote(pkt *network.Packet) {

}

// 远程处理
func (this *Application) sendToServer(pkt *network.Packet) {
	if nil == this.rpcClient {
		return
	}

	//res := this.rpcClient

	// agent.send(res)
}
