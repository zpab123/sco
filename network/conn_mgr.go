// /////////////////////////////////////////////////////////////////////////////
// 连接管理

package network

import (
	"golang.org/x/net/websocket" // websocket
)

// /////////////////////////////////////////////////////////////////////////////
// ConnectionManager

// 连接管理
type ConnectionManager struct {
}

// 收到1个新的 websocket 连接对象
func (this *ConnectionManager) OnNewWsConn(wsconn *websocket.Conn) {
	zaplog.Debugf("收到1个新的 websocket 连接。ip=%s", wsconn.RemoteAddr())

}
