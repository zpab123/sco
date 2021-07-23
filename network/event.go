package network

type ConnEvent struct {
	Type uint8    // 事件类型
	conn *TcpConn // 连接
}
