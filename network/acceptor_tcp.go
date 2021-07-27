// /////////////////////////////////////////////////////////////////////////////
// tcp 连接器

package network

import (
	"crypto/tls"
	"net"
	"sync"

	"github.com/pkg/errors"
	"github.com/zpab123/sco/log"
)

// /////////////////////////////////////////////////////////////////////////////
// TcpAcceptor 对象

// tcp 接收器
type TcpAcceptor struct {
	laddr     string          // 侦听地址
	listener  net.Listener    // 侦听器
	connMgr   ITcpConnManager // tcp 连接管理
	certFile  string          // TLS加密文件
	keyFile   string          // TLS解密key
	stopGroup sync.WaitGroup  // 停止等待组
}

// 新建1个 tcp 接收器
// 成功： 返回 *TcpAcceptor, nil
// 失败： 返回 nil, error
func NewTcpAcceptor(laddr string) (*TcpAcceptor, error) {
	var err error
	// 参数效验
	if "" == laddr {
		err = errors.New("参数 laddr 为空")
		return nil, err
	}

	a := TcpAcceptor{
		laddr: laddr,
	}

	return &a, nil
}

// -----------------------------------------------------------------------------
// IAcceptor 接口

// 启动 wsAcceptor
//
// 成功，返回 nil
// 失败，返回 error
func (this *TcpAcceptor) Run() error {
	if this.certFile != "" && this.keyFile != "" {
		return this.runTLS()
	}

	lis, err := net.Listen("tcp", this.laddr)
	if nil != err {
		return err
	}

	this.listener = lis
	this.stopGroup.Add(1)
	go this.accept()

	return nil
}

// 停止 TcpAcceptor
//
// 成功，返回 nil
// 失败，返回 error
func (this *TcpAcceptor) Stop() error {
	log.Logger.Debug(
		"[TcpAcceptor] 停止中...",
		log.String("ip", this.laddr),
	)

	if this.listener == nil {
		return nil
	}

	err := this.listener.Close()
	if err != nil {
		log.Logger.Warn(
			"[TcpAcceptor] 停止失败",
			log.String("ip", this.laddr),
			log.String("err", err.Error()),
		)

		return err
	}

	this.stopGroup.Wait()

	log.Logger.Debug(
		"[TcpAcceptor] 停止",
		log.String("ip", this.laddr),
	)

	return nil
}

// -----------------------------------------------------------------------------
// public

// 设置连接管理
func (this *TcpAcceptor) SetConnMgr(mgr ITcpConnManager) {
	if nil != mgr {
		this.connMgr = mgr
	}
}

// 设置 tls
func (this *TcpAcceptor) SetTLS(cert string, key string) {
	this.certFile = cert
	this.keyFile = key
}

// -----------------------------------------------------------------------------
// private

// 以 tls 方式启动
func (this *TcpAcceptor) runTLS() error {
	crt, err := tls.LoadX509KeyPair(this.certFile, this.keyFile)
	if nil != err {
		return err
	}

	c := tls.Config{
		Certificates: []tls.Certificate{crt},
	}

	lis, err := tls.Listen("tcp", this.laddr, &c)
	if nil != err {
		return err
	}

	this.listener = lis
	go this.accept()

	return nil
}

// 侦听连接
func (this *TcpAcceptor) accept() {
	defer func() {
		this.stopGroup.Done()
	}()

	log.Logger.Debug(
		"[TcpAcceptor] 启动成功",
		log.String("ip=", this.laddr),
	)

	for {
		conn, err := this.listener.Accept()
		if nil != err {
			log.Logger.Debug(
				"[TcpAcceptor] 停止侦听新连接",
				log.String("err", err.Error()),
			)

			return
		}

		if this.connMgr != nil {
			go this.connMgr.OnTcpConn(conn)
		}
	}
}
