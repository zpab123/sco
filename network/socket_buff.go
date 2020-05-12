// /////////////////////////////////////////////////////////////////////////////
// 有 读/写io缓冲 的 socket 对象

package network

import (
	"bufio"
)

// /////////////////////////////////////////////////////////////////////////////
// BufferSocket 对象

// BufferSocket
type BufferSocket struct {
	ISocket                 // 接口继承： 符合 ISocket 的对象
	bufReader *bufio.Reader // 读取类 buffer
	bufWriter *bufio.Writer // 写入类 buffer
}

// 创建1个新的 BufferSocket 对象
func NewBufferSocket(socket ISocket, opts *TBufferSocketOpt) *BufferSocket {
	// 参数效验
	if nil == opts {
		opts = NewTBufferSocketOpt()
	}

	// 创建对象
	bufsocket := BufferSocket{
		ISocket: socket,
	}

	// 创建 buffer
	bufsocket.bufReader = bufio.NewReaderSize(socket, opts.ReadBufferSize)
	bufsocket.bufWriter = bufio.NewWriterSize(socket, opts.WriteBufferSize)

	return &bufsocket
}

// 将 bufReader 中部分读取数据到 p 中
func (this *BufferSocket) Read(p []byte) (int, error) {
	return this.bufReader.Read(p)
}

// 将 p 中部分数据写入 buffer 中
func (this *BufferSocket) Write(p []byte) (int, error) {
	return this.bufWriter.Write(p)
}

// 刷新写入类 buffer 缓冲
func (this *BufferSocket) Flush() error {
	// 刷新 bufWriter
	err := this.bufWriter.Flush()
	if nil != err {
		return err
	}

	// 刷新 socket
	return this.ISocket.Flush()
}
