// /////////////////////////////////////////////////////////////////////////////
// network 需要的协议

package protocol

// /////////////////////////////////////////////////////////////////////////////
// network

// 远端->本地握手请求
type HandshakeReq struct {
	Key string // 通信key
}

// 本地->远端握手成功
type HandshakeOk struct {
	Code      uint32 // 握手结果
	Heartbeat uint32 // 心跳时间
}

// 本地->远端握手失败
type HandshakeFail struct {
	Code uint32 // 握手结果
}
