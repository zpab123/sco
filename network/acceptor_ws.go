// /////////////////////////////////////////////////////////////////////////////
// websocket 接收器

package network

import (
	"net"
	"net/http"

	"github.com/pkg/errors"      // 异常库
	"github.com/zpab123/zaplog"  // 日志库
	"golang.org/x/net/websocket" // websocket 库
)

// /////////////////////////////////////////////////////////////////////////////
// wsAcceptor 对象

// websocket 接收器
type WsAcceptor struct {
	laddr      string         // 监听地址
	listener   net.Listener   // 侦听器： 用于http服务器
	httpServer http.Server    // http 服务器
	certFile   string         // TLS加密文件
	keyFile    string         // TLS解密key
	connMgr    IWsConnManager // websocket 连接管理
}

// 创建1个新的 wsAcceptor 对象
//
// 创建成功，error=nil
func NewWsAcceptor(laddr string) (IAcceptor, error) {
	var err error

	// 参数效验
	if "" == laddr {
		err = errors.New("创建 WsAcceptor 失败:参数 laddr 为空")
		return nil, err
	}

	ws := WsAcceptor{
		laddr: laddr,
	}

	return &ws, nil
}

// 启动 wsAcceptor
func (this *WsAcceptor) Run() error {
	var err error
	this.listener, err = net.Listen("tcp", this.laddr)
	if nil != err {
		return err
	}

	// 侦听新连接
	go this.accept()

	return nil
}

// 停止 wsAcceptor
func (this *WsAcceptor) Stop() error {
	var err error
	err = this.httpServer.Close()
	if nil == err {
		this.listener.Close()
	} else {
		err = this.listener.Close()
	}

	zaplog.Debugf("WsAcceptor 停止服务。ip=%s", this.laddr)

	return err
}

// 设置连接管理
func (this *WsAcceptor) SetConnMgr(mgr IConnManager) {
	if nil != mgr {
		this.connMgr = mgr
	}
}

// 侦听连接
func (this *WsAcceptor) accept() {
	// 创建 mux
	mux := http.NewServeMux()
	handler := websocket.Handler(this.connMgr.OnNewWsConn) // 路由函数
	mux.Handle("/", handler)                               // 客户端需要在url后面加上 /ws 路由

	// 创建 httpServer
	this.httpServer = http.Server{
		Addr:    this.laddr,
		Handler: mux,
	}

	// 开启服务器
	var err error
	zaplog.Debugf("WsAcceptor 启动成功。ip=%s", this.laddr)

	if this.certFile != "" && this.keyFile != "" {
		err = this.httpServer.ServeTLS(this.listener, this.certFile, this.keyFile)
	} else {
		err = this.httpServer.Serve(this.listener)
	}

	// 错误信息
	if nil != err {
		zaplog.Debugf("WsAcceptor 停止侦听新连接，goroutine 退出。ip=%s，err=%s", this.laddr, err)
	}
}
