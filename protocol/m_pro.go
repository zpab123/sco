// /////////////////////////////////////////////////////////////////////////////
// protocol 包模型

package protocol

// /////////////////////////////////////////////////////////////////////////////
// 常量

// sco 框架消息 (1-100)
const (
	C_MID_INVALID       uint16 = iota // 无效消息
	C_MID_HANDSHAKE                   // 握手消息ID
	C_MID_HANDSHAKE_ACK               // 握手 ACK
	C_MID_SCO                         // 分界线： 以上由 SocConn 处理的消息
)

// sco 框架消息 (101-)
const (
	C_PKT_ID_HEARTBEAT uint16 = iota + 101 // 心跳消息
	C_PKT_ID_DATA                          // 通用消息
)

// 通用消息码(1-1000)
const (
	C_CODE_ERROR uint32 = iota // 错误类消息 0
	C_CODE_OK                  // 成功类消息 1
)

// 其他消息(1001-)
const (
	C_CODE_SHAKE_KEY_ERROR      uint32 = iota + 1001 // 握手 key 消息错误 1001
	C_CODE_SHAKE_ACCEPTOR_ERROR                      // 网络方式错误 1002
)
