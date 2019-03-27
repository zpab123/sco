// /////////////////////////////////////////////////////////////////////////////
// pcket 需要用到的对象池

package network

import (
	"sync"
)

// /////////////////////////////////////////////////////////////////////////////
// 包初始化 -- 创建 buffer 对象池 和 packet 对象池

const (
	_CAP_GROW_SHIFT = uint(2)                        // 二进制数据 位计算变量
	_MAX_BODY_LEN   = C_PKT_MAX_LEN - C_PKT_HEAD_LEN // body 最大长度
)

var (
	// buffer 容量切片，比如 [256 512 2048 ...] 容量
	buffCapSlice []uint32

	// Packet 对象池
	packetPool = sync.Pool{
		New: newPacket,
	}

	// buffer 对象池 buffLen -> *sync.Pool
	bufferPools map[uint32]*sync.Pool = map[uint32]*sync.Pool{}
)

// network 初始化函数
func init() {
	// 添加 256 512 2048 ... 等各个容量数组
	bufCap := uint32(_MIN_PAYLOAD_CAP) << _CAP_GROW_SHIFT
	for bufCap < _MAX_BODY_LEN {
		buffCapSlice = append(buffCapSlice, bufCap)

		bufCap <<= _CAP_GROW_SHIFT
	}

	// 添加最大 body 长度
	buffCapSlice = append(buffCapSlice, _MAX_BODY_LEN)

	// 创建 buffer 对象池
	for _, bufCap := range buffCapSlice {
		// 创建函数
		newBuf := func() interface{} {
			return make([]byte, _HEAD_LEN+bufCap) // 消息头 + 有效载荷
		}

		// 创建对象池
		bufferPools[bufCap] = &sync.Pool{
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

// 根据 need ，计算需要从 bufferPools 中取出哪个对象池
func getPoolKey(need uint32) uint32 {
	for _, ln := range buffCapSlice {
		if ln >= need {
			return ln
		}
	}

	return _MAX_BODY_LEN
}
