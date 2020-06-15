// /////////////////////////////////////////////////////////////////////////////
// 连接管理

package network

import (
	"net"
	"sync"

	"github.com/zpab123/syncutil"
	"github.com/zpab123/zaplog"
	"golang.org/x/net/websocket"
)

// /////////////////////////////////////////////////////////////////////////////
// ConnMgr

// 连接管理
type ConnMgr struct {
	maxConn  int32                // 最大连接数量，超过此数值后，不再接收新连接
	connNum  syncutil.AtomicInt32 // 当前连接数
	agentMap sync.Map             // agent 集合
	agentId  syncutil.AtomicInt32 // agent id
	handler  IHandler             // 消息处理器
}

// 新建1个 ConnMgr
func NewConnMgr(max int32) IConnManager {
	if max <= 0 {
		max = C_CONN_MAX
	}

	mgr := ConnMgr{
		maxConn: max,
	}

	return &mgr
}

// 停止连接管理
func (this *ConnMgr) Stop() {
	this.agentMap.Range(func(key, v interface{}) bool {
		if a, ok := v.(*Agent); ok {
			a.Stop()
		}

		this.agentMap.Delete(key)
		return true
	})
}

// 收到1个新的 websocket 连接对象 [IWsConnManager]
func (this *ConnMgr) OnWsConn(wsconn *websocket.Conn) {
	// 超过最大连接数
	if this.connNum.Load() >= this.maxConn {
		wsconn.Close()
		zaplog.Warnf("[ConnMgr] 达到最大连接数，关闭新连接。当前连接数=%d", this.connNum.Load())
	}

	// 参数设置
	wsconn.PayloadType = websocket.BinaryFrame // 以二进制方式接受数据

	// 创建代理
	zaplog.Debugf("[ConnMgr] 新 ws 连接，ip=%s", wsconn.RemoteAddr())
	this.newAgent(wsconn, true)
}

// 收到1个新的 tcp 连接对象 [ITcpConnManager]
func (this *ConnMgr) OnTcpConn(conn net.Conn) {
	// 超过最大连接数
	if this.connNum.Load() >= this.maxConn {
		conn.Close()
		zaplog.Warnf("[ConnMgr] 达到最大连接数，关闭新连接。当前连接数=%d", this.connNum.Load())
	}

	// 创建代理
	zaplog.Debugf("[ConnMgr] 新 tcp 连接，ip=%s", conn.RemoteAddr())
	this.newAgent(conn, true)
}

// 某个 Agent 停止
func (this *ConnMgr) OnAgentStop(a *Agent) {
	if nil == a {
		return
	}

	id := a.GetId()
	if _, ok := this.agentMap.Load(id); ok {
		this.agentMap.Delete(id)
		this.connNum.Add(-1)
		zaplog.Debugf("[ConnMgr] Agent 断开，当前连接数=%d", this.connNum.Load())
	}
}

// 设置 handler
func (this *ConnMgr) SetHandler(h IHandler) {
	if nil != h {
		this.handler = h
	}
}

// 创建代理
func (this *ConnMgr) newAgent(conn net.Conn, isWebSocket bool) {
	ao := NewAgentOpt()
	s := NewSocket(conn)

	a, err := NewAgent(s, ao)
	if nil != err {
		return
	}

	a.SetHandler(this.handler)
	id := this.agentId.Add(1)
	a.SetId(id)
	a.SetConnMgr(this)

	this.agentMap.Store(id, a)
	this.connNum.Add(1)
	a.Run()
}
