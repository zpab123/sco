// /////////////////////////////////////////////////////////////////////////////
// 客户端接收器

package network

import (
	"sync"

	"github.com/zpab123/syncutil" // 原子操作工具
	"github.com/zpab123/zaplog"   // log 日志库
	"golang.org/x/net/websocket"  // websocket
)

// /////////////////////////////////////////////////////////////////////////////
// ClientAcceptor

// 客户端接收器
type ClientAcceptor struct {
	acceptors []IAcceptor           // 接收器切片
	stopGroup sync.WaitGroup        // 停止等待组
	option    *TClientAcceptorOpt   // 配置参数
	connNum   syncutil.AtomicUint32 // 当前连接数
}

// 新建1个客户端接收器
func NewClientAcceptor() *ClientAcceptor {
	acc := ClientAcceptor{
		acceptors: make([]IAcceptor, 0),
	}

	return &acc
}

// 添加1个接收器
func (this *ClientAcceptor) Add(acc IAcceptor) {
	if nil != acc {
		append(this.acceptors, acc)
	}
}

// 启动
func (this *ClientAcceptor) Run() {
	if len(this.acceptors) <= 0 {
		return
	}

	for _, acc := range this.acceptors {
		err := acc.Run()
		if nil == err {
			this.stopGroup.Add(1)
		}
	}
}

// 停止
func (this *ClientAcceptor) Stop() {
	if len(this.acceptors) <= 0 {
		return
	}

	for _, acc := range this.acceptors {
		err := acc.Stop()
		if nil == err {
			this.stopGroup.Add(-1)
		}
	}

	this.stopGroup.Wait()

	zaplog.Debugf("ClientAcceptor 停止")
}

// 收到1个新的 websocket 连接对象 [IWsConnManager]
func (this *ClientAcceptor) OnNewWsConn(wsconn *websocket.Conn) {
	// 超过最大连接数
	if this.connNum.Load() >= this.option.MaxConn {
		wsconn.Close()

		zaplog.Warnf("NetService 达到最大连接数，关闭新连接。当前连接数=%d", this.connNum.Load())
	}

	// 参数设置
	wsconn.PayloadType = websocket.BinaryFrame // 以二进制方式接受数据

	// 创建 session 对象
	// this.createSession(wsconn, true)
}
