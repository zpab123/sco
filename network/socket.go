// /////////////////////////////////////////////////////////////////////////////
// 基础 Socket 对象

package network

import (
	"encoding/binary"
	"io"
	"io/ioutil"
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
	sendQueue []*Packet  // 发送队列
	mutex     sync.Mutex // 线程互斥锁（发送队列使用）
	cond      *sync.Cond // 条件同步（发送队列使用）
}

// 创建1个 *Socket
// 成功：返回 *Socket
// 失败：返回 nil
func NewSocket(conn net.Conn) *Socket {
	sq := make([]*Packet, 0)

	s := Socket{
		conn:      conn,
		sendQueue: sq,
	}
	s.cond = sync.NewCond(&s.mutex)

	return &s
}

// 关闭
func (this *Socket) Close() error {
	return this.conn.Close()
}

// 接收下1个 packet 数据
//
// 成功，返回 *Packet nil
// 失败，返回 nil error
func (this *Socket) RecvPacket() (*Packet, error) {
	head, err := ioutil.ReadAll(io.LimitReader(this.conn, headLenInt64))
	if err != nil {
		return nil, err
	}

	mid := socketEndian.Uint16(head[0:C_PKT_MID_LEN])
	bl := socketEndian.Uint32(head[C_PKT_MID_LEN:])
	if bl > C_PKT_BODY_MAX_LEN {
		return nil, V_ERR_BODY_LEN
	}

	body, err := ioutil.ReadAll(io.LimitReader(this.conn, int64(bl)))
	if err != nil {
		return nil, err
	}

	data := append(head, body...)
	pkt := NewPacket(mid, data)

	return pkt, nil
}

// 发送1个 *Packe 数据
func (this *Socket) SendPacket(pkt *Packet) error {
	// 添加到消息队列
	this.mutex.Lock()
	this.sendQueue = append(this.sendQueue, pkt)
	this.mutex.Unlock()

	this.cond.Signal()

	return nil
}

// 将消息队列中的数据写入缓冲
func (this *Socket) Flush() error {
	// 等待数据
	this.mutex.Lock()
	for len(this.sendQueue) == 0 {
		this.cond.Wait()
	}
	this.mutex.Unlock()

	// 复制数据
	this.mutex.Lock()
	packets := make([]*Packet, 0, len(this.sendQueue))
	packets, this.sendQueue = this.sendQueue, packets // 交换数据
	this.mutex.Unlock()

	// 写入数据
	for _, pkt := range packets {
		if nil != pkt {
			err := iotool.WriteAll(this.conn, pkt.Data())
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
