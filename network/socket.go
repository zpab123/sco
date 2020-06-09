// /////////////////////////////////////////////////////////////////////////////
// 基础 Socket 对象

package network

import (
	"encoding/binary"
	"fmt"
	"io"
	"net"
	"sync"

	"github.com/pkg/errors"
	"github.com/zpab123/sco/iotool"
)

// /////////////////////////////////////////////////////////////////////////////
// 初始化

// 变量
var (
	socketEndian       = binary.LittleEndian   // 小端读取对象
	headLenInt   int   = int(C_PKT_HEAD_LEN)   // 消息头长度(int)
	headLenInt64 int64 = int64(C_PKT_HEAD_LEN) // 消息头长度(int64)
	errClose           = errors.New("socket 关闭")
)

// /////////////////////////////////////////////////////////////////////////////
// Socket 对象

// Socket
type Socket struct {
	conn      net.Conn   // 接口继承： 符合 Conn 接口的对象
	sendQueue [][]byte   // 发送队列
	mutex     sync.Mutex // 线程互斥锁（发送队列使用）
	cond      *sync.Cond // 条件同步（发送队列使用）
	head      []byte     // 消息头
}

// 创建1个 *Socket
// 成功：返回 *Socket
// 失败：返回 nil
func NewSocket(conn net.Conn) *Socket {
	// 参数效验
	if nil == conn {
		return nil
	}

	// 创建对象
	sq := make([][]byte, 0)

	s := Socket{
		conn:      conn,
		sendQueue: sq,
		head:      make([]byte, C_PKT_HEAD_LEN),
	}
	s.cond = sync.NewCond(&s.mutex)

	return &s
}

// 关闭 socket
// 成功，返回 nil
// 失败，返回 error
func (this *Socket) Close() error {
	return this.conn.Close()
}

// 接收下1个 packet 数据
//
// 成功，返回 *Packet nil
// 失败，返回 nil error
func (this *Socket) RecvPacket() (*Packet, error) {
	_, err := io.ReadFull(this.conn, this.head)
	if err != nil {
		return nil, err
	}

	mid := socketEndian.Uint16(this.head[0:C_PKT_MID_LEN])
	bl := socketEndian.Uint32(this.head[C_PKT_MID_LEN:])
	if bl > C_PKT_BODY_MAX_LEN {
		return nil, V_ERR_BODY_LEN
	}

	body := make([]byte, bl)
	_, err = io.ReadFull(this.conn, body)
	if err != nil {
		return nil, err
	}

	pkt := NewPacket(mid)
	pkt.AppendBytes(body)

	return pkt, nil
}

// 发送1个 *Packe 数据
func (this *Socket) SendPacket(pkt *Packet) {
	var bytes []byte
	if nil != pkt {
		bytes = pkt.Data()
	}

	// 添加到消息队列
	this.mutex.Lock()
	this.sendQueue = append(this.sendQueue, bytes)
	this.mutex.Unlock()

	this.cond.Signal()
}

// 发送 []byte
func (this *Socket) SendBytes(bytes []byte) {
	// 添加到消息队列
	this.mutex.Lock()
	this.sendQueue = append(this.sendQueue, bytes)
	this.mutex.Unlock()

	this.cond.Signal()
}

// 将消息队列中的数据写入缓冲
// 成功，返回 nil
// 失败，返回 error
func (this *Socket) Flush() error {
	// 等待数据
	this.mutex.Lock()
	for len(this.sendQueue) == 0 {
		this.cond.Wait()
	}
	this.mutex.Unlock()

	// 复制数据
	this.mutex.Lock()
	newsq := make([][]byte, 0, len(this.sendQueue))
	newsq, this.sendQueue = this.sendQueue, newsq // 交换数据
	this.mutex.Unlock()

	// 写入数据
	for _, bytes := range newsq {
		if nil != bytes {
			err := iotool.WriteAll(this.conn, bytes)
			if nil != err {
				return err
			}
		} else {
			return errClose
		}
	}

	// 编码数据
	return nil
}

// 获取客户端 ip 地址
func (this *Socket) RemoteAddr() net.Addr {
	return this.conn.RemoteAddr()
}

// 获取本地 ip 地址
func (this *Socket) LocalAddr() net.Addr {
	return this.conn.LocalAddr()
}

// 打印信息
func (this *Socket) String() string {
	return fmt.Sprintf("[%s >>> %s]", this.LocalAddr(), this.RemoteAddr())
}
