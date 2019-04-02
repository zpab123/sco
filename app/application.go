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

	"github.com/zpab123/sco/config" // 配置管理
	"github.com/zpab123/sco/path"   // 路径
	"github.com/zpab123/sco/state"  // 状态管理
	"github.com/zpab123/zaplog"     // log
)

// /////////////////////////////////////////////////////////////////////////////
// Application 对象

// 1个通用服务器对象
type Application struct {
	stateMgr     *state.StateManager // 状态管理
	baseInfo     *TBaseInfo          // 基础信息
	appDelegate  IAppDelegate        // 代理对象
	stopGroup    sync.WaitGroup      // stop 等待组
	serverInfo   *config.TServerInfo // 配置信息
	componentMgr *ComponentManager   // 组件管理
	signalChan   chan os.Signal      // 操作系统信号
	ctx          context.Context     // 上下文
	cancel       context.CancelFunc  // 退出通知函数
	NetService   INetService         // 网络服务组件
}

// 创建1个新的 Application 对象
func NewApplication(appType string, delegate IAppDelegate) *Application {
	// 参数验证
	if "" == appType {
		zaplog.Error("app 创建失败。 appType为空")

		os.Exit(1)
	}

	if nil == delegate {
		zaplog.Error("app 创建失败。 delegate=nil")

		os.Exit(1)
	}

	// 创建对象
	st := state.NewStateManager()
	base := &TBaseInfo{}
	cmptMgr := NewComponentManager()
	signal := make(chan os.Signal, 1)

	// 创建 app
	app := &Application{
		stateMgr:     st,
		baseInfo:     base,
		appDelegate:  delegate,
		componentMgr: cmptMgr,
		signalChan:   signal,
	}

	// 设置类型
	app.baseInfo.AppType = appType

	// 设置为无效状态
	app.stateMgr.SetState(state.C_INVALID)

	// 通知代理
	app.appDelegate.OnCreat(app)

	return app
}

// 初始化 Application
func (this *Application) Init() {
	// 状态效验
	st := this.stateMgr.GetState()
	if st != state.C_INVALID {
		zaplog.Fatal("app Init 失败，状态错误。当前状态=%d，正确状态=%d", st, state.C_INVALID)

		os.Exit(1)
	}

	// 获取主程序路径
	dir, err := path.GetMainPath()
	if err != nil {
		zaplog.Fatal("app Init 失败。读取 main 根目录失败")

		os.Exit(1)
	}
	this.baseInfo.MainPath = dir

	// 退出通知
	this.ctx, this.cancel = context.WithCancel(context.Background())

	// 默认设置
	defaultConfig(this)

	// 通知代理
	this.appDelegate.OnInit(this)

	// 状态： 初始化
	this.stateMgr.SetState(state.C_INIT)

	zaplog.Infof("app 状态：init完成 ...")
}

// 启动 app
func (this *Application) Run() {
	// 状态效验
	if !this.stateMgr.CompareAndSwap(state.C_INIT, state.C_RUNING) {
		if !this.stateMgr.CompareAndSwap(state.C_STOPED, state.C_RUNING) {
			st := this.stateMgr.GetState()
			zaplog.Fatalf("app 启动失败，状态错误。当前状态=%d，正确状态=%d或%d", st, state.C_INIT, state.C_STOPED)

			os.Exit(1)
		}
	}

	zaplog.Infof("app 状态：正在启动中 ...")

	// 设置随机种子
	rand.Seed(time.Now().UnixNano())

	// 记录启动时间
	this.baseInfo.RunTime = time.Now()

	// 创建组件
	createComponent(this)

	// 启动所有组件
	this.runComponent()

	// 侦听系统信号
	//this.waitSysSignal()

	// 状态：工作中
	this.stateMgr.SetState(state.C_WORKING)

	zaplog.Infof("app 状态：启动成功，工作中 ...")

	// 侦听结束信号
	this.waitStopSignal()
}

// 停止 app
func (this *Application) Stop() {
	// 状态效验
	if !this.stateMgr.CompareAndSwap(state.C_WORKING, state.C_STOPING) {
		zaplog.Fatalf("app 启动失败，状态错误。当前状态=%d，正确状态=%d", this.stateMgr.GetState(), state.C_WORKING)
	}

	// 停止所有组件
	this.stopComponent()

	// 等待完成
	this.stopGroup.Wait()
	this.stateMgr.SetState(state.C_STOPED)

	zaplog.Infof("%s 服务器，优雅退出", this.baseInfo.Name)

	os.Exit(0)
}

// 获取组件管理
func (this *Application) GetCmptMgr() *ComponentManager {
	return this.componentMgr
}

// 启动所有组件
func (this *Application) runComponent() {
	for _, cpt := range this.componentMgr.componentMap {
		this.stopGroup.Add(1)

		go func() {
			defer this.stopGroup.Done()

			cpt.Run(this.ctx)
		}()
	}
}

// 停止所有组件
func (this *Application) stopComponent() {
	for _, cpt := range this.componentMgr.componentMap {
		go cpt.Stop()
	}
}

// 侦听系统信号
func (this *Application) waitSysSignal() {
	// 排除信号
	signal.Ignore(syscall.Signal(10), syscall.Signal(12), syscall.SIGPIPE, syscall.SIGHUP)
	signal.Notify(this.signalChan, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		for {
			sig := <-this.signalChan
			if syscall.SIGINT == sig || syscall.SIGTERM == sig {
				zaplog.Infof("%s 服务器，正在停止中，请等待 ...", this.baseInfo.Name)

				this.Stop()

				time.Sleep(C_STOP_OUT_TIME)
				zaplog.Warnf("%s 服务器，超过 %v 秒未优雅关闭，强制关闭", this.baseInfo.Name, C_STOP_OUT_TIME)

				os.Exit(1)
			} else {
				zaplog.Errorf("异常的操作系统信号=%s", sig)
			}
		}
	}()
}

// 侦听结束信号
func (this *Application) waitStopSignal() {
	// 排除信号
	signal.Ignore(syscall.Signal(10), syscall.Signal(12), syscall.SIGPIPE, syscall.SIGHUP)
	signal.Notify(this.signalChan, syscall.SIGINT, syscall.SIGTERM)

	for {
		sig := <-this.signalChan
		if syscall.SIGINT == sig || syscall.SIGTERM == sig {
			zaplog.Infof("%s 服务器，正在停止中，请等待 ...", this.baseInfo.Name)

			this.Stop()

			time.Sleep(C_STOP_OUT_TIME)
			zaplog.Warnf("%s 服务器，超过 %v 秒未优雅关闭，强制关闭", this.baseInfo.Name, C_STOP_OUT_TIME)

			os.Exit(1)
		} else {
			zaplog.Errorf("异常的操作系统信号=%s", sig)
		}
	}
}
