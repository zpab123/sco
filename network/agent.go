// /////////////////////////////////////////////////////////////////////////////
// 代理对应于用户，用于存储原始连接信息

package network

import (
	"github.com/zpab123/zaplog"
	// "github.com/zpab123/sco/protocol"
)

// /////////////////////////////////////////////////////////////////////////////
// Agent

// 代理对应于用户，用于存储原始连接信息
type Agent struct {
	options *TAgentOpt // 配置参数
	socket  *Socket    // socket
}

// 新建1个 *Agent 对象
func NewAgent(socket *Socket, opt *TAgentOpt) *Agent {
	if nil == opt {
		opt = NewTAgentOpt()
	}

	a := Agent{
		options: opt,
		socket:  socket,
	}

	return &a
}

// 启动
func (this *Agent) Run() {
	// 发送线程
	go this.sendLoop()
	this.recvLoop() // 接收循环，这里不能 go this.recvLoop()，会导致 websocket 连接直接断开
}

// 停止
func (this *Agent) Stop() {
	this.socket.Close()
}

// 接收线程
func (this *Agent) recvLoop() {
	defer func() {
		zaplog.Debugf("[Agent] recvLoop 结束")
		this.socket.SendPacket(nil) // 用于结束 sendLoop
	}()

	for {
		pkt, err := this.socket.RecvPacket()
		if nil != err {
			return
		}

		if nil != pkt {
			// 处理
			zaplog.Debugf("Agent recvLoop mid=%d", pkt.mid)
			continue
		}
	}
}

// 发送线程
func (this *Agent) sendLoop() {
	defer func() {
		zaplog.Debugf("[Agent] sendLoop 结束")
		this.Stop()
	}()

	for {
		err := this.socket.Flush()
		if nil != err {
			break
		}
	}
}
