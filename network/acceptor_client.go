// /////////////////////////////////////////////////////////////////////////////
// 客户端接收器

package network

import (
	"sync"

	"github.com/zpab123/zaplog" // log 日志库
)

// /////////////////////////////////////////////////////////////////////////////
// ClientAcceptor

// 客户端接收器
type ClientAcceptor struct {
	acceptors []IAcceptor    // 接收器切片
	stopGroup sync.WaitGroup // 停止等待组
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
