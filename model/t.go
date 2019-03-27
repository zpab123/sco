// /////////////////////////////////////////////////////////////////////////////
// 全局 types

package model

import (
	"time"
)

// /////////////////////////////////////////////////////////////////////////////
// TTcpConnOpt 对象

// TcpSocket 配置参数
type TTcpConnOpt struct {
	ReadBufferSize  int           // 读取 buffer 字节大小
	WriteBufferSize int           // 写入 buffer 字节大小
	NoDelay         bool          // 写入数据后，是否立即发送
	MaxPacketSize   int           // 单个 packet 最大字节数
	ReadTimeout     time.Duration // 读数据超时时间
	WriteTimeout    time.Duration // 写数据超时时间
}

// 创建1个新的 TTcpConnOpt 对象
func NewTTcpConnOpt() *TTcpConnOpt {
	// 创建对象
	tcpOpt := &TTcpConnOpt{
		ReadBufferSize:  C_TCP_BUFFER_READ_SIZE,  // 张鹏：原先是-1，这里被修改了
		WriteBufferSize: C_TCP_BUFFER_WRITE_SIZE, // 张鹏：原先是-1，这里被修改了
		NoDelay:         C_TCP_NO_DELAY,          // 张鹏：原先没有这个设置项，这里被修改了
	}

	return tcpOpt
}
