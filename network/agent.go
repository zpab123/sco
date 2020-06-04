// /////////////////////////////////////////////////////////////////////////////
// 代理对应于用户，用于存储原始连接信息

package network

// "github.com/zpab123/sco/protocol"

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

// 接收线程
func (this *Agent) Run() {
	// 发送线程
	go this.sendLoop()
	this.sendLoop()
}

// 接收线程
func (this *Agent) recvLoop() {
	for {
		pkt, err := this.socket.RecvPacket()
		if nil == err {
			return
		}

		if nil != pkt {
			// 处理
			continue
		}
	}
}

// 发送线程
func (this *Agent) sendLoop() {
	for {
		err := this.socket.Flush()
		if nil != err {
			break
		}
	}
}
