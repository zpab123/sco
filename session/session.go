// /////////////////////////////////////////////////////////////////////////////
// 用户会话信息

package session

// /////////////////////////////////////////////////////////////////////////////
// Session

// 用户会话信息
type Session struct {
	conn IConnection // 连接信息
	id   int64       // Session 唯一id
	uid  string      // 绑定的用户id
}

// 新建1个 Session
func NewSession(conn IConnection) *Session {
	s := Session{
		conn: conn,
	}

	return &s
}
