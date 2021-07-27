// /////////////////////////////////////////////////////////////////////////////
// network 需要的协议

package protocol

// /////////////////////////////////////////////////////////////////////////////
// network

// 客户端->服务器握手请求
type HandshakeReq struct {
	Key      string `json:"key"`      // 通信key
	Acceptor uint32 `json:"acceptor"` // 1=tcp;2=websocket;3=;通信方式
}

// 服务器->客户端握手结果(握手成功)
type HandshakeRes struct {
	Code      uint32 `json:"code"`      // 握手结果
	Heartbeat uint16 `json:"heartbeat"` // 心跳时间(秒)
}
