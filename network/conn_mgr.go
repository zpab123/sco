// /////////////////////////////////////////////////////////////////////////////
// 连接管理

package network

import (
	"github.com/zpab123/syncutil"
	"github.com/zpab123/zaplog"
	"golang.org/x/net/websocket"
)

// /////////////////////////////////////////////////////////////////////////////
// ConnMgr

// 连接管理
type ConnMgr struct {
	maxConn uint32                // 最大连接数量，超过此数值后，不再接收新连接
	connNum syncutil.AtomicUint32 // 当前连接数
}

// 新建1个 ConnMgr
func NewConnMgr(maxConn uint32) *ConnMgr {
	if 0 == maxConn {
		maxConn = C_MAX_CONN
	}

	mgr := ConnMgr{
		maxConn: maxConn,
	}

	return &mgr
}

// 停止连接管理
func (this *ConnMgr) Stop() {

}

// 收到1个新的 websocket 连接对象 [IWsConnManager]
func (this *ConnMgr) OnNewWsConn(wsconn *websocket.Conn) {
	// 超过最大连接数
	if this.connNum.Load() >= this.maxConn {
		wsconn.Close()
		zaplog.Warnf("[ConnMgr] 达到最大连接数，关闭新连接。当前连接数=%d", this.connNum.Load())
	}

	// 参数设置
	wsconn.PayloadType = websocket.BinaryFrame // 以二进制方式接受数据

	// 创建代理
	zaplog.Debugf("[ConnMgr] 新 ws 连接，ip=%s", wsconn.RemoteAddr())
	this.newAgent(wsconn, true)
}

// 创建代理
func (this *ConnMgr) newAgent(conn net.Conn, isWebSocket bool) {
	/*
		// 创建 socket
		socket := Socket{
			Conn: conn,
		}

		this.connNum.Add(1)
	*/
}
