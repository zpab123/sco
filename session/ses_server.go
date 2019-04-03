// /////////////////////////////////////////////////////////////////////////////
// 面向服务器连接的 session 组件

package session

import (
	"github.com/pkg/errors"          // 异常
	"github.com/zpab123/sco/network" // 网络
	"github.com/zpab123/syncutil"    // 原子变量
)

// /////////////////////////////////////////////////////////////////////////////
// 包初始化

// /////////////////////////////////////////////////////////////////////////////
// ServerSession 对象

// 面向服务器连接的 session 对象
type ServerSession struct {
	option      *TServerSessionOpt   // 配置参数
	sesssionMgr ISessionManage       // sessiong 管理对象
	msgHandler  IServerMsgHandler    // 消息处理器
	sessionId   syncutil.AtomicInt64 // session ID
	session     *Session             // session 对象
}

// 创建1个新的 ServerSession 对象
func NewServerSession(socket network.ISocket, mgr ISessionManage, handler IServerMsgHandler, opt *TServerSessionOpt) (ISession, error) {
	var err error

	// 参数效验
	if nil == socket {
		err = errors.New("创建 ServerSession 失败：参数 socket=nil")

		return nil, err
	}

	if nil == mgr {
		err = errors.New("创建 ServerSession 失败：参数 mgr=nil")

		return nil, err
	}

	if nil == handler {
		err = errors.New("创建 ServerSession 失败：参数 handler=nil")

		return nil, err
	}

	// 创建 ServerSession
	ss := &ServerSession{
		option:      opt,
		sesssionMgr: mgr,
		msgHandler:  handler,
	}

	// 创建 session
	if opt == nil {
		opt = NewTServerSessionOpt()
	}

	sesOpt := &TSessionOpt{
		Heartbeat:  opt.Heartbeat,
		ScoConnOpt: opt.ScoConnOpt,
	}

	var ses *Session
	ses, err = NewSession(socket, ss, sesOpt)

	ss.session = ses

	return ss, nil
}

// 启动 session
func (this *ServerSession) Run() (err error) {
	if this.sesssionMgr != nil {
		// 将 session 添加到管理器, 在线程处理前添加到管理器(分配id), 避免ID还未分配,就开始使用id的竞态问题
		this.sesssionMgr.OnNewSession(this)
	}

	err = this.session.Run()

	return
}

// 关闭 session
func (this *ServerSession) Stop() (err error) {
	err = this.session.Stop()

	if this.sesssionMgr != nil {
		this.sesssionMgr.OnSessionClose(this)
	}

	return
}

// 获取 session ID
func (this *ServerSession) GetId() int64 {
	return this.sessionId.Load()
}

// 设置 session ID
func (this *ServerSession) SetId(v int64) {
	this.sessionId.Store(v)
}

// session 消息处理
func (this *ServerSession) OnSessionMessage(ses *Session, packet *network.Packet) {
	if this.msgHandler != nil {
		this.msgHandler.OnServerMessage(this, packet)
	}
}
