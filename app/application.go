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

	"github.com/zpab123/sco/cluster"
	"github.com/zpab123/sco/log"
	"github.com/zpab123/sco/network"
	"github.com/zpab123/sco/state"
	"github.com/zpab123/sco/svc"
)

// /////////////////////////////////////////////////////////////////////////////
// Application

// 1个通用服务器对象
type Application struct {
	Options      *Options              // 配置选项
	agentMgr     network.IAgentManager // agent 管理
	acceptors    []network.IAcceptor   // 接收器切片
	clientPacket chan *network.Packet  // 网络数据包
	serverPacket chan *network.Packet  // 服务器数据包
	postman      *cluster.Postman      // 消息转发
	svcs         []svc.IService        // 服务列表
	stopGroup    sync.WaitGroup        // 停止等待组
	signalChan   chan os.Signal        // 操作系统信号
	state        state.State           // 状态
	ctx          context.Context       // 退出 ctx
	cancel       context.CancelFunc    // 退出 ctx
	delegate     IDelegate             // 代理对象
}

// 创建1个新的 Application 对象
func NewApplication() *Application {

	// 创建对象
	acc := make([]network.IAcceptor, 0)
	ss := make([]svc.IService, 0)
	sch := make(chan os.Signal, 1)
	cx, cc := context.WithCancel(context.Background())
	pc := make(chan *network.Packet, 1000)

	// 创建 app
	a := Application{
		Options:      NewOptions(),
		acceptors:    acc,
		svcs:         ss,
		signalChan:   sch,
		ctx:          cx,
		cancel:       cc,
		clientPacket: pc,
	}

	return &a
}

// -----------------------------------------------------------------------------
// public

// 启动 app
func (this *Application) Run() {
	// 设置随机种子
	rand.Seed(time.Now().UnixNano())

	log.Logger.Debug("启动 app")

	// 启动网络
	this.runNet()

	// 集群服务
	if this.Options.Cluster {
		this.runCluster()
	}

	// 侦听信号
	this.listenSignal()

	// 代理
	if this.delegate != nil {
		this.delegate.Working()
	}

	// 主循环
	this.mainLoop()
}

// 停止app
func (this *Application) Stop() {
	if this.state.Get() == state.C_STOPING {
		return
	}
	this.state.Set(state.C_STOPING)

	log.Logger.Info(
		"[Application] 关闭中",
	)

	// 发出关闭信号
	//this.cancel()
	// go this.onStop()
	// this.stopGroup.Wait()
	for _, acc := range this.acceptors {
		acc.Stop()
	}

	if this.agentMgr != nil {
		this.agentMgr.Stop()
	}

	if this.postman != nil {
		this.postman.Stop()
	}

	if this.delegate != nil {
		this.delegate.Stop()
	}

	log.Logger.Info(
		"[Application] 优雅退出",
	)

	os.Exit(0)
}

// 添加 1个 tcp 接收器
//
// laddr=接收器地址，格式 192.168.1.222:6980
func (this *Application) AddTcpAcceptor(laddr string) error {
	acc, err := network.NewTcpAcceptor(laddr)
	if err != nil {
		return err
	}

	if this.agentMgr == nil {
		this.newAgentMgr()
	}
	acc.SetConnMgr(this.agentMgr)

	this.acceptors = append(this.acceptors, acc)

	return nil
}

// 添加 1个 weboscket 接收器
//
// laddr=接收器地址，格式 192.168.1.222:6980
func (this *Application) AddWsAcceptor(laddr string) error {
	acc, err := network.NewWsAcceptor(laddr)
	if err != nil {
		return err
	}

	if this.agentMgr == nil {
		this.newAgentMgr()
	}
	acc.SetConnMgr(this.agentMgr)

	this.acceptors = append(this.acceptors, acc)

	return nil
}

// 设置代理
func (this *Application) SetDelegate(d IDelegate) {
	if d != nil {
		this.delegate = d
		this.delegate.Init(this)
	}
}

// 注册服务
func (this *Application) RegService(s svc.IService) {
	if s != nil {
		s.Init()
		this.svcs = append(this.svcs, s)
	}
}

// 分发消息
func (this *Application) Post(pkt *network.Packet) {
	if this.postman != nil {
		this.postman.Post(pkt)
	}
}

// -----------------------------------------------------------------------------
// private

// 启动网络
func (this *Application) runNet() {
	// 连接管理
	if this.agentMgr != nil {
		this.agentMgr.Run()
	}

	// 接收器
	this.runAcceptor()
}

// 创建 agent 管理
func (this *Application) newAgentMgr() {
	this.agentMgr = network.NewAgentMgr(10000)
	this.agentMgr.SetPacketChan(this.clientPacket)
}

// 启动接收器
func (this *Application) runAcceptor() {
	for _, acc := range this.acceptors {
		// go acc.Run()
		acc.Run()
	}
}

// 启动集群
func (this *Application) runCluster() {
	if len(this.Options.Gates) <= 0 {
		return
	}

	this.newPktChan()
	this.newPostman()

	this.postman.Run()
}

// 服务器消息通道
func (this *Application) newPktChan() {
	this.serverPacket = make(chan *network.Packet, 1000)
}

// 转发
func (this *Application) newPostman() {
	this.postman = cluster.NewPostman(this.Options.Mid, this.Options.Gates)
	this.postman.SetPacketChan(this.serverPacket)
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
		case pkt := <-this.clientPacket: // 网络消息
			this.onClientPacket(pkt)
		case pkt := <-this.serverPacket: // 服务器数据包
			this.onServerPacket(pkt)
		case sig := <-this.signalChan: // os 信号
			this.onSignal(sig)
		}
	}
}

// 操作系统信号
func (this *Application) onSignal(sig os.Signal) {
	defer log.Logger.Sync()

	if syscall.SIGINT == sig || syscall.SIGTERM == sig {
		this.Stop()
	} else {
		log.Logger.Warn(
			"[Application] 异常的操作系统信号",
			log.String("sid=", sig.String()),
		)
	}
}

// app 准备关闭
func (this *Application) onStop() {
	time.Sleep(C_STOP_TIME_OUT)
	log.Logger.Warn(
		"[Application] 关闭超时，强制关闭",
		log.Uint16("超时时间(秒)=", uint16(C_STOP_TIME_OUT/time.Second)),
	)

	os.Exit(1)
}

// 接收到网络数据
func (this *Application) onClientPacket(pkt *network.Packet) {
	if this.delegate != nil {
		this.delegate.OnPacket(pkt)
	}
}

// 服务器数据包
func (this *Application) onServerPacket(pkt *network.Packet) {
	if this.delegate != nil {
		this.delegate.OnPacket(pkt)
	}
}
