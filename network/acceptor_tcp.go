// /////////////////////////////////////////////////////////////////////////////
// tcp 连接器

package network

import (
	"net"
	"strings"

	"github.com/pkg/errors"          // 异常库
	"github.com/zpab123/world/model" // 全局模型
	"github.com/zpab123/world/state" // 状态管理
	"github.com/zpab123/world/utils" // 工具库
	"github.com/zpab123/world/wderr" // 异常库
	"github.com/zpab123/zaplog"      // 日志库
)

// /////////////////////////////////////////////////////////////////////////////
// 包初始化

// /////////////////////////////////////////////////////////////////////////////
// TcpAcceptor 对象

// tcp 接收器
type TcpAcceptor struct {
	name     string              // 连接器名字
	laddr    *TLaddr             // 地址集合
	stateMgr *state.StateManager // 状态管理
	connMgr  ITcpConnManager     // 符合 ITcpAcceptorManager 连接管理接口的对象
	listener net.Listener        // 侦听器
}

// 创建1个新的 TcpAcceptor 对象
func NewTcpAcceptor(addr *TLaddr, mgr ITcpConnManager) (IAcceptor, error) {
	var err error

	// 参数效验
	if addr.TcpAddr == "" {
		err = errors.New("创建 TcpAcceptor 失败。参数 TcpAddr 为空")

		return nil, err
	}

	if nil == mgr {
		err = errors.New("创建 TcpAcceptor 失败。参数 ITcpConnManager=nil")

		return nil, err
	}

	// 对象
	st := state.NewStateManager()

	// 创建 TcpAcceptor
	aptor := &TcpAcceptor{
		name:     C_ACCEPTOR_NAME_TCP,
		laddr:    addr,
		stateMgr: st,
		connMgr:  mgr,
	}

	aptor.stateMgr.SetState(state.C_INIT)

	return aptor, nil
}

// 异步侦听新连接 [IAcceptor 接口]
func (this *TcpAcceptor) Run() (err error) {
	// 状态效验
	s := this.stateMgr.GetState()
	if s != state.C_INIT && s != state.C_STOPED {
		err = errors.Errorf("TcpAcceptor 启动失败，状态错误。当前状态=%d，正确状态=%d或=%d", s, state.C_INIT, state.C_STOPED)

		return
	}

	// 创建侦听器
	var ln interface{}
	ln, err = utils.DetectPort(this.laddr.TcpAddr, func(a *model.TAddress, port int) (interface{}, error) {
		return net.Listen("tcp", a.HostPortString(port))
	})

	// 创建失败
	if nil != err {
		return
	}

	// 创建成功
	this.stateMgr.SetState(state.C_WORKING)

	var ok bool
	this.listener, ok = ln.(net.Listener)
	if !ok {
		err = errors.New("TcpAcceptor 启动失败，创建 net.Listener 失败")

		return
	}

	zaplog.Infof("TcpAcceptor 启动成功。ip=%s", this.GetListenAddress())

	// 侦听连接
	go this.accept()

	return
}

// 停止侦听器 [IAcceptor 接口]
func (this *TcpAcceptor) Stop() error {
	if this.stateMgr.GetState() != state.C_WORKING {
		err := errors.Errorf("TcpAcceptor停止失败，状态错误。当前状态=%d，正确状态=%d", this.stateMgr.GetState(), state.C_WORKING)

		return err
	}

	this.stateMgr.SetState(state.C_STOPED)

	return this.listener.Close()
}

// 获取监听成功的端口
func (this *TcpAcceptor) GetPort() int {
	if this.listener == nil {
		return 0
	}

	return this.listener.Addr().(*net.TCPAddr).Port
}

// 获取监听成功的地址
func (this *TcpAcceptor) GetListenAddress() string {
	// 获取 host
	pos := strings.Index(this.laddr.TcpAddr, ":")
	if pos == -1 {
		return this.laddr.TcpAddr
	}
	host := this.laddr.TcpAddr[:pos]

	// 获取 port
	port := this.GetPort()

	return utils.JoinAddress(host, port)
}

// 侦听连接
func (this *TcpAcceptor) accept() {
	defer this.Stop()

	// 主循环
	for {
		// 接收新连接
		conn, err := this.listener.Accept()

		if this.stateMgr.GetState() != state.C_WORKING {
			return
		}

		// 监听错误
		if nil != err {
			if wderr.IsTimeoutError(err) {
				continue
			}

			zaplog.Errorf("TcpAcceptor接收新连接出现错误。err=%v", err.Error())

			break
		}

		// 处理连接进入独立线程, 防止 accept 无法响应
		if nil != this.connMgr {
			go this.connMgr.OnNewTcpConn(conn)
		}
	}
}
