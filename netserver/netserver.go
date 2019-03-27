// /////////////////////////////////////////////////////////////////////////////
// net 服务器

package netserver

import (
	"context"
	"net"
	"sync"

	"github.com/pkg/errors"          // 异常
	"github.com/zpab123/sco/model"   // 全局模型
	"github.com/zpab123/sco/network" // 网络

	//"github.com/zpab123/sco/session" // session 组件
	"github.com/zpab123/sco/state" // 状态管理
	"github.com/zpab123/syncutil"  // 原子操作工具
	"github.com/zpab123/zaplog"    // log 日志库
	"golang.org/x/net/websocket"   // websocket
)

// /////////////////////////////////////////////////////////////////////////////
// NetServer 组件

// 网络服务器
type NetServer struct {
	cmptName  string                // 组件名字
	stopGroup sync.WaitGroup        // 停止等待组
	acceptor  network.IAcceptor     // acceptor 连接器
	stateMgr  *state.StateManager   // 状态管理
	connNum   syncutil.AtomicUint32 // 当前连接数
	option    *TNetServerOpt        // 配置参数
}

// 新建1个 NetServer 对象
func NewNetServer(addr *network.TLaddr, opt *TNetServerOpt) (model.IComponent, error) {
	var err error
	var a network.IAcceptor

	// 参数效验
	if nil == opt {
		opt = NewTNetServerOpt()
	}

	// 创建对象
	sm := state.NewStateManager()

	// 创建 NetServer
	actor := &NetServer{
		cmptName: C_CMPT_NAME,
		stateMgr: sm,
		option:   opt,
	}

	// 创建 NetServer
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

// 启动 NetServer
func (this *NetServer) Run(ctx context.Context) {
	var err error

	// 改变状态： 启动中
	if !this.stateMgr.SwapState(state.C_INIT, state.C_RUNING) {
		if !this.stateMgr.SwapState(state.C_STOPED, state.C_RUNING) {
			err = errors.Errorf("network.NetServer 组件启动失败，状态错误。当前状态=%d，正确状态=%d或=%d", this.stateMgr.GetState(), state.C_INIT, state.C_STOPED)
		}

		return
	}

	// acceptor 检查
	if nil == this.acceptor {
		err = errors.New("network.NetServer 组件启动失败。acceptor=nil")

		return
	}

	// 启动 acceptor
	if err = this.acceptor.Run(); nil != err {
		return
	}

	this.stateMgr.SetState(state.C_WORKING)

	zaplog.Infof("NetServer 组件启动成功")

	// 等待结束信号
	// <-ctx.Done()
}

// 停止 NetServer
func (this *NetServer) Stop() {
	var err error

	// 状态效验
	if !this.stateMgr.SwapState(state.C_WORKING, state.C_STOPING) {
		err = errors.Errorf("network.NetServer 组件停止失败，状态错误。当前状态=%d，正确状态=%d", this.stateMgr.GetState(), state.C_WORKING)

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

	zaplog.Infof("network.NetServer 组件停止成功")

	return
}

// 获取组件名字
func (this *NetServer) Name() string {
	return this.cmptName
}

// 收到1个新的 websocket 连接对象
func (this *NetServer) OnNewWsConn(wsconn *websocket.Conn) {
	zaplog.Debugf("收到1个新的 websocket 连接。ip=%s", wsconn.RemoteAddr())

	// 超过最大连接数
	if this.connNum.Load() >= this.option.MaxConn {
		wsconn.Close()

		zaplog.Debugf("Acceptor 达到最大连接数，关闭新连接。当前连接数=%d", this.connNum.Load())
	}

	// 参数设置
	wsconn.PayloadType = websocket.BinaryFrame // 以二进制方式接受数据

	// 创建 session 对象
	this.createSession(wsconn, true)
}

// 创建 session 对象
func (this *NetServer) createSession(netconn net.Conn, isWebSocket bool) {
	/*
		// 创建 socket
		socket := &network.Socket{
			Conn: netconn,
		}

		// 创建 session
		if this.option.ForClient {
			cses := session.NewClientSession(socket, this.sessionMgr, this.option.ClientSesOpt)

			cses.Run()
		} else {
			sses := session.NewServerSession(socket, this.sessionMgr, this.option.ServerSesOpt)

			sses.Run()
		}

		this.connNum.Add(1)
	*/
}
