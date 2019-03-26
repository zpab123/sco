// /////////////////////////////////////////////////////////////////////////////
// 支持格式配置的网络接收器组件

package acceptor

import (
	"sync"
)

// 网络连接对象，支持 websocket tcp
type Acceptor struct {
	cmptName  string         // 组件名字
	stopGroup sync.WaitGroup // 停止等待组
}

// 初始化 Acceptor
func (this *Acceptor) Init() {

	return
}

// 启动 Acceptor
func (this *Acceptor) Run() {

	// 等待停止
	this.stopGroup.Wait()
}

// 停止 Acceptor
func (this *Acceptor) Stop() {

	return
}
