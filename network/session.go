// /////////////////////////////////////////////////////////////////////////////
// session

package network

// /////////////////////////////////////////////////////////////////////////////
// session

type Session struct {
	kind   int8   // 种类 0=前端 1=后端
	client uint32 // 客户端id
	sender uint16 // 发送者
	conn   IConn  // 连接
}

func NewSession() *Session {
	s := Session{}

	return &s
}

// /////////////////////////////////////////////////////////////////////////////
// public

// 发送给客户端
func (this *Session) ToClient(sid, mid uint16, data []byte) {
	if this.sender == 0 {
		// 直连
	} else {
		// 非直连
	}
}

// 发送给某类服务
func (this *Session) ToService(sid, mid uint16, data []byte) {

}

// 发送给某个服务器
func (this *Session) ToServer(sid, mid uint16, data []byte) {

}

// /////////////////////////////////////////////////////////////////////////////
// private

// 转发给服务器
func (this *Session) forward(sid, mid uint16, data []byte) {
	// 创建pkt
	// 发送给 sender
}
