// /////////////////////////////////////////////////////////////////////////////
// session

package network

// /////////////////////////////////////////////////////////////////////////////
// session

type Session struct {
	id      uint32   // id 编号
	kind    int8     // 种类 0=前端 1=后端
	client  uint32   // 客户端id
	sender  uint16   // 发送者
	conn    IConn    // 连接
	postMan *Postman // 转发对象
	tag1    uint8    // 自定义标签
	tag2    uint32   // 自定义标签
	tag3    int      // 自定义标签
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

// 获取 id
func (this *Session) ID() uint32 {
	return this.id
}

// 获取 client
func (this *Session) Cilent() uint32 {
	return this.Cilent()
}

// 设置标签
// 这里多线程是不安全的
func (this *Session) SetTag1(v uint8) {
	this.tag1 = v
}

// 设置标签
// 这里多线程是不安全的
func (this *Session) SetTag2(v uint32) {
	this.tag2 = v
}

// 设置标签
// 这里多线程是不安全的
func (this *Session) SetTag3(v int) {
	this.tag3 = v
}

// 设置标签
// 这里多线程是不安全的
func (this *Session) Tag1() uint8 {
	return this.tag1
}

// 设置标签
// 这里多线程是不安全的
func (this *Session) Tag2() uint32 {
	return this.tag2
}

// 设置标签
// 这里多线程是不安全的
func (this *Session) Tag3() int {
	return this.tag3
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

// 发送给客户端
func (this *Session) ToClientPacket(pkt *Packet) {
	if pkt == nil || this.conn == nil {
		return
	}

	if pkt.kind != C_PKT_KIND_SER_CLI {
		return
	}

	if this.sender == 0 {
		// 直连
		this.conn.Send(pkt)
	} else {
		// 非直连
		this.forwardPkt(pkt)
	}
}

// 发送给某类服务
func (this *Session) ToService(sid, mid uint16, data []byte) {

}

// 发送给某类服务
func (this *Session) ToServicePacket(pkt *Packet) {
	if pkt == nil || this.postMan == nil {
		return
	}

	this.postMan.Post(pkt)
}

// 发送给某个服务器
func (this *Session) ToServer(sid, mid uint16, data []byte) {

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

// 转发给服务器
func (this *Session) forwardPkt(pkt *Packet) {
	if this.postMan == nil {
		return
	}

	this.postMan.Post(pkt)
}
