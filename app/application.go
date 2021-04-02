// /////////////////////////////////////////////////////////////////////////////
// 1个通用服务器对象

package app

import (
	"math/rand"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/zpab123/sco/log"
)

// /////////////////////////////////////////////////////////////////////////////
// Application

// 1个通用服务器对象
type Application struct {
	signalChan chan os.Signal // 操作系统信号
}

// 创建1个新的 Application 对象
func NewApplication() *Application {

	// 创建对象
	sch := make(chan os.Signal, 1)

	// 创建 app
	a := Application{
		signalChan: sch,
	}

	return &a
}

// 启动 app
func (this *Application) Run() {
	// 设置随机种子
	rand.Seed(time.Now().UnixNano())

	log.Logger.Debug("启动 app")

	// 侦听信号
	this.listenSignal()

	// 主循环
	this.mainLoop()
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
		log.Logger.Warn(
			"[Application] 关闭",
		)

		os.Exit(1)
	} else {
		log.Logger.Warn(
			"[Application] 异常的操作系统信号",
			log.String("sid=", sig.String()),
		)
	}
}
