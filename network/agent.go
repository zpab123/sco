// /////////////////////////////////////////////////////////////////////////////
// 代理对应于用户，用于存储原始连接信息

package network

// /////////////////////////////////////////////////////////////////////////////
// Agent

// 代理对应于用户，用于存储原始连接信息
type Agent struct {
}

func (this *Agent) Run() {
	go this.sendLoop()
	this.recvLoop()
}

func (this *Agent) Stop() {

}

// 接收线程
func (this *Agent) recvLoop() {

}

// 发送线程
func (this *Agent) sendLoop() {

}
