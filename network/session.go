// /////////////////////////////////////////////////////////////////////////////
// session

package network

// /////////////////////////////////////////////////////////////////////////////
// session

type Session struct {
	kind    int8     // 种类 0=前端 1=后端
	client  uint32   // 客户端id
	sender  uint16   // 发送者
	conn    IConn    // 连接
	postMan *Postman // 转发对象
}

func NewSession() *Session {
	s := Session{}

	return &s
}

// /////////////////////////////////////////////////////////////////////////////
// public

// 停止
func (this *Session) Stop() {
	if this.conn != nil {
		this.conn.Stop()
	}
}

// 直接发送
func (this *Session) Send(pkt *Packet) {
	if this.conn != nil {
		this.conn.Send(pkt)
	}
}

// 发送给客户端
func (this *Session) ToClient(sid, mid uint16, data []byte) {
	if this.sender == 0 {
		// 直连
		this.sendToClient(sid, mid, data)
	} else {
		// 非直连
		this.forward(sid, mid, data)
	}
}

// 发送给某类服务
func (this *Session) ToService(sid, mid uint16, data []byte) {

}

// 发送给某个服务器
func (this *Session) ToServer(sid, mid uint16, data []byte) {

}

// 转发
func (this *Session) Forward(pkt *Packet) {
	if pkt != nil && this.postMan != nil {
		this.postMan.Post(pkt)
	}
}

// 设置 postman
func (this *Session) SetPostman(man *Postman) {
	if man != nil {
		this.postMan = man
	}
}

// /////////////////////////////////////////////////////////////////////////////
// private

// 发送给客户端
func (this *Session) sendToClient(sid, mid uint16, data []byte) {
	if this.conn == nil {
		return
	}

	pkt := NewPacket(C_PKT_KIND_SER_CLI, 0, this.sender, sid, mid)
	if data != nil {
		pkt.AppendBytes(data)
	}

	this.conn.Send(pkt)
}

// 转发给服务器
func (this *Session) forward(sid, mid uint16, data []byte) {
	if this.postMan == nil {
		return
	}

	pkt := NewPacket(C_PKT_KIND_SER_CLI, this.client, this.sender, sid, mid)
	if data != nil {
		pkt.AppendBytes(data)
	}

	this.postMan.Post(pkt)
}
