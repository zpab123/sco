// /////////////////////////////////////////////////////////////////////////////
// 面向服务器连接的 session 组件

package session

import (
	"time"

	"github.com/pkg/errors"            // 异常库
	"github.com/zpab123/world/network" // 网络库
	"github.com/zpab123/world/state"   // 状态管理
	"github.com/zpab123/world/wderr"   // 异常库
	"github.com/zpab123/zaplog"        // 日志库
)

// /////////////////////////////////////////////////////////////////////////////
// 包初始化

// /////////////////////////////////////////////////////////////////////////////
// Session 对象

// 面向服务器连接的 session 对象
type Session struct {
	option       *TSessionOpt             // 配置参数
	stateMgr     *state.StateManager      // 状态管理
	worldConn    *network.WorldConnection // world 引擎连接对象
	msgHandler   ISessionMsgHandler       // 消息处理器
	ticker       *time.Ticker             // 心跳计时器
	timeOut      time.Duration            // 心跳超时时间
	lastRecvTime time.Time                // 上次接收消息的时间
	lastSendTime time.Time                // 上次发送消息的时间
}

// 创建1个新的 Session 对象
func NewSession(socket network.ISocket, opt *TSessionOpt) *Session {
	// 创建 StateManager
	st := state.NewStateManager()

	// 创建 WorldConnection
	if nil == opt {
		opt = NewTSessionOpt(nil)
	}
	wc := network.NewWorldConnection(socket, opt.WorldConnOpt)

	// 创建对象
	ss := &Session{
		option:     opt,
		stateMgr:   st,
		worldConn:  wc,
		msgHandler: opt.MsgHandler,
		timeOut:    opt.Heartbeat * 2,
	}

	// 修改为初始化状态
	ss.stateMgr.SetState(state.C_INIT)

	return ss
}

// 启动 session [ISession 接口]
func (this *Session) Run() (err error) {
	// 状态效验
	if !this.stateMgr.SwapState(state.C_INIT, state.C_RUNING) {
		if !this.stateMgr.SwapState(state.C_STOPED, state.C_RUNING) {
			err = errors.Errorf("Session 启动失败，状态错误。当前状态=%d，正确状态=%d或%d", this.stateMgr.GetState(), state.C_INIT, state.C_STOPED)

			return
		}
	}
	// 变量重置？ 状态? 发送队列？

	// 开启接收 goroutine
	go this.recvLoop()

	// 开启发送 goroutine
	go this.sendLoop()

	// 计时器 goroutine
	if this.timeOut > 0 {
		this.ticker = time.NewTicker(this.timeOut)
		go this.mainLoop()
	}

	// 改变状态： 工作中
	if !this.stateMgr.SwapState(state.C_RUNING, state.C_WORKING) {
		err = errors.Errorf("Session 启动失败，状态错误。当前状态=%d，正确状态=%d", this.stateMgr.GetState(), state.C_RUNING)

		return
	}

	return
}

// 关闭 session [ISession 接口]
func (this *Session) Stop() (err error) {
	// 状态改变为关闭中
	if !this.stateMgr.SwapState(state.C_WORKING, state.C_CLOSEING) {
		err = errors.Errorf("Session %s 关闭失败，状态错误。当前状态=%d, 正确状态=%d", this, this.stateMgr.GetState(), state.C_WORKING)

		return
	}

	// 关闭连接
	err = this.worldConn.Close()
	if nil != err {
		err = errors.Errorf("Session %s 关闭失败。错误=%s", this, err)

		return
	}

	// 状态改变为关闭完成
	if !this.stateMgr.SwapState(state.C_CLOSEING, state.C_CLOSED) {
		err = errors.Errorf("Session %s 关闭失败，状态错误。当前状态=%d, 正确状态=%d", this, this.stateMgr.GetState(), state.C_CLOSEING)

		return
	}

	return
}

// 打印信息
func (this *Session) String() string {
	return this.worldConn.String()
}

// 发送心跳消息
func (this *Session) SendHeartbeat() {
	this.worldConn.SendHeartbeat()
}

// 发送通用消息
func (this *Session) SendData(data []byte) {
	this.worldConn.SendData(data)
}

// 接收线程
func (this *Session) recvLoop() {
	defer func() {
		this.Stop()

		if err := recover(); nil != err && !wderr.IsConnectionError(err.(error)) {
			zaplog.TraceError("Session %s 接收数据出现错误：%s", this, err.(error))
		} else {
			zaplog.Debugf("Session %s 断开连接", this)
		}
	}()

	// 这里有bug 不应该在这里监测状态
	for this.stateMgr.GetState() == state.C_WORKING {
		// 接收消息
		pkt, err := this.worldConn.RecvPacket()

		// 错误处理
		if nil != err && !wderr.IsTimeoutError(err) {
			if wderr.IsConnectionError(err) {
				break
			} else {
				panic(err)
			}
		}

		// 消息处理
		if nil == pkt {
			continue
		}

		if this.msgHandler != nil {
			this.msgHandler.OnSessionMessage(this, pkt) // 这里还需要增加异常处理
		}

	}
}

// 发送线程
func (this *Session) sendLoop() {
	var err error

	for this.stateMgr.GetState() == state.C_WORKING {
		err = this.worldConn.Flush() // 刷新缓冲区

		if nil != err {
			break
		}
	}
}

// 主循环
func (this *Session) mainLoop() {
	for {
		select {
		case <-this.ticker.C:
			this.checkRecvTime() // 检查接收是否超时
			this.checkSendTime() // 检查发送是否超时
		}
	}

}

// 检查接收是否超时
func (this *Session) checkRecvTime() {
	if time.Now().After(this.lastRecvTime.Add(this.timeOut)) {
		zaplog.Errorf("Session %s 接收消息超时，关闭连接", this)

		this.Stop()
	}
}

// 检查发送是否超时
func (this *Session) checkSendTime() {
	if time.Now().After(this.lastSendTime.Add(this.timeOut)) {
		zaplog.Debugf("Session %s 发送心跳", this)

		this.SendHeartbeat()
	}
}
