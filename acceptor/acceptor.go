// /////////////////////////////////////////////////////////////////////////////
// 支持格式配置的网络接收器组件

package acceptor

import (
	"sync"

	"github.com/zpab123/sco/model"   // 全局模型
	"github.com/zpab123/sco/network" // 网络
	"golang.org/x/net/websocket"     // websocket
)

// 网络连接对象，支持 websocket tcp
type Acceptor struct {
	cmptName  string            // 组件名字
	stopGroup sync.WaitGroup    // 停止等待组
	acceptor  network.IAcceptor // acceptor 连接器
}

// 新建1个 Acceptor 对象
func NewAcceptor(addr *network.TLaddr, opt *TAcceptorOpt) (model.IComponent, error) {
	var err error
	var a network.IAcceptor

	// 创建 Acceptor
	actor := &Acceptor{
		cmptName: C_CMPT_NAME,
	}

	// 创建 Acceptor
	a, err = network.NewAcceptor(opt.AcceptorName, addr, actor)
	if nil != err {
		return nil, err
	} else {
		actor.acceptor = a
	}

	return actor, err
}

// 启动 Acceptor
func (this *Acceptor) Run() {
	this.stopGroup.Add(1)

	go func() {
		defer this.stopGroup.Done()

		this.acceptor.Run()
	}()

	// 阻塞-等待停止
	this.stopGroup.Wait()
}

// 停止 Acceptor
func (this *Acceptor) Stop() {
	this.acceptor.Stop()
}

// 获取组件名字
func (this *Acceptor) Name() string {
	return this.cmptName
}

// 收到1个新的 websocket 连接对象
func (this *Acceptor) OnNewWsConn(wsconn *websocket.Conn) {

}
