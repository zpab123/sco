// /////////////////////////////////////////////////////////////////////////////
// tcp 连接器

package network

import (
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

// 启动 wsAcceptor
//
// 成功，返回 nil
// 失败，返回 error
func (this *TcpAcceptor) Run() error {
	lt, err := net.Listen("tcp", this.laddr)
	if nil != err {
		return err
	}

	this.listener = lt

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

// 侦听连接
func (this *TcpAcceptor) accept() {
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
