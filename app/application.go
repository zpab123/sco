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
	connMgr    network.IConnManager // 连接管理
	acceptors  []network.IAcceptor  // 接收器切片
	packetChan chan *network.Packet // 网络数据包
	svcs       []svc.IService       // 服务列表
	stopGroup  sync.WaitGroup       // 停止等待组
	signalChan chan os.Signal       // 操作系统信号
	state      state.State          // 状态
	ctx        context.Context      // 退出 ctx
	cancel     context.CancelFunc   // 退出 ctx
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
		acceptors:  acc,
		svcs:       ss,
		signalChan: sch,
		ctx:        cx,
		cancel:     cc,
		packetChan: pc,
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

	// 侦听信号
	this.listenSignal()

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
	this.cancel()
	go this.onStop()
	this.stopGroup.Wait()

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

	this.acceptors = append(this.acceptors, acc)

	return nil
}

// 注册服务
func (this *Application) RegService(s svc.IService) {
	if s != nil {
		s.Init()
		this.svcs = append(this.svcs, s)
	}
}

// -----------------------------------------------------------------------------
// private

// 启动网络
func (this *Application) runNet() {
	// 连接管理
	if nil == this.connMgr {
		this.newConnMgr()
	}

	// 接收器
	this.runAcceptor()
}

// 创建默认连接管理
func (this *Application) newConnMgr() {
	this.connMgr = network.NewConnMgr(10000)
	this.connMgr.SetPacketChan(this.packetChan)
}

// 启动接收器
func (this *Application) runAcceptor() {
	for _, acc := range this.acceptors {
		acc.SetConnMgr(this.connMgr)
		// go acc.Run()
		acc.Run()
	}
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
		case pkt := <-this.packetChan: // 网络消息
			this.onPacket(pkt)
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
func (this *Application) onPacket(pkt *network.Packet) {

}
