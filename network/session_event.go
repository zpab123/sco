// /////////////////////////////////////////////////////////////////////////////
// session 事件

package network

// /////////////////////////////////////////////////////////////////////////////
// SessionEvent

// session 事件
type SessionEvent struct {
	id      int8     // 事件 id
	agent   *Agent   // 连接对象
	session *Session // session
}
