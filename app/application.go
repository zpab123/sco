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

	"github.com/zpab123/sco/log"
	"github.com/zpab123/sco/network"
	"github.com/zpab123/sco/state"
	"github.com/zpab123/sco/svc"
)

// /////////////////////////////////////////////////////////////////////////////
// Application

// 1个通用服务器对象
type Application struct {
	Options      *Options                 // 配置选项
	agentMgr     network.IAgentManager    // agent 管理
	acceptors    []network.IAcceptor      // 接收器切片
	agentEvt     chan *network.AgentEvent // agent 事件
	clientPacket chan *network.Packet     // 网络数据包
	serverPacket chan *network.Packet     // 服务器数据包
	stcPkt       chan *network.Packet     // server -> client
	postman      *network.Postman         // 消息转发
	svcs         []svc.IService           // 服务列表
	stopGroup    sync.WaitGroup           // 停止等待组
	signalChan   chan os.Signal           // 操作系统信号
	state        state.State              // 状态
	ctx          context.Context          // 退出 ctx
	cancel       context.CancelFunc       // 退出 ctx
	delegate     IDelegate                // 代理对象
}

// 创建1个新的 Application 对象
func NewApplication() *Application {

	// 创建对象
	acc := make([]network.IAcceptor, 0)
	ss := make([]svc.IService, 0)
	sch := make(chan os.Signal, 1)
	cx, cc := context.WithCancel(context.Background())

	// 创建 app
	a := Application{
		Options:    NewOptions(),
		agentMgr:   network.NewAgentMgr(10000),
		acceptors:  acc,
		svcs:       ss,
		signalChan: sch,
		ctx:        cx,
		cancel:     cc,
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

	// 集群服务
	if this.Options.Cluster {
		this.runCluster()
	}

	// 启动网络
	this.runNet()

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

	acc.SetConnMgr(this.agentMgr)

	this.acceptors = append(this.acceptors, acc)

	return nil
}

// 设置 AgentEvent 消息通道
func (this *Application) SetAgentEventChan(ch chan *network.AgentEvent) {
	if ch != nil {
		this.agentEvt = ch
	}
}

// 设置客户端消息通道
func (this *Application) SetClientPacketChan(ch chan *network.Packet) {
	if ch != nil {
		this.clientPacket = ch
	}
}

// 设置服务端消息通道
func (this *Application) SetServerPacketChan(ch chan *network.Packet) {
	if ch != nil {
		this.serverPacket = ch
	}
}

// 设置 serve-> client 消息通道
func (this *Application) SetStcPacketChan(ch chan *network.Packet) {
	if ch != nil {
		this.stcPkt = ch
	}
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
		this.agentMgr.SetPostman(this.postman)
		this.agentMgr.SetEventChan(this.agentEvt)
		this.agentMgr.SetClientPacketChan(this.clientPacket)
		this.agentMgr.SetServerPacketChan(this.serverPacket)
		this.agentMgr.SetStcPacketChan(this.stcPkt)
		this.agentMgr.Run()
	}

	// 接收器
	this.runAcceptor()
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
	if len(this.Options.Clusters) <= 0 {
		return
	}

	this.newPostman()

	this.postman.Run()
}

// 转发
func (this *Application) newPostman() {
	this.postman = network.NewPostman(this.Options.Appid, this.Options.Sid, this.Options.Clusters)
	this.postman.SetClientPacketChan(this.clientPacket)
	this.postman.SetServerPacketChan(this.serverPacket)
	this.postman.SetStcPacketChan(this.stcPkt)
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
