// /////////////////////////////////////////////////////////////////////////////
// 面向客户端的 session 组件

package session

import (
	"github.com/zpab123/sco/network" // 网络
	"github.com/zpab123/syncutil"    // 原子变量
)

// /////////////////////////////////////////////////////////////////////////////
// 包初始化

// /////////////////////////////////////////////////////////////////////////////
// ClientSession 对象

// 面向客户端的 session 对象
type ClientSession struct {
	option      *TClientSessionOpt   // 配置参数
	sesssionMgr ISessionManage       // sessiong 管理对象
	sessionId   syncutil.AtomicInt64 // session ID
	session     *Session             // session 对象
	msgHandler  IClientMsgHandler    // 消息处理器
}

// 创建1个新的 ClientSession 对象
func NewClientSession(socket network.ISocket, mgr ISessionManage, opt *TClientSessionOpt) ISession {
	// 创建 ClientSession
	cs := &ClientSession{
		option:      opt,
		sesssionMgr: mgr,
		msgHandler:  opt.MsgHandler,
	}

	// 创建 session
	if opt == nil {
		opt = NewTClientSessionOpt(nil)
	}
	sesOpt := &TSessionOpt{
		Heartbeat:  opt.Heartbeat,
		ScoConnOpt: opt.ScoConnOpt,
		MsgHandler: cs,
	}
	ses := NewSession(socket, sesOpt)

	cs.session = ses

	return cs
}

// 启动 session
func (this *ClientSession) Run() error {
	if this.sesssionMgr != nil {
		// 将 session 添加到管理器, 在线程处理前添加到管理器(分配id), 避免ID还未分配,就开始使用id的竞态问题
		this.sesssionMgr.OnNewSession(this)
	}

	err := this.session.Run()

	return err
}

// 关闭 session
func (this *ClientSession) Stop() error {
	err := this.session.Stop()

	if this.sesssionMgr != nil {
		this.sesssionMgr.OnSessionClose(this)
	}

	return err
}

// 获取 session ID
func (this *ClientSession) GetId() int64 {
	return this.sessionId.Load()
}

// 设置 session ID
func (this *ClientSession) SetId(v int64) {
	this.sessionId.Store(v)
}

// session 消息处理
func (this *ClientSession) OnSessionMessage(ses *Session, packet *network.Packet) {
	if this.msgHandler != nil {
		this.msgHandler.OnClientMessage(this, packet)
	}
}
