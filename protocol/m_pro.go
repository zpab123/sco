// /////////////////////////////////////////////////////////////////////////////
// 常量-接口-types

package protocol

// /////////////////////////////////////////////////////////////////////////////
// 常量

// sco 框架消息
const (
	C_MID_INVALID       uint16 = 0   // 无效消息
	C_MID_SCO           uint16 = 1   // sco 内部消息
	C_SID_AGENT_WORKING uint16 = 4   // agent 进入工作状态
	C_SID_HANDSHAKE_REQ uint16 = 100 // 握手请求
	C_SID_HANDSHAKE_RES uint16 = 101 // 握手回复
	C_SID_ACK           uint16 = 102 // 握手 ack
	C_SID_HEARTBEAT     uint16 = 103 // 心跳
)

// 消息码
const (
	C_CODE_FAIL    uint32 = 0 // 失败
	C_CODE_OK      uint32 = 1 // 成功
	C_CODE_KEY_ERR uint32 = 3 // 握手 key 错误
)
