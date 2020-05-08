// /////////////////////////////////////////////////////////////////////////////
// 1个通用服务器对象

package app

import (
	"math/rand"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/zpab123/zaplog"
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
	sig := make(chan os.Signal, 1)

	// 创建 app
	a := Application{
		signalChan: sig,
	}

	return &a
}

// 启动 app
func (this *Application) Run() {
	// 设置随机种子
	rand.Seed(time.Now().UnixNano())

	zaplog.Infof("服务器，启动成功...")

	// 侦听结束信号
	this.waitStopSignal()
}

// 停止 app
func (this *Application) Stop() {
	zaplog.Infof("服务器，优雅退出")
	os.Exit(0)
}

// 侦听结束信号
func (this *Application) waitStopSignal() {
	// 排除信号
	signal.Ignore(syscall.Signal(10), syscall.Signal(12), syscall.SIGPIPE, syscall.SIGHUP)
	signal.Notify(this.signalChan, syscall.SIGINT, syscall.SIGTERM)

	for {
		sig := <-this.signalChan
		if syscall.SIGINT == sig || syscall.SIGTERM == sig {
			zaplog.Infof("服务器，正在停止中，请等待 ...")
			this.Stop()

			time.Sleep(C_STOP_OUT_TIME)

			zaplog.Warnf("服务器，超过 %v 秒未优雅关闭，强制关闭", C_STOP_OUT_TIME)
			os.Exit(1)
		} else {
			zaplog.Warnf("异常的操作系统信号=%s", sig)
		}
	}
}
