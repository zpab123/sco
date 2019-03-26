// /////////////////////////////////////////////////////////////////////////////
// 支持格式配置的网络接收器组件

package acceptor

import (
	"sync"

	"github.com/pkg/errors"          // 异常
	"github.com/zpab123/sco/model"   // 全局模型
	"github.com/zpab123/sco/network" // 网络
	"github.com/zpab123/world/state" // 状态管理
	"github.com/zpab123/zaplog"      // log 日志库
	"golang.org/x/net/websocket"     // websocket
)

// 网络连接对象，支持 websocket tcp
type Acceptor struct {
	cmptName  string              // 组件名字
	stopGroup sync.WaitGroup      // 停止等待组
	acceptor  network.IAcceptor   // acceptor 连接器
	stateMgr  *state.StateManager // 状态管理
}

// 新建1个 Acceptor 对象
func NewAcceptor(addr *network.TLaddr, opt *TAcceptorOpt) (model.IComponent, error) {
	var err error
	var a network.IAcceptor

	// 参数效验
	if nil == opt {
		opt = NewTAcceptorOpt()
	}

	// 创建对象
	sm := state.NewStateManager()

	// 创建 Acceptor
	actor := &Acceptor{
		cmptName: C_CMPT_NAME,
		stateMgr: sm,
	}

	// 创建 Acceptor
	a, err = network.NewAcceptor(opt.AcceptorName, addr, actor)
	if nil != err {
		return nil, err
	} else {
		actor.acceptor = a
	}

	// 设置为初始状态
	actor.stateMgr.SetState(state.C_INIT)

	return actor, err
}

// 启动 Acceptor
func (this *Acceptor) Run() {
	var err error

	// 改变状态： 启动中
	if !this.stateMgr.SwapState(state.C_INIT, state.C_RUNING) {
		if !this.stateMgr.SwapState(state.C_STOPED, state.C_RUNING) {
			err = errors.Errorf("Acceptor 组件启动失败，状态错误。当前状态=%d，正确状态=%d或=%d", this.stateMgr.GetState(), state.C_INIT, state.C_STOPED)
		}

		return
	}

	// acceptor 检查
	if nil == this.acceptor {
		err = errors.New("Acceptor 组件启动失败。acceptor=nil")

		return
	}

	// 启动 acceptor
	if err = this.acceptor.Run(); nil != err {
		return
	}

	this.stateMgr.SetState(state.C_WORKING)

	zaplog.Infof("Acceptor 组件启动成功")
}

// 停止 Acceptor
func (this *Acceptor) Stop() {
	var err error

	// 状态效验
	if !this.stateMgr.SwapState(state.C_WORKING, state.C_STOPING) {
		err = errors.Errorf("Acceptor 组件停止失败，状态错误。当前状态=%d，正确状态=%d", this.stateMgr.GetState(), state.C_WORKING)

		return
	}

	// 停止 acceptor
	if err = this.acceptor.Stop(); nil != err {
		return
	}

	// 关闭所有 session

	// 阻塞-等待停止
	// this.stopGroup.Wait()

	// 改变状态：关闭完成
	this.stateMgr.SetState(state.C_STOPED)

	zaplog.Infof("Acceptor 组件停止成功")

	return
}

// 获取组件名字
func (this *Acceptor) Name() string {
	return this.cmptName
}

// 收到1个新的 websocket 连接对象
func (this *Acceptor) OnNewWsConn(wsconn *websocket.Conn) {

}
