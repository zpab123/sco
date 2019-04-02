// /////////////////////////////////////////////////////////////////////////////
// 网络连接服务

package netservice

import (
	"context"
	"net"
	"sync"

	"github.com/pkg/errors"          // 异常
	"github.com/zpab123/sco/model"   // 全局模型
	"github.com/zpab123/sco/network" // 网络
	"github.com/zpab123/sco/session" // session 组件
	"github.com/zpab123/sco/state"   // 状态管理
	"github.com/zpab123/syncutil"    // 原子操作工具
	"github.com/zpab123/zaplog"      // log 日志库
	"golang.org/x/net/websocket"     // websocket
)

// /////////////////////////////////////////////////////////////////////////////
// NetService 组件

// 网络连接接收服务
type NetService struct {
	cmptName   string                  // 组件名字
	stopGroup  sync.WaitGroup          // 停止等待组
	acceptor   network.IAcceptor       // acceptor 连接器
	stateMgr   *state.StateManager     // 状态管理
	connNum    syncutil.AtomicUint32   // 当前连接数
	option     *TNetServiceOpt         // 配置参数
	sessionMgr *session.SessionManager // session 管理对象
}

// 新建1个 NetService 对象
func NewNetService(addr *network.TLaddr, opt *TNetServiceOpt) (INetService, error) {
	var err error
	var a network.IAcceptor

	// 参数效验
	if nil == opt {
		opt = NewTNetServerOpt(nil)
	}

	// 创建对象
	sm := state.NewStateManager()
	sesMgr := session.NewSessionManager()

	// 创建 NetService
	ns := &NetService{
		cmptName:   C_CMPT_NAME,
		stateMgr:   sm,
		option:     opt,
		sessionMgr: sesMgr,
	}

	// 创建 NetService
	a, err = network.NewAcceptor(opt.AcceptorName, addr, ns)
	if nil != err {
		return nil, err
	} else {
		ns.acceptor = a
	}

	// 设置为初始状态
	ns.stateMgr.SetState(state.C_INIT)

	return ns, err
}

// 启动 NetService
func (this *NetService) Run(ctx context.Context) {
	var err error

	// 改变状态： 启动中
	if !this.stateMgr.CompareAndSwap(state.C_INIT, state.C_RUNING) {
		if !this.stateMgr.CompareAndSwap(state.C_STOPED, state.C_RUNING) {
			err = errors.Errorf("network.NetService 组件启动失败，状态错误。当前状态=%d，正确状态=%d或=%d", this.stateMgr.GetState(), state.C_INIT, state.C_STOPED)
		}

		return
	}

	// acceptor 检查
	if nil == this.acceptor {
		err = errors.New("network.NetService 组件启动失败。acceptor=nil")

		return
	}

	// 启动 acceptor
	if err = this.acceptor.Run(); nil != err {
		return
	}

	this.stateMgr.SetState(state.C_WORKING)

	zaplog.Infof("NetService 组件启动成功")

	// 等待结束信号
	// <-ctx.Done()
}

// 停止 NetService
func (this *NetService) Stop() {
	var err error

	// 状态效验
	if !this.stateMgr.CompareAndSwap(state.C_WORKING, state.C_STOPING) {
		err = errors.Errorf("network.NetService 组件停止失败，状态错误。当前状态=%d，正确状态=%d", this.stateMgr.GetState(), state.C_WORKING)

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

	zaplog.Infof("network.NetService 组件停止成功")

	return
}

// 获取组件名字
func (this *NetService) Name() string {
	return this.cmptName
}

// 收到1个新的 websocket 连接对象
func (this *NetService) OnNewWsConn(wsconn *websocket.Conn) {
	zaplog.Debugf("收到1个新的 websocket 连接。ip=%s", wsconn.RemoteAddr())

	// 超过最大连接数
	if this.connNum.Load() >= this.option.MaxConn {
		wsconn.Close()

		zaplog.Warnf("NetService 达到最大连接数，关闭新连接。当前连接数=%d", this.connNum.Load())
	}

	// 参数设置
	wsconn.PayloadType = websocket.BinaryFrame // 以二进制方式接受数据

	// 创建 session 对象
	this.createSession(wsconn, true)
}

// 创建 session 对象
func (this *NetService) createSession(netconn net.Conn, isWebSocket bool) {
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
}
