// /////////////////////////////////////////////////////////////////////////////
// 客户端接收器

package network

import (
	"net"
	"sync"

	"github.com/pkg/errors"       // 异常库
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
	options   *TClientAcceptorOpt   // 配置参数
	connNum   syncutil.AtomicUint32 // 当前连接数
}

// 新建1个客户端接收器
func NewClientAcceptor(opt *TClientAcceptorOpt) *ClientAcceptor {
	if nil == opt {
		opt = NewTClientAcceptorOpt()
	}

	acc := ClientAcceptor{
		acceptors: make([]IAcceptor, 0),
		options:   opt,
	}

	acc.init()

	return &acc
}

// 添加1个接收器
func (this *ClientAcceptor) Add(acc IAcceptor) {
	if nil != acc {
		this.acceptors = append(this.acceptors, acc)
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
func (this *ClientAcceptor) Stop() error {
	var err error
	var accerr error
	if len(this.acceptors) <= 0 {
		return nil
	}

	for _, acc := range this.acceptors {
		accerr = acc.Stop()
		if nil != accerr && nil == err {
			err = errors.New("ClientAcceptor 停止过程中出现错误")
		}

		this.stopGroup.Done()
	}

	this.stopGroup.Wait()

	zaplog.Debugf("ClientAcceptor 停止")

	return err
}

// 收到1个新的 websocket 连接对象 [IWsConnManager]
func (this *ClientAcceptor) OnNewWsConn(wsconn *websocket.Conn) {
	// 超过最大连接数
	if this.connNum.Load() >= this.options.MaxConn {
		wsconn.Close()

		zaplog.Warnf("NetService 达到最大连接数，关闭新连接。当前连接数=%d", this.connNum.Load())
	}

	// 参数设置
	wsconn.PayloadType = websocket.BinaryFrame // 以二进制方式接受数据

	// 创建 session 对象
	// this.createSession(wsconn, true)
	// 创建代理
	zaplog.Debugf("新连接,ip=%s", wsconn.RemoteAddr())
	this.newAgent(wsconn, true)
}

// 初始化
func (this *ClientAcceptor) init() {
	// var err error
	if this.options.WsAddr != "" {
		ws, err := NewWsAcceptor(this.options.WsAddr, this)
		if err == nil {
			this.acceptors = append(this.acceptors, ws)
		}
	}
}

// 创建代理
func (this *ClientAcceptor) newAgent(netconn net.Conn, isWebSocket bool) {
	// 创建 socket
	socket := Socket{
		Conn: netconn,
	}

	opt := NewTAgentOpt()
	opt.Handler = this.options.Handler

	a, err := NewAgent(socket, opt)
	if nil != err {
		zaplog.Error("创建 Agent 失败..")
		return
	}

	a.Run()

	this.connNum.Add(1)
}
