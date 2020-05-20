// /////////////////////////////////////////////////////////////////////////////
// 通信使用的 pcket 数据包

package network

import (
	"encoding/binary"
	"unsafe"

	"github.com/zpab123/zaplog" // log 工具
)

var (
	// packet 二进制数据操作 （小端）
	packetEndian = binary.LittleEndian
)

// /////////////////////////////////////////////////////////////////////////////
// Packet 对象

// 网络通信二进制数据
type Packet struct {
	mid       uint16                                  // packet 主id
	initBytes [C_PKT_HEAD_LEN + _MIN_PAYLOAD_CAP]byte // bytes 初始化时候的 buffer 4 + 128
	bytes     []byte                                  // 用于存放需要通过网络 发送/接收 的数据 （head + body）
	readCount uint32                                  // bytes 中已经读取的字节数
}

// 创建1个新的 packet 对象
func newPacket() interface{} {
	pkt := Packet{}
	pkt.bytes = pkt.initBytes[:]
	return pkt
}

// 新建1个 Packet 对象 (从对象池创建)
func NewPacket(mid uint16) *Packet {
	pkt := getPacketFromPool()

	pkt.SetMid(mid)

	return pkt
}

// 设置 Packet 的 id
func (this *Packet) SetMid(v uint16) {
	// 记录消息类型
	packetEndian.PutUint16(this.bytes[0:C_PKT_MID_LEN], v)
	this.mid = v
}

// 获取 Packet 的 id
func (this *Packet) GetMid() uint16 {
	return this.mid
}

// 获取 Packet 的 body 部分
func (this *Packet) GetBody() []byte {
	bl := this.GetBodyLen()
	end := (C_PKT_HEAD_LEN + bl)

	return this.bytes[C_PKT_HEAD_LEN:end]
}

// 获取 packet 的 body 字节长度
func (this *Packet) GetBodyLen() uint32 {
	ln := *(*uint32)(unsafe.Pointer(&this.bytes[C_PKT_MID_LEN]))

	return ln
}

// 在 Packet 的 bytes 后面添加1个 byte 数据
func (this *Packet) AppendByte(b byte) {
	// 申请buffer
	this.allocCap(1)

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
	this.allocCap(2)

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
	this.allocCap(4)

	// 赋值
	wPos := this.getWirtePos()
	packetEndian.PutUint32(this.bytes[wPos:wPos+4], v)

	// body 长度+4
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
	this.allocCap(8)

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
	ln := uint32(len(v))

	// 申请内存
	this.allocCap(ln)

	// 复制数据
	wPos := this.getWirtePos()
	copy(this.bytes[wPos:wPos+ln], v)

	// 记录长度
	this.addBodyLen(ln)
}

// 从 Packet 的 bytes 中读取1个固定 size 大小的 []byte 数据
//
// size=读取字节数量
func (this *Packet) ReadBytes(size uint32) []byte {
	// 读取位置
	pPos := this.getReadPos()

	// 越界错误
	if pPos > uint32(len(this.bytes)) || (pPos+size) > uint32(len(this.bytes)) {
		zaplog.Panicf("从 Packet 包中读取 Bytes 出错：Bytes 大小超过 packet 剩余可读数据大小")
	}

	// 读取数据
	bytes := this.bytes[pPos : pPos+size]

	// 记录读取数
	this.readCount += size

	return bytes
}

// 在 Packet 的 bytes 后面，添加1个可变大小 []byte 数据
// 用 uint32 记录 v 的长度
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
// 用 uint16 记录 s 的长度
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
	return string(varBytes)
}

// Packet.bytes 中的所有有效数据
func (this *Packet) Data() []byte {
	end := C_PKT_HEAD_LEN + this.GetBodyLen()

	return this.bytes[0:end]
}

// 根据 need 数量， 为 packet 的 bytes 扩大容量，并完成旧数据复制
func (this *Packet) allocCap(need uint32) {
	// 现有长度满足需求
	newLen := this.GetBodyLen() + need //body 新长度 = 旧长度 + size
	pcap := this.getPayloadCap()       // 有效容量

	if newLen <= pcap {
		return
	}

	// 根据 newLen 大小，从 bufferPools 中获取 buffer 对象池
	poolKey := getPoolKey(newLen)
	newbuf := bufPools[poolKey].Get().([]byte)
	if len(newbuf) != int(poolKey+C_PKT_HEAD_LEN) {
		zaplog.Panicf("buffer 申请错误，申请的长度=%d,获得的长度=%d", poolKey+C_PKT_HEAD_LEN, len(newbuf))
	}

	// 新旧 buff 数据交换
	copy(newbuf, this.Data())
	oldPayloadLen := this.getPayloadCap()
	oldBytes := this.bytes
	this.bytes = newbuf

	// 将旧 buffer 放入对象池
	if oldPayloadLen > _MIN_PAYLOAD_CAP {
		pool, ok := bufPools[oldPayloadLen]
		if ok {
			pool.Put(oldBytes)
		}
	}
}

// 获取 packet 的 bytes 中有效容量（总容量 - 消息头）
func (this *Packet) getPayloadCap() uint32 {
	bl := len(this.bytes)
	pcap := uint32(bl) - C_PKT_HEAD_LEN

	return pcap
}

// 增加 body 长度
func (this *Packet) addBodyLen(ln uint32) {
	bl := (*uint32)(unsafe.Pointer(&this.bytes[C_PKT_MID_LEN]))

	*bl += ln
}

// 设置 body 长度
func (this *Packet) setBodyLen(ln uint32) {
	bl := (*uint32)(unsafe.Pointer(&this.bytes[C_PKT_MID_LEN]))

	*bl = ln
}

// 获取读取位置
func (this *Packet) getReadPos() uint32 {
	return C_PKT_HEAD_LEN + this.readCount
}

// 获取写入位置
func (this *Packet) getWirtePos() uint32 {
	return C_PKT_HEAD_LEN + this.GetBodyLen()
}

// 将1个 Packet包中的数据初始化，并存入 对象池
func (this *Packet) release() {
	// 后续添加
	refcount := 0

	// 对象池处理
	if 0 == refcount {
		// buff 放回对象池， 并对 Packet 包中 bytes 重新初始化
		pc := this.getPayloadCap() // 有效载荷长度
		if pc > _MIN_PAYLOAD_CAP {
			// 初始化
			buf := this.bytes
			this.bytes = this.initBytes[:]

			// 放回对象池
			if pool, ok := bufPools[pc]; ok {
				pool.Put(buf)
			}
		}

		// 将 pakcet 放回对象池
		this.readCount = 0
		this.setBodyLen(0)
		packetPool.Put(this)
	} else if refcount < 0 {
		// zaplog.Panicf("释放1个 packet 错误，剩余 refcount=%d", p.refcount)
	}
}
