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
	"github.com/zpab123/sco/module"
	"github.com/zpab123/sco/network"
	"github.com/zpab123/sco/state"
)

// /////////////////////////////////////////////////////////////////////////////
// Application

// 1个通用服务器对象
type Application struct {
	mods       []module.IModule                // 模块集合
	stopGroup  sync.WaitGroup                  // 停止等待组
	signalChan chan os.Signal                  // 操作系统信号
	state      state.State                     // 状态
	ctx        context.Context                 // 退出 ctx
	cancel     context.CancelFunc              // 退出 ctx
	subMap     map[uint16]chan *network.Packet // 订阅消息列表
	suber      map[uint16]int64                // id -> 订阅者 集合，预防重复订阅
}

// 创建1个新的 Application 对象
func NewApplication() *Application {

	// 创建对象
	mod := make([]module.IModule, 0)
	sch := make(chan os.Signal, 1)
	cx, cc := context.WithCancel(context.Background())

	// 创建 app
	a := Application{
		mods:       mod,
		signalChan: sch,
		ctx:        cx,
		cancel:     cc,
	}

	return &a
}

// 启动 app
func (this *Application) Run() {
	// 设置随机种子
	rand.Seed(time.Now().UnixNano())

	log.Logger.Debug("启动 app")

	// 启动所有模块
	for _, mod := range this.mods {
		this.stopGroup.Add(1)
		go this.runMod(mod)
	}

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

// 注册模块
func (this *Application) RegisterMod(mod module.IModule) {
	if mod != nil {
		this.mods = append(this.mods, mod)
		mod.OnInit()
	}
}

// 订阅消息
func (this *Application) Subscribe(mid uint16, c chan *network.Packet) {
	// 重复订阅验证?
	// 加入订阅列表
}

// 发布消息 dispatch  broadcast
func (this *Application) Publish(mid uint16) {
	// 复制消息
	// 发送给每个订阅者
}

// -----------------------------------------------------------------------------
// private

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
		case sig := <-this.signalChan:
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

// 启动一个 mod
func (this *Application) runMod(mod module.IModule) {
	mod.Run(this.ctx)
	this.stopGroup.Done()
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
