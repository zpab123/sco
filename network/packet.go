// /////////////////////////////////////////////////////////////////////////////
// 通信使用的 pcket 数据包

package network

import (
	"encoding/binary"
	"unsafe"

	"github.com/zpab123/zaplog" // log 工具
)

// /////////////////////////////////////////////////////////////////////////////
// 包初始化

// 常量 -- packet 数据大小定义
const (
	_HEAD_LEN        = C_PKT_HEAD_LEN // 消息头长度
	_LEN_POS         = 2              // Packet 的 buffer 中，记录长度信息开始的位置： 用于 body 长度计算
	_MIN_PAYLOAD_CAP = 128            // buff 最小有效容量（buff 对象池使用）
	_BODY_LEN_MASK   = 0x7FFFFFFF     // 等于 1111111111111111111111111111111 (32个1)
)

var (
	// packet 二进制数据操作 （小端）
	packetEndian = binary.LittleEndian
)

// /////////////////////////////////////////////////////////////////////////////
// Packet 对象

// 网络通信二进制数据
type Packet struct {
	pktId     uint16                             // 用于记录 packet 类型
	bytes     []byte                             // 用于存放需要通过网络 发送/接收 的数据 （head + body）
	initBytes [_HEAD_LEN + _MIN_PAYLOAD_CAP]byte // bytes 初始化时候的 buffer 4 + 128
	readCount uint32                             // bytes 中已经读取的字节数
}

// 创建1个新的 packet 对象
func newPacket() interface{} {
	// 创建1个 packet 对象
	pkt := &Packet{}

	// 初始化buffer
	pkt.bytes = pkt.initBytes[:]

	return pkt
}

// 新建1个 Packet 对象 (从对象池创建)
func NewPacket(pktId uint16) *Packet {
	pkt := getPacketFromPool()

	pkt.SetId(pktId)

	return pkt
}

// 设置 Packet 的 id
func (this *Packet) SetId(v uint16) {
	// 记录消息类型
	NETWORK_ENDIAN.PutUint16(this.bytes[0:_LEN_POS], v)
	this.pktId = v
}

// 获取 Packet 的 id
func (this *Packet) GetId() uint16 {
	return this.pktId
}

// 获取 Packet 的 body 部分
func (this *Packet) GetBody() []byte {
	bodyLen := this.GetBodyLen()
	end := (_HEAD_LEN + bodyLen)

	return this.bytes[_HEAD_LEN:end]
}

// 获取 bytes
func (this *Packet) GetBytes() []byte {
	return this.bytes
}

// 获取 packet 的 body 字节长度
func (this *Packet) GetBodyLen() uint32 {
	bodyLen := *(*uint32)(unsafe.Pointer(&this.bytes[_LEN_POS])) & _BODY_LEN_MASK

	return bodyLen
}

// 在 Packet 的 bytes 后面添加1个 byte 数据
func (this *Packet) AppendByte(b byte) {
	// 申请buffer
	this.AllocBuffer(1)

	// 赋值
	wPos := this.getWirtePos()
	this.bytes[wPos] = b

	// body 长度+1
	this.addBodyLen(1)
}

// 从 Packet 的 bytes 中读取1个 byte 数据，并赋值给 v
func (this *Packet) ReadByte() byte {
	// 读取位置
	pPos := this.getReadPos()

	// 赋值
	v := this.bytes[pPos]

	// 读取数量+1
	this.readCount += 1

	return v
}

// 在 Packet 的 bytes 后面添加1个 bool 数据
func (this *Packet) AppendBool(b bool) {
	if b {
		this.AppendByte(1)
	} else {
		this.AppendByte(0)
	}
}

// 从 Packet 的 bytes 中读取1个 bool 数据，并赋值给v
func (this *Packet) ReadBool() bool {
	return this.ReadByte() != 0
}

// 在 Packet 的 bytes 后面，添加1个 uint16 数据
func (this *Packet) AppendUint16(v uint16) {
	// 申请buffer
	this.AllocBuffer(2)

	// 赋值
	wPos := this.getWirtePos()
	packetEndian.PutUint16(this.bytes[wPos:wPos+2], v)

	// body 长度+2
	this.addBodyLen(2)
}

// 从 Packet 的 bytes 中读取1个 uint16 数据，并赋值给v
func (this *Packet) ReadUint16() (v uint16) {
	// 读取
	pPos := this.getReadPos()
	v = packetEndian.Uint16(this.bytes[pPos : pPos+2])

	// 读取数量+2
	this.readCount += 2

	return
}

// 在 Packet 的 bytes 后面，添加1个 uint32 数据
func (this *Packet) AppendUint32(v uint32) {
	// 申请buffer
	this.AllocBuffer(4)

	// 赋值
	wPos := this.getWirtePos()
	packetEndian.PutUint32(this.bytes[wPos:wPos+4], v)

	// body 长度+2
	this.addBodyLen(4)
}

// 从 Packet 的 bytes 中读取1个 uint32 数据
func (this *Packet) ReadUint32() (v uint32) {
	// 读取
	pPos := this.getReadPos()
	v = packetEndian.Uint32(this.bytes[pPos : pPos+4])

	// 读取数量+4
	this.readCount += 4

	return
}

// 在 Packet 的 bytes 后面，添加1个 uint64 数据
func (this *Packet) AppendUint64(v uint64) {
	// 申请内存
	this.AllocBuffer(8)

	// 添加数据
	wPos := this.getWirtePos()
	packetEndian.PutUint64(this.bytes[wPos:wPos+8], v)

	// 记录长度
	this.addBodyLen(8)
}

// 从 Packet 的 bytes 中读取1个 uint64 数据
func (this *Packet) ReadUint64() (v uint64) {
	// 读取
	pPos := this.getReadPos()
	v = packetEndian.Uint64(this.bytes[pPos : pPos+8])

	// 记录读取数量
	this.readCount += 8

	return
}

// 在 Packet 的 bytes 后面，添加1个 float32 数据
func (this *Packet) AppendFloat32(f float32) {
	// 数据转换
	u32 := (*uint32)(unsafe.Pointer(&f))

	// 添加数据
	this.AppendUint32(*u32)
}

// 从 Packet 的 bytes 中读取1个 float32
func (this *Packet) ReadFloat32() float32 {
	// 读取数据
	u32 := this.ReadUint32()

	// 数据转化
	f32 := (*float32)(unsafe.Pointer(&u32))

	return *f32
}

// 在 Packet 的 bytes 后面，添加1个 float64 数据
func (this *Packet) AppendFloat64(f float64) {
	// 数据转换
	u64 := (*uint64)(unsafe.Pointer(&f))

	// 添加数据
	this.AppendUint64(*u64)
}

// 从 Packet 的 bytes 中读取1个 float64
func (this *Packet) ReadFloat64() float64 {
	// 读取 uint64
	u64 := this.ReadUint64()

	// 数据转换
	f64 := (*float64)(unsafe.Pointer(&u64))

	return *f64
}

// 在 Packet 的 bytes 后面，添加1个固定大小的 []byte 数据
func (this *Packet) AppendBytes(v []byte) {
	// byte 长度
	bytesLen := uint32(len(v))

	// 申请内存
	this.AllocBuffer(bytesLen)

	// 复制数据
	wPos := this.getWirtePos()
	copy(this.bytes[wPos:wPos+bytesLen], v)

	// 记录长度
	this.addBodyLen(bytesLen)
}

// 从 Packet 的 bytes 中读取1个固定 size 大小的 []byte 数据
//
// size=读取字节数量
func (this *Packet) ReadBytes(size uint32) []byte {
	// 读取位置
	pPos := this.getReadPos()

	// 越界错误
	if pPos > uint32(len(this.bytes)) || (pPos+size) > uint32(len(this.bytes)) {
		zaplog.Panicf("从 Packet 包中读取 Bytes 出错。Bytes 大小超过 packet 剩余可读数据大小")
	}

	// 读取数据
	bytes := this.bytes[pPos : pPos+size]

	// 记录读取数
	this.readCount += size

	return bytes
}

// 在 Packet 的 bytes 后面，添加1个可变大小 []byte 数据
func (this *Packet) AppendVarBytes(v []byte) {
	// 记录 v 长度
	this.AppendUint32(uint32(len(v)))

	// 添加数据
	this.AppendBytes(v)
}

// 从 Packet 的 bytes 中读取1个可变大小 []byte 数据
func (this *Packet) ReadVarBytes() []byte {
	// 读取长度
	ln := this.ReadUint32()

	// 读取 buff
	return this.ReadBytes(ln)
}

// 在 Packet 的 bytes 后面，添加1个 string 数据
func (this *Packet) AppendString(s string) {
	// 数据转换
	bytes := []byte(s)

	// 添加数据
	this.AppendVarBytes(bytes)
}

// 从 Packet 的 bytes 中读取1个 string 数据
func (this *Packet) ReadString() string {
	// 读取 varBytes
	varBytes := this.ReadVarBytes()

	// 数据转化
	s := string(varBytes)

	return s
}

// 添加1个任意数据
func (this *Packet) AppendData(msg interface{}) {

}

// 将1个 Packet包中的数据初始化，并存入 对象池
func (this *Packet) Release() {
	// 后续添加
	refcount := 0

	// 对象池处理
	if 0 == refcount {
		// buff 放回对象池， 并对 Packet 包中 bytes 重新初始化
		payloadLn := this.getPayloadCap() // 有效载荷长度
		if payloadLn > _MIN_PAYLOAD_CAP {
			// 初始化
			buffer := this.bytes
			this.bytes = this.initBytes[:]

			// 放回对象池
			bufferPools[uint32(payloadLn)].Put(buffer)
		}

		// 将 pakcet 放回对象池
		this.readCount = 0
		packetPool.Put(this)
	} else if refcount < 0 {
		// zaplog.Panicf("释放1个 packet 错误，剩余 refcount=%d", p.refcount)
	}
}

// Packet.bytes 中的所有有效数据
func (this *Packet) Data() []byte {
	end := _HEAD_LEN + this.GetBodyLen()

	return this.bytes[0:end]
}

// 根据 need 数量， 为 packet 的 bytes 扩大内存，并完成旧数据复制
func (this *Packet) AllocBuffer(need uint32) {
	// 现有长度满足需求
	newLen := this.GetBodyLen() + need //body 新长度 = 旧长度 + size
	payloadCap := this.getPayloadCap() // 有效容量

	if newLen <= payloadCap {
		return
	}

	// 根据 newLen 大小，从 bufferPools 中获取 buffer 对象池
	poolKey := getPoolKey(newLen)
	newBuffer := bufferPools[poolKey].Get().([]byte)
	if len(newBuffer) != int(poolKey+_HEAD_LEN) {
		zaplog.Panicf("buffer 申请错误，申请的长度=%d,获得的长度=%d", poolKey+_HEAD_LEN, len(newBuffer))
	}

	// 新旧 buff 数据交换
	copy(newBuffer, this.Data())
	oldPayloadLen := this.getPayloadCap()
	oldBytes := this.bytes
	this.bytes = newBuffer

	// 将旧 buffer 放入对象池
	if oldPayloadLen > _MIN_PAYLOAD_CAP {
		pool, ok := bufferPools[oldPayloadLen]

		if ok {
			pool.Put(oldBytes)
		}
	}
}

// 获取 packet 的 bytes 中有效容量（总容量 - 消息头）
func (this *Packet) getPayloadCap() uint32 {
	// bytes 长度
	byteLen := len(this.bytes)
	payloadCap := uint32(byteLen) - _HEAD_LEN

	return payloadCap
}

// 增加 body 长度
func (this *Packet) addBodyLen(ln uint32) {
	pLen := (*uint32)(unsafe.Pointer(&this.bytes[_LEN_POS]))

	*pLen += ln
}

// 设置 body 长度
func (this *Packet) setBodyLen(ln uint32, compressed bool) {
	pBody := (*uint32)(unsafe.Pointer(&this.bytes[_LEN_POS]))

	// 根据压缩计算
	if compressed {

	} else {
		*pBody = ln
	}
}

// 获取读取位置
func (this *Packet) getReadPos() uint32 {
	return _HEAD_LEN + this.readCount
}

// 获取写入位置
func (this *Packet) getWirtePos() uint32 {
	return _HEAD_LEN + this.GetBodyLen()
}
