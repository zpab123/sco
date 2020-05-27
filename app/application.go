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

	"github.com/zpab123/sco/network"

	"github.com/zpab123/zaplog"
)

// /////////////////////////////////////////////////////////////////////////////
// Application

// 1个通用服务器对象
type Application struct {
	Options        *Options                // 配置选项
	clientAcceptor *network.ClientAcceptor // 客户端接收器
	signalChan     chan os.Signal          // 操作系统信号
	stopGroup      sync.WaitGroup          // 停止等待组
	remoteChan     chan *network.Packet    // remote 消息
	handleChan     chan *network.Packet    // 本地消息
	handler        IHandler                // 处理器
}

// 创建1个新的 Application 对象
func NewApplication() *Application {
	// 创建对象
	sig := make(chan os.Signal, 1)
	opt := NewOptions()
	rc := make(chan *network.Packet, 1000)
	hc := make(chan *network.Packet, 1000)

	// 创建 app
	a := Application{
		signalChan: sig,
		Options:    opt,
		remoteChan: rc,
		handleChan: hc,
	}
	a.init()

	return &a
}

// 启动 app
func (this *Application) Run() {
	// 设置随机种子
	rand.Seed(time.Now().UnixNano())

	// 客户端网络
	if C_APP_TYPE_FRONTEND == this.Options.AppType {
		this.newClientAcceptor()
		this.clientAcceptor.Run()
		this.stopGroup.Add(1)
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

// 注册 handler
func (this *Application) RegisterHandler(handler IHandler) {

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

// 创建 clientAcceptor
func (this *Application) newClientAcceptor() {
	this.clientAcceptor = network.NewClientAcceptor(this.Options.ClientAcceptorOpt)
}

// 分发消息
func (this *Application) dispatch() {
	for {
		select {
		case pkt := <-this.handleChan: // 本地消息
			this.handle(pkt)
		case pkt := <-this.remoteChan: // rpc 消息
			this.remote(pkt)
		}
	}
}

// 处理本地消息
func (this *Application) handle(pkt *network.Packet) {
	if this.handler != nil {
		// this.handler
	}
}

// 处理 rpc 消息
func (this *Application) remote(pkt *network.Packet) {

}
