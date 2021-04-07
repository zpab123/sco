// /////////////////////////////////////////////////////////////////////////////
// 通信使用的 pcket 数据包

package network

import (
	"encoding/binary"
	"unsafe"

	"github.com/zpab123/sco/log"
)

// 常量 -- packet 数据大小定义
const (
	mincap        int = 512                     // buff 最小有效容量（buff 对象池使用）
	bodyMaxLenInt int = int(C_PKT_BODY_MAX_LEN) // body最大长度（int）
)

var (
	// packet 二进制数据操作 （小端）
	packetEndian = binary.LittleEndian
)

// /////////////////////////////////////////////////////////////////////////////
// Packet 对象

// 网络通信二进制数据
type Packet struct {
	mid      uint16 // packet 主 id
	sid      uint16 // packet 子 id
	bytes    []byte // 用于存放需要通过网络 发送/接收 的数据 （head + body）
	readPos  int    // 读取位置
	wirtePos int    // 写入位置
	agent    *Agent // 代理
}

// 新建1个 Packet 对象 (从对象池创建)
func NewPacket(mid uint16, sid uint16) *Packet {
	pkt := Packet{
		bytes:    make([]byte, headLenInt+mincap),
		readPos:  headLenInt,
		wirtePos: headLenInt,
	}
	pkt.SetMid(mid)
	pkt.SetSid(sid)

	return &pkt
}

// 设置 Packet 的 id
func (this *Packet) SetMid(v uint16) {
	// 记录消息类型
	packetEndian.PutUint16(this.bytes[0:C_PKT_MID_LEN], v)
	this.mid = v
}

// 设置 packet 的 sid
func (this *Packet) SetSid(v uint16) {
	// 记录消息类型
	packetEndian.PutUint16(this.bytes[C_PKT_MID_LEN:C_PKT_MID_LEN+C_PKT_SID_LEN], v)
	this.sid = v
}

// 获取 Packet 的 id
func (this *Packet) GetMid() uint16 {
	return this.mid
}

// 获取 Agent
func (this *Packet) GetAgent() *Agent {
	return this.agent
}

// 获取 Packet 的 body 部分
func (this *Packet) GetBody() []byte {
	return this.bytes[C_PKT_HEAD_LEN:this.wirtePos]
}

// 在 Packet 的 bytes 后面添加1个 byte 数据
func (this *Packet) AppendByte(b byte) {
	// 申请buffer
	this.allocCap(1)

	// 赋值
	this.bytes[this.wirtePos] = b

	// body 长度+1
	this.addBodyLen(1)
}

// 从 Packet 的 bytes 中读取1个 byte 数据，并赋值给 v
func (this *Packet) ReadByte() byte {
	// 赋值
	v := this.bytes[this.readPos]

	// 读取数量+1
	this.readPos += 1

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
	packetEndian.PutUint16(this.bytes[this.wirtePos:this.wirtePos+2], v)

	// body 长度+2
	this.addBodyLen(2)
}

// 从 Packet 的 bytes 中读取1个 uint16 数据，并赋值给v
func (this *Packet) ReadUint16() (v uint16) {
	// 读取
	v = packetEndian.Uint16(this.bytes[this.readPos : this.readPos+2])

	// 读取数量+2
	this.readPos += 2

	return
}

// 在 Packet 的 bytes 后面，添加1个 uint32 数据
func (this *Packet) AppendUint32(v uint32) {
	// 申请buffer
	this.allocCap(4)

	// 赋值
	packetEndian.PutUint32(this.bytes[this.wirtePos:this.wirtePos+4], v)

	// body 长度+4
	this.addBodyLen(4)
}

// 从 Packet 的 bytes 中读取1个 uint32 数据
func (this *Packet) ReadUint32() (v uint32) {
	// 读取
	v = packetEndian.Uint32(this.bytes[this.readPos : this.readPos+4])

	// 读取数量+4
	this.readPos += 4

	return
}

// 在 Packet 的 bytes 后面，添加1个 uint64 数据
func (this *Packet) AppendUint64(v uint64) {
	// 申请内存
	this.allocCap(8)

	// 添加数据
	packetEndian.PutUint64(this.bytes[this.wirtePos:this.wirtePos+8], v)

	// 记录长度
	this.addBodyLen(8)
}

// 从 Packet 的 bytes 中读取1个 uint64 数据
func (this *Packet) ReadUint64() (v uint64) {
	// 读取
	v = packetEndian.Uint64(this.bytes[this.readPos : this.readPos+8])

	// 记录读取数量
	this.readPos += 8

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
	ln := len(v)

	// 申请内存
	this.allocCap(ln)

	// 复制数据
	copy(this.bytes[this.wirtePos:this.wirtePos+ln], v)

	// 记录长度
	this.addBodyLen(uint32(ln))
}

// 从 Packet 的 bytes 中读取1个固定 size 大小的 []byte 数据
//
// size=读取字节数量
func (this *Packet) ReadBytes(size uint32) []byte {
	defer log.Logger.Sync()

	s := int(size)

	// 越界错误
	if this.readPos > len(this.bytes) || (this.readPos+s) > len(this.bytes) {
		log.Logger.Panic(
			"从 Packet 包中读取 Bytes 出错：Bytes 大小超过 packet 剩余可读数据大小",
		)

	}

	// 读取数据
	bytes := this.bytes[this.readPos : this.readPos+s]

	// 记录读取数
	this.readPos += s

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
	return this.bytes[0:this.wirtePos]
}

// 根据 need 数量， 为 packet 的 bytes 扩大容量，并完成旧数据复制
func (this *Packet) allocCap(need int) {
	defer log.Logger.Sync()

	// 超长
	pcap := this.getPayloadCap() // 有效容量
	if pcap >= bodyMaxLenInt {
		return
	}

	// 现有长度满足需求
	newLen := this.wirtePos + need //body 新长度 = 旧长度 + size
	if newLen <= pcap {
		return
	}

	// 创建新的 []byte
	nb := (newLen + mincap)
	if nb > bodyMaxLenInt {
		nb = bodyMaxLenInt
	}

	if nb > pcap {
		b := make([]byte, headLenInt+nb)
		copy(b, this.Data())
	} else {
		log.Logger.Warn(
			"[Packet] 容量达到最大",
			log.Int64("容量=", int64(bodyMaxLenInt)),
		)
	}
}

// 获取 packet 的 bytes 中有效容量（总容量 - 消息头）
func (this *Packet) getPayloadCap() int {
	return len(this.bytes) - headLenInt
}

// 增加 body 长度
func (this *Packet) addBodyLen(ln uint32) {
	bl := (*uint32)(unsafe.Pointer(&this.bytes[C_PKT_MID_LEN]))

	*bl += ln
	this.wirtePos += int(ln)
}
