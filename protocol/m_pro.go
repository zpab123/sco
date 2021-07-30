// /////////////////////////////////////////////////////////////////////////////
// 常量-接口-types

package protocol

// /////////////////////////////////////////////////////////////////////////////
// 常量

// 服务id
const (
	C_SID_NET     uint16 = 0 // 网络消息
	C_SID_CLUSTER uint16 = 1 // 集群服务
)

// 连接消息id
const (
	C_MID_HANDSHAKE_REQ uint16 = 0 // 握手请求
	C_MID_HANDSHAKE_RES uint16 = 1 // 握手回复
	C_MID_ACK           uint16 = 2 // 握手 ack
	C_MID_HEARTBEAT     uint16 = 3 // 心跳
)

// 网络消息
const (
	C_MID_NET_WORK uint16 = 0 // 某个连接进入工作
)

// 集群消息
const (
	C_MID_SVCREG_REQ uint16 = 0 // 服务注册请求
	C_MID_SVCREG_RES uint16 = 1 // 服务注册回复
)

// 消息码
const (
	C_CODE_FAIL    uint32 = 0 // 失败
	C_CODE_OK      uint32 = 1 // 成功
	C_CODE_KEY_ERR uint32 = 3 // 握手 key 错误
)
