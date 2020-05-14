// /////////////////////////////////////////////////////////////////////////////
// tcp 连接

package network

// /////////////////////////////////////////////////////////////////////////////
// Agent

// 代理对应于用户，用于存储原始连接信息
type TcpConn struct {
}

func (this *TcpConn) Run() {
	go this.sendLoop()
	this.recvLoop()
}

func (this *TcpConn) Stop() {

}

// 接收线程
func (this *TcpConn) recvLoop() {

}

// 发送线程
func (this *TcpConn) sendLoop() {

}
