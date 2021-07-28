// /////////////////////////////////////////////////////////////////////////////
// 常量-接口-types

package protocol

// /////////////////////////////////////////////////////////////////////////////
// 常量

// mid
const (
	C_MID_INVALID uint16 = 0 // 无效消息
	C_MID_SCO     uint16 = 1 // sco 内部消息
	C_MID_CLUSTER uint16 = 2 // 集群服务
)

// sco 框架消息
const (
	C_SID_CONN_WORKING  uint16 = 4   // conn 进入工作状态
	C_SID_HANDSHAKE_REQ uint16 = 100 // 握手请求
	C_SID_HANDSHAKE_RES uint16 = 101 // 握手回复
	C_SID_ACK           uint16 = 102 // 握手 ack
	C_SID_HEARTBEAT     uint16 = 103 // 心跳
)

// 网关消息
const (
	C_SID_SVCREG_REQ uint16 = 0 // 服务注册请求
	C_SID_SVCREG_RES uint16 = 1 // 服务注册回复
)

// 消息码
const (
	C_CODE_FAIL    uint32 = 0 // 失败
	C_CODE_OK      uint32 = 1 // 成功
	C_CODE_KEY_ERR uint32 = 3 // 握手 key 错误
)
