// /////////////////////////////////////////////////////////////////////////////
// tcp 连接器

package network

import (
	"crypto/tls"
	"net"

	"github.com/pkg/errors"
	"github.com/zpab123/zaplog"
)

// /////////////////////////////////////////////////////////////////////////////
// TcpAcceptor 对象

// tcp 接收器
type TcpAcceptor struct {
	laddr    string          // 侦听地址
	listener net.Listener    // 侦听器
	connMgr  ITcpConnManager // websocket 连接管理
	certFile string          // TLS加密文件
	keyFile  string          // TLS解密key
}

// 新建1个 tcp 接收器
// 成功： 返回 *TcpAcceptor, nil
// 失败： 返回 nil, error
func NewTcpAcceptor(laddr string) (IAcceptor, error) {
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
	go this.accept()

	return nil
}

// 停止 TcpAcceptor
//
// 成功，返回 nil
// 失败，返回 error
func (this *TcpAcceptor) Stop() error {
	return this.listener.Close()
}

// 设置连接管理
func (this *TcpAcceptor) SetConnMgr(mgr IConnManager) {
	if nil != mgr {
		this.connMgr = mgr
	}
}

// 设置 tls
func (this *TcpAcceptor) SetTLS(cert string, key string) {
	this.certFile = cert
	this.keyFile = key
}

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
	zaplog.Debugf("[TcpAcceptor] 启动成功。ip=%s", this.laddr)

	for {
		conn, err := this.listener.Accept()
		if nil != err {
			zaplog.Debugf("[TcpAcceptor] 停止侦听新连接。err=%s", err.Error())
			return
		}

		if nil != this.connMgr {
			go this.connMgr.OnTcpConn(conn)
		}
	}
}
