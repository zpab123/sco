// /////////////////////////////////////////////////////////////////////////////
// websocket 网络服务器

package network

import (
	"net"
	"net/http"
	"sync"

	"github.com/pkg/errors"        // 异常库
	"github.com/zpab123/sco/state" // 状态管理
	"github.com/zpab123/zaplog"    // log 日志库
	"golang.org/x/net/websocket"   // websocket 库
)

// /////////////////////////////////////////////////////////////////////////////
// WsServer 对象

// websocket 网络服务器
type WsServer struct {
	name       string              // 接收器名字
	laddr      string              // 监听地址
	connMgr    IWsConnManager      // websocket 连接管理
	certFile   string              // TLS加密文件
	keyFile    string              // TLS解密key
	listener   net.Listener        // 侦听器： 用于http服务器
	httpServer *http.Server        // http 服务器
	stopGroup  sync.WaitGroup      // 停止组
	stateMgr   *state.StateManager // 状态管理
}

// 创建1个新的 WsServer 对象
func NewWsServer(laddr string, mgr IWsConnManager, opt *TWsServerOpt) (IServer, error) {
	var err error
	// 参数效验
	if laddr == "" {
		err = errors.New("创建 WsServer 失败。参数 laddr 为空")

		return nil, err
	}

	if nil == mgr {
		err = errors.New("创建 WsServer 失败。参数 IWsConnManager=nil")

		return nil, err
	}

	// 对象
	st := state.NewStateManager()

	// 创建服务器
	s := &WsServer{
		name:     C_SERVER_NAME_WS,
		laddr:    laddr,
		connMgr:  mgr,
		certFile: opt.CertFile,
		keyFile:  opt.KeyFile,
		stateMgr: st,
	}

	s.stateMgr.SetState(state.C_INIT)

	return s, nil
}

// 启动 wsAcceptor
func (this *WsServer) Run() error {
	var err error

	// 状态效验
	if !this.stateMgr.CompareAndSwap(state.C_INIT, state.C_RUNING) {
		if !this.stateMgr.CompareAndSwap(state.C_STOPED, state.C_RUNING) {
			err = errors.Errorf("WsServer 启动失败，状态错误。当前状态=%d，正确状态=%d或=%d", this.stateMgr.GetState(), state.C_INIT, state.C_STOPED)

			return err
		}
	}

	// 创建侦听器
	this.listener, err = net.Listen("tcp", this.laddr)
	if nil != err {
		return err
	}

	this.stopGroup.Add(1)

	// 侦听新连接
	go this.accept()

	this.stateMgr.SetState(state.C_WORKING)

	return nil
}

// 停止 wsAcceptor
func (this *WsServer) Stop() error {
	var err error
	// 状态效验
	if !this.stateMgr.CompareAndSwap(state.C_WORKING, state.C_STOPING) {
		err = errors.Errorf("WsServer 停止失败，状态错误。当前状态=%d，正确状态=%d", this.stateMgr.GetState(), state.C_WORKING)

		return err
	}

	zaplog.Debugf("主动关闭 WsServer 服务。ip=%s", this.laddr)

	err = this.httpServer.Close()
	if nil != err {
		this.listener.Close()
	} else {
		err = this.listener.Close()
	}

	// 阻塞等待
	this.stopGroup.Wait()

	this.stateMgr.SetState(state.C_STOPED)

	zaplog.Debugf("WsServer 停止服务。ip=%s", this.laddr)

	return err
}

// 侦听连接
func (this *WsServer) accept() {
	defer this.stopGroup.Done()

	// 创建 mux
	mux := http.NewServeMux()
	handler := websocket.Handler(this.connMgr.OnNewWsConn) // 路由函数
	mux.Handle("/ws", handler)                             // 客户端需要在url后面加上 /ws 路由

	// 创建 httpServer
	this.httpServer = &http.Server{
		Addr:    this.laddr,
		Handler: mux,
	}

	// 开启服务器
	var err error
	zaplog.Debugf("WsServer 启动成功。ip=%s", this.laddr)

	if this.certFile != "" && this.keyFile != "" {
		err = this.httpServer.ServeTLS(this.listener, this.certFile, this.keyFile)
	} else {
		err = this.httpServer.Serve(this.listener)
	}

	// 错误信息
	if nil != err {
		zaplog.Debugf("WsServer 停止侦听新连接。ip=%s，err=%s", this.laddr, err)
	}
}
