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

	"github.com/zpab123/sco/log"
	"github.com/zpab123/sco/module"
)

// /////////////////////////////////////////////////////////////////////////////
// Application

// 1个通用服务器对象
type Application struct {
	mods       []module.IModule // 模块集合
	stopGroup  sync.WaitGroup   // 停止等待组
	closeSig   chan bool        // 关闭信号
	signalChan chan os.Signal   // 操作系统信号
}

// 创建1个新的 Application 对象
func NewApplication() *Application {

	// 创建对象
	cs := make(chan bool, 1)
	mod := make([]module.IModule, 0)
	sch := make(chan os.Signal, 1)

	// 创建 app
	a := Application{
		mods:       mod,
		closeSig:   cs,
		signalChan: sch,
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
	log.Logger.Info(
		"[Application] 关闭中",
	)

	// 发出关闭信号
	this.closeSig <- true
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

		time.Sleep(C_STOP_TIME_OUT)
		log.Logger.Warn(
			"[Application] 关闭超时，强制关闭",
			log.Uint16("超时时间(秒)=", uint16(C_STOP_TIME_OUT/time.Second)),
		)

		os.Exit(1)
	} else {
		log.Logger.Warn(
			"[Application] 异常的操作系统信号",
			log.String("sid=", sig.String()),
		)
	}
}

// 启动一个 mod
func (this *Application) runMod(mod module.IModule) {
	mod.Run(this.closeSig)
	this.stopGroup.Done()
}
