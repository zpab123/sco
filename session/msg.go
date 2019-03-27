// /////////////////////////////////////////////////////////////////////////////
// 消息库

package session

import (
	"github.com/zpab123/sco/network" // 网络
)

// /////////////////////////////////////////////////////////////////////////////
// ClientMsg

// ClientSession 消息
type ClientMsg struct {
	session *ClientSession  // session 对象
	packet  *network.Packet // packet 数据包
}

// 创建1个 ClientMsg
func NewClientMsg(ses *ClientSession, pkt *network.Packet) *ClientMsg {
	msg := &ClientMsg{
		session: ses,
		packet:  pkt,
	}

	return msg
}

// 获取 session 对象
func (this *ClientMsg) GetSession() *ClientSession {
	return this.session
}

// 获取 packet 对象
func (this *ClientMsg) GetPacket() *network.Packet {
	return this.packet
}

// /////////////////////////////////////////////////////////////////////////////
// ServerMsg

// ServerSession 消息
type ServerMsg struct {
	session *ServerSession  // session 对象
	packet  *network.Packet // packet 数据包
}

// 创建1个 ClientMsg
func NewServerMsg(ses *ServerSession, pkt *network.Packet) *ServerMsg {
	msg := &ServerMsg{
		session: ses,
		packet:  pkt,
	}

	return msg
}

// 获取 session 对象
func (this *ServerMsg) GetSession() *ServerSession {
	return this.session
}

// 获取 packet 对象
func (this *ServerMsg) GetPacket() *network.Packet {
	return this.packet
}
