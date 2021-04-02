// /////////////////////////////////////////////////////////////////////////////
// 常量-接口-types

package protocol

// /////////////////////////////////////////////////////////////////////////////
// 常量

// sco 框架消息 (1-100)
const (
	C_MID_INVALID       uint16 = 0 // 无效消息
	C_MID_HANDSHAKE     uint16 = 1 // 握手消息ID
	C_MID_HANDSHAKE_ACK uint16 = 2 // 握手 ACK
	C_MID_HEARTBEAT     uint16 = 3 // 心跳
	C_MID_SCO           uint16 = 4 // 分界线： 以上由 SocConn 处理的消息
)

// 消息码
const (
	C_CODE_FAIL    uint32 = 0 // 失败
	C_CODE_OK      uint32 = 1 // 成功
	C_CODE_KEY_ERR uint32 = 3 // 握手 key 错误
)
