// /////////////////////////////////////////////////////////////////////////////
// 连接管理

package network

import (
	"github.com/zpab123/syncutil"
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

// 收到1个新的 websocket 连接对象 [IWsConnManager]
func (this *ConnMgr) OnNewWsConn(wsconn *websocket.Conn) {

}
