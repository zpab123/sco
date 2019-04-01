// /////////////////////////////////////////////////////////////////////////////
// 能够读写 packet 数据的 socket

package network

import (
	"encoding/binary"
	"fmt"
	"net"
	"sync"
	"time"

	"github.com/pkg/errors"           // 错误
	"github.com/zpab123/sco/ioutil"   // io工具
	"github.com/zpab123/sco/protocol" // 通信协议
	"github.com/zpab123/zaplog"       // 日志
)

// /////////////////////////////////////////////////////////////////////////////
// 初始化

// 常量
const (
	_MAX_BODY_LENGTH = C_PKT_MAX_LEN - C_PKT_HEAD_LEN // body 数据最大长度 （ pcket总长度 - 消息头）
)

// 变量
var (
	NETWORK_ENDIAN = binary.LittleEndian // 小端读取对象
	errRecvAgain   = _ErrRecvAgain{}     // 重新接收错误
)

// /////////////////////////////////////////////////////////////////////////////
// PacketSocket 对象

// PacketSocket
type PacketSocket struct {
	socket        ISocket         // 符合 ISocket 的对象
	mutex         sync.Mutex      // 线程互斥锁（发送队列使用）
	cond          *sync.Cond      // 条件同步（发送队列使用）
	sendQueue     []*Packet       // 发送队列
	recvedHeadLen int             // 从 socket 的 readbuffer 中已经读取的 head 数据大小：字节（用于消息读取记录）
	recvedBodyLen int             // 从 socket 的 readbuffer 中已经读取的 body 数据大小：字节（用于消息读取记录）
	headBuff      [_HEAD_LEN]byte // 存放消息头二进制数据
	mid           uint16          // packet id 用于记录消息主id
	bodylen       int             // 本次 pcket body 总大小
	packet        *Packet         // 用于存储本次即将接收的 Packet 对象
}

// 创建1个新的 PacketSocket 对象
func NewPacketSocket(socket ISocket) *PacketSocket {
	pktSocket := &PacketSocket{
		socket: socket,
	}

	pktSocket.cond = sync.NewCond(&pktSocket.mutex)

	return pktSocket
}

// 接收下1个 packet 数据
//
// 返回 nil=没收到完整的 packet 数据; packet=完整的 packet 数据包
func (this *PacketSocket) RecvPacket() (*Packet, error) {
	// 持续接收消息头
	if this.recvedHeadLen < _HEAD_LEN {
		n, err := this.socket.Read(this.headBuff[this.recvedHeadLen:]) // 读取数据
		this.recvedHeadLen += n

		// 消息头不完整
		if this.recvedHeadLen < _HEAD_LEN {
			if nil == err {
				err = errRecvAgain
			}

			return nil, err
		}

		// 收到消息头: 保存本次 packet 消息 id
		this.mid = NETWORK_ENDIAN.Uint16(this.headBuff[0:_LEN_POS])

		// 收到消息头: 保存本次 packet 消息 body 总大小
		bodylen := NETWORK_ENDIAN.Uint32(this.headBuff[_LEN_POS:])
		this.bodylen = int(bodylen)

		// 解密

		// 长度效验
		if bodylen > _MAX_BODY_LENGTH {
			err := errors.Errorf("接收 packet 出错：消息头标记长度=%d，可允许最大长度=%d", bodylen, _MAX_BODY_LENGTH)
			zaplog.Errorf("%s", err)

			this.resetRecvStates()
			this.Close()

			return nil, err
		}

		// 创建新的 packet 对象
		this.recvedBodyLen = 0 // 重置，准备记录 body
		this.packet = NewPacket(this.mid)
		this.packet.allocCap(bodylen)
	}

	// 长度为0类型数据处理
	if this.bodylen == 0 {
		packet := this.packet
		this.resetRecvStates()

		return packet, nil
	}

	// 接收 pcket 数据的 body 部分
	n, err := this.socket.Read(this.packet.bytes[_HEAD_LEN+this.recvedBodyLen : _HEAD_LEN+this.bodylen])
	this.recvedBodyLen += n

	// 接收完成， packet 数据包完整
	if this.recvedBodyLen == this.bodylen {
		// 解密

		// 准备接收下1个
		packet := this.packet
		ln := uint32(this.bodylen)
		packet.setBodyLen(ln, false)

		this.resetRecvStates()

		return packet, nil
	} else if this.recvedBodyLen > this.bodylen {
		err := errors.Errorf("接收 packet 出错：接收长度超过body长度。接收长度=%d，body长度=%d", this.recvedBodyLen, this.bodylen)
		zaplog.Errorf("%s", err)

		this.resetRecvStates()

		return nil, err
	}

	// body 未收完
	if nil == err {
		err = errRecvAgain
	}

	return nil, err
}

// 发送1个 *Packe 数据
func (this *PacketSocket) SendPacket(pkt *Packet) error {
	// 状态效验

	// 添加到消息队列
	this.mutex.Lock()
	this.sendQueue = append(this.sendQueue, pkt)
	this.mutex.Unlock()

	this.cond.Signal()

	return nil
}

// 将消息队列中的数据写入 writebuff
func (this *PacketSocket) Flush() (err error) {
	// 等待数据
	this.mutex.Lock()
	for len(this.sendQueue) == 0 {
		this.cond.Wait()
	}
	this.mutex.Unlock()

	// 复制数据
	this.mutex.Lock()
	packets := make([]*Packet, 0, len(this.sendQueue)) // 复制准备
	packets, this.sendQueue = this.sendQueue, packets  // 交换数据, 并把原来的数据置空
	this.mutex.Unlock()

	// 刷新数据
	if 1 == len(packets) {
		pkt := packets[0]

		// 将 data 写入 conn
		err = ioutil.WriteAll(this.socket, pkt.Data())

		pkt.Release()

		if nil == err {
			err = this.socket.Flush()
		}

		return
	}

	for _, pkt := range packets {
		err = ioutil.WriteAll(this.socket, pkt.Data())

		pkt.Release()
	}

	if nil == err {
		err = this.socket.Flush()
	}

	return
}

// 关闭 socket
func (this *PacketSocket) Close() error {
	return this.socket.Close()
}

// 设置读超时
func (this *PacketSocket) SetRecvDeadline(deadline time.Time) error {
	return this.socket.SetReadDeadline(deadline)
}

// 获取客户端 ip 地址
func (this *PacketSocket) RemoteAddr() net.Addr {
	return this.socket.RemoteAddr()
}

// 获取本地 ip 地址
func (this *PacketSocket) LocalAddr() net.Addr {
	return this.socket.LocalAddr()
}

// 打印信息
func (this *PacketSocket) String() string {
	return fmt.Sprintf("[%s >>> %s]", this.LocalAddr(), this.RemoteAddr())
}

// 重置数据接收状态
func (this *PacketSocket) resetRecvStates() {
	this.recvedHeadLen = 0
	this.recvedBodyLen = 0
	this.mid = protocol.C_MID_INVALID
	this.bodylen = 0
	this.packet = nil
}

// /////////////////////////////////////////////////////////////////////////////
// _ErrRecvAgain 对象

type _ErrRecvAgain struct{}

func (err _ErrRecvAgain) Error() string {
	e := "packet 尚未完整，请继续接收"

	return e
}

func (err _ErrRecvAgain) Temporary() bool {
	return true
}

func (err _ErrRecvAgain) Timeout() bool {
	return true
}
