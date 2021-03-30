// /////////////////////////////////////////////////////////////////////////////
// 连接管理

package network

import (
	"net"
	"sync"
	"time"

	"github.com/zpab123/sco/log"
	"github.com/zpab123/sco/syncutil"
	"golang.org/x/net/websocket"
)

// /////////////////////////////////////////////////////////////////////////////
// ConnMgr

// 连接管理
type ConnMgr struct {
	maxConn   int32                // 最大连接数量，超过此数值后，不再接收新连接
	key       string               // 握手key
	heartbeat time.Duration        // 心跳周期
	handler   IHandler             // 消息处理器
	connNum   syncutil.AtomicInt32 // 当前连接数
	agentMap  sync.Map             // agent 集合
	agentId   syncutil.AtomicInt32 // agent id
}

// 新建1个 ConnMgr
func NewConnMgr(max int32) IConnManager {
	if max <= 0 {
		max = C_F_MAX_CONN
	}

	mgr := ConnMgr{
		maxConn:   max,
		key:       C_F_KEY,
		heartbeat: C_F_HEARTBEAT,
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

// 设置握手 key
func (this *ConnMgr) SetKey(k string) {
	if "" != k {
		this.key = k
	}
}

// 设置心跳 key
func (this *ConnMgr) SetHeartbeat(h time.Duration) {
	this.heartbeat = h
}

// 设置 handler
func (this *ConnMgr) SetHandler(h IHandler) {
	if nil != h {
		this.handler = h
	}
}

// 收到1个新的 websocket 连接对象 [IWsConnManager]
func (this *ConnMgr) OnWsConn(wsconn *websocket.Conn) {
	defer log.Logger.Sync()

	// 参数效验
	if nil == wsconn {
		return
	}

	// 超过最大连接数
	if this.connNum.Load() >= this.maxConn {
		wsconn.Close()
		log.Logger.Warn(
			"[ConnMgr] 达到最大连接数，关闭新连接",
			log.Int32("当前连接数=", this.connNum.Load()),
		)
	}

	// 参数设置
	wsconn.PayloadType = websocket.BinaryFrame // 以二进制方式接受数据

	// 创建代理
	log.Logger.Debug(
		"[ConnMgr] 新 ws 连接",
		log.String("ip=", wsconn.RemoteAddr().String()),
	)

	this.newAgent(wsconn)
}

// 收到1个新的 tcp 连接对象 [ITcpConnManager]
func (this *ConnMgr) OnTcpConn(conn net.Conn) {
	defer log.Logger.Sync()

	// 参数效验
	if nil == conn {
		return
	}

	// 超过最大连接数
	if this.connNum.Load() >= this.maxConn {
		conn.Close()

		log.Logger.Warn(
			"[ConnMgr] 达到最大连接数，关闭新连接",
			log.Int32("当前连接数=", this.connNum.Load()),
		)
	}

	// 创建代理
	log.Logger.Debug(
		"[ConnMgr] 新 tcp 连接",
		log.String("ip=", conn.RemoteAddr().String()),
	)

	this.newAgent(conn)
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

		log.Logger.Debug(
			"[ConnMgr] Agent 断开",
			log.Int32("当前连接数=", this.connNum.Load()),
		)

	}
}

// 创建代理
func (this *ConnMgr) newAgent(conn net.Conn) {
	s, err := NewSocket(conn)
	if nil != err {
		return
	}

	a, err := NewAgent(s)
	if nil != err {
		return
	}

	a.SetKey(this.key)
	a.SetHeartbeat(this.heartbeat)
	a.SetHandler(this.handler)
	id := this.agentId.Add(1)
	a.SetId(id)
	a.SetConnMgr(this)

	this.agentMap.Store(id, a)
	this.connNum.Add(1)
	a.Run()
}
