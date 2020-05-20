// /////////////////////////////////////////////////////////////////////////////
// 通信使用的 pcket 数据包

package network

import (
	"sync"
)

// 常量 -- packet 数据大小定义
const (
	_CAP_GROW_SHIFT  uint32 = 2   // 二进制数据 位计算变量
	_MIN_PAYLOAD_CAP uint32 = 128 // buff 最小有效容量（buff 对象池使用）
)

var (
	// body 容量切片，比如 [256 512 2048 ...] 容量
	bufCapSlice []uint32

	// Packet 对象池，并发安全
	packetPool = sync.Pool{
		New: newPacket,
	}

	// body 对象池 bodyLen -> *sync.Pool
	bufPools map[uint32]*sync.Pool = map[uint32]*sync.Pool{}
)

// packet_pool 初始化
func init() {
	// 添加 256 512 2048 ... max 等各个容量数组
	bufCap := _MIN_PAYLOAD_CAP << _CAP_GROW_SHIFT

	for bufCap < C_PKT_BODY_MAX_LEN && bufCap > 0 {
		bufCapSlice = append(bufCapSlice, bufCap)
		bufCap <<= _CAP_GROW_SHIFT
	}
	bufCapSlice = append(bufCapSlice, C_PKT_BODY_MAX_LEN)

	// 创建 body 对象池
	for _, bodyCap := range bufCapSlice {
		// 创建函数
		newBuf := func() interface{} {
			return make([]byte, C_PKT_HEAD_LEN+bodyCap) // 消息头 + 有效载荷
		}

		// 创建对象池
		bufPools[bodyCap] = &sync.Pool{
			New: newBuf,
		}
	}
}

// /////////////////////////////////////////////////////////////////////////////
// 私有 api

// 从对象池中获取1个 packet
func getPacketFromPool() *Packet {
	// 获取 *Packet
	pkt := packetPool.Get().(*Packet)

	return pkt
}

// 根据 need ，计算需要从 bufPools 中取出哪个对象池
// 只读不写，所以并发安全
func getPoolKey(need uint32) uint32 {
	for _, ln := range bufCapSlice {
		if ln >= need {
			return ln
		}
	}

	return C_PKT_BODY_MAX_LEN
}
