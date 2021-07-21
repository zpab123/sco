// /////////////////////////////////////////////////////////////////////////////
// websocket 接收器

package network

import (
	"net"
	"net/http"
	"sync"

	"github.com/pkg/errors"
	"github.com/zpab123/sco/log"
	"golang.org/x/net/websocket"
)

// /////////////////////////////////////////////////////////////////////////////
// wsAcceptor 对象

// websocket 接收器
type WsAcceptor struct {
	laddr      string         // 监听地址
	listener   net.Listener   // 侦听器： 用于http服务器
	httpServer http.Server    // http 服务器
	certFile   string         // TLS 加密文件
	keyFile    string         // TLS 解密key
	connMgr    IWsConnManager // websocket 连接管理
	stopGroup  sync.WaitGroup // 停止等待组
}

// 创建1个新的 wsAcceptor 对象
//
// 成功，返回 *WsAcceptor nil
// 失败，返回 nil error
func NewWsAcceptor(laddr string) (IAcceptor, error) {
	var err error

	// 参数效验
	if "" == laddr {
		err = errors.New("创建 WsAcceptor 失败。参数 laddr 为空")
		return nil, err
	}

	ws := WsAcceptor{
		laddr: laddr,
	}

	return &ws, nil
}

// -----------------------------------------------------------------------------
// IAcceptor 接口

// 启动 wsAcceptor
//
// 成功，返回 nil
// 失败，返回 error
func (this *WsAcceptor) Run() error {
	if nil == this.connMgr {
		err := errors.New("WsAcceptor 启动失败：connMgr 为空")
		return err
	}

	lis, err := net.Listen("tcp", this.laddr)
	if nil != err {
		return err
	}
	this.listener = lis

	// 侦听新连接
	this.stopGroup.Add(1)
	go this.accept()

	return nil
}

// 停止 wsAcceptor
//
// 成功，返回 nil
// 失败，返回 error
func (this *WsAcceptor) Stop() error {
	defer log.Logger.Sync()

	log.Logger.Debug(
		"[WsAcceptor] 停止中...",
		log.String("ip=", this.laddr),
	)

	err := this.httpServer.Close()
	if nil != err {
		this.listener.Close()

		log.Logger.Warn(
			"[WsAcceptor] 停止 httpServer 失败",
			log.String("ip=", this.laddr),
			log.String("err=", err.Error()),
		)

		return err
	}

	this.stopGroup.Wait()

	log.Logger.Debug(
		"[WsAcceptor] 停止",
		log.String("ip=", this.laddr),
	)

	return nil
}

// -----------------------------------------------------------------------------
// public

// 设置连接管理
func (this *WsAcceptor) SetConnMgr(mgr IConnManager) {
	if nil != mgr {
		this.connMgr = mgr
	}
}

// 设置 tls
func (this *WsAcceptor) SetTLS(cert string, key string) {
	this.certFile = cert
	this.keyFile = key
}

// -----------------------------------------------------------------------------
// private

// 侦听连接
func (this *WsAcceptor) accept() {
	defer func() {
		this.stopGroup.Done()
	}()

	// 创建 mux
	mux := http.NewServeMux()
	handler := websocket.Handler(this.connMgr.OnWsConn) // 路由函数
	mux.Handle("/", handler)                            // 不带路由
	//mux.Handle("/ws", handler)                        // ws 路由

	// 创建 httpServer
	this.httpServer = http.Server{
		Addr:    this.laddr,
		Handler: mux,
	}

	// 开启服务器
	var err error

	log.Logger.Debug(
		"[WsAcceptor] 启动成功",
		log.String("ip=", this.laddr),
	)

	if this.certFile != "" && this.keyFile != "" {
		err = this.httpServer.ServeTLS(this.listener, this.certFile, this.keyFile)
	} else {
		err = this.httpServer.Serve(this.listener)
	}

	// 错误信息
	if nil != err {
		log.Logger.Debug(
			"[WsAcceptor] 停止侦听新连接",
			log.String("ip=", this.laddr),
			log.String("err=", err.Error()),
		)
	}
}
