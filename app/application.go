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
	"github.com/zpab123/sco/state"
)

// /////////////////////////////////////////////////////////////////////////////
// Application

// 1个通用服务器对象
type Application struct {
	mods       []module.IModule           // 模块集合
	stopGroup  sync.WaitGroup             // 停止等待组
	signalChan chan os.Signal             // 操作系统信号
	state      state.State                // 状态
	ctx        context.Context            // 退出 ctx
	cancel     context.CancelFunc         // 退出 ctx
	subAdd     chan *module.Subscriber    // 订阅消息
	subDel     chan *module.Subscriber    // 取消订阅
	postmans   map[uint32]*module.Postman // 模块消息投递员
	modMsg     chan module.Messge         // 模块消息
}

// 创建1个新的 Application 对象
func NewApplication() *Application {

	// 创建对象
	mod := make([]module.IModule, 0)
	sch := make(chan os.Signal, 1)
	cx, cc := context.WithCancel(context.Background())
	sa := make(chan *module.Subscriber, 100)
	sd := make(chan *module.Subscriber, 100)
	pm := make(map[uint32]*module.Postman, 0)
	msg := make(chan module.Messge, 100)

	// 创建 app
	a := Application{
		mods:       mod,
		signalChan: sch,
		ctx:        cx,
		cancel:     cc,
		subAdd:     sa,
		subDel:     sd,
		postmans:   pm,
		modMsg:     msg,
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
		mod.SetMsgMgr(this)
		mod.OnInit()
		this.mods = append(this.mods, mod)
	}
}

// -----------------------------------------------------------------------------
// module.IMessgeMgr 接口

// 订阅消息
func (this *Application) Subscribe(mod module.IModule, msgId uint32, ch chan module.Messge) {
	if mod == nil || ch == nil {
		return
	}

	suber := module.Subscriber{
		SuberId: mod.GetId(),
		MsgId:   msgId,
		MsgChan: ch,
	}

	this.subAdd <- &suber
}

// 取消订阅
func (this *Application) Unsubscribe(mod module.IModule, msgId uint32) {
	if mod == nil {
		return
	}

	suber := module.Subscriber{
		SuberId: mod.GetId(),
		MsgId:   msgId,
	}

	this.subDel <- &suber
}

// 广播消息
func (this *Application) Broadcast(mod module.IModule, msgId uint32, data interface{}) {
	msg := module.Messge{
		Id:     msgId,
		Type:   module.C_MSG_TYPE_BROAD,
		Sender: mod.GetId(),
		Data:   data,
	}

	this.modMsg <- msg
}

// 向某个模块发送消息
func (this *Application) Post(mod module.IModule, recver uint32, msgId uint32, data interface{}) {
	msg := module.Messge{
		Id:     msgId,
		Type:   module.C_MSG_TYPE_DIRECT,
		Sender: mod.GetId(),
		Recver: recver,
		Data:   data,
	}

	this.modMsg <- msg
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
		case sig := <-this.signalChan: // os 信号
			this.onSignal(sig)
		case suber := <-this.subAdd: // 订阅请求
			this.onSubAdd(suber)
		case suber := <-this.subDel: // 取消订阅
			this.onSubDel(suber)
		case msg := <-this.modMsg: // 模块消息
			this.onModMsg(msg)
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

// 订阅请求
func (this *Application) onSubAdd(suber *module.Subscriber) {
	pm, ok := this.postmans[suber.MsgId]
	if !ok {
		pm = module.NewPostman(suber.MsgId)
		this.postmans[suber.MsgId] = pm
	}

	pm.AddSuber(suber)
}

// 取消订阅
func (this *Application) onSubDel(suber *module.Subscriber) {
	pm, ok := this.postmans[suber.MsgId]
	if ok {
		pm.DelSuber(suber)
	}
}

// 模块消息
func (this *Application) onModMsg(msg module.Messge) {
	switch msg.Type {
	case module.C_MSG_TYPE_BROAD: // 广播类
		this.publish(msg)
	case module.C_MSG_TYPE_DIRECT: // 定向类
		this.toMod(msg)
	}
}

// 发布一个消息
func (this *Application) publish(msg module.Messge) {
	pm, ok := this.postmans[msg.Id]
	if ok {
		pm.Dispath(msg)
	}
}

// 发送给某个 mod
func (this *Application) toMod(msg module.Messge) {
	for i, _ := range this.mods {
		if this.mods[i].GetId() == msg.Recver {
			ch := this.mods[i].GetMsgChan()
			if ch != nil {
				ch <- msg
			}
			return
		}
	}
}
