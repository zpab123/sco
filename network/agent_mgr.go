// /////////////////////////////////////////////////////////////////////////////
// agent 管理

package network

import (
	"net"
	"sync"
	"time"

	"github.com/zpab123/sco/log"
	"github.com/zpab123/sco/syncs"
	"golang.org/x/net/websocket"
)

// /////////////////////////////////////////////////////////////////////////////
// AgentMgr

// 连接管理
type AgentMgr struct {
	maxConn    int32             // 最大连接数量，超过此数值后，不再接收新连接
	key        string            // 握手key
	heartbeat  time.Duration     // 心跳周期
	heartSend  int64             // 心跳-发送(纳秒)
	heartRecv  int64             // 心跳-接受(纳秒)
	handler    IHandler          // 消息处理器
	connNum    syncs.AtomicInt32 // 当前连接数
	agentMap   sync.Map          // agent 集合
	agentId    syncs.AtomicInt32 // agent id
	packetChan chan *Packet      // 消息通道
	stoping    bool              // 正在停止中
	chDie      chan struct{}     // 关闭通道
	clientPkt  chan *Packet      // client 消息
	serverPkt  chan *Packet      // server 消息
}

// 新建1个 AgentMgr
func NewAgentMgr(max int32) IAgentManager {
	if max <= 0 {
		max = C_F_MAX_CONN
	}

	mgr := AgentMgr{
		maxConn:   max,
		key:       C_F_KEY,
		heartbeat: C_F_HEARTBEAT,
		heartSend: int64(C_F_HEARTBEAT) / 2,
		heartRecv: int64(C_F_HEARTBEAT),
		chDie:     make(chan struct{}),
	}

	return &mgr
}

// -----------------------------------------------------------------------------
// ITcpConnManager 接口

// 收到1个新的 tcp 连接对象
func (this *AgentMgr) OnTcpConn(conn net.Conn) {
	//defer log.Logger.Sync()

	// 参数效验
	if nil == conn {
		return
	}

	// 超过最大连接数
	if this.connNum.Load() >= this.maxConn {
		conn.Close()

		log.Logger.Warn(
			"[AgentMgr] 达到最大连接数，关闭新连接",
			log.Int32("当前连接数=", this.connNum.Load()),
		)
	}

	// 创建代理
	log.Logger.Debug(
		"[AgentMgr] 新 tcp 连接",
		log.String("ip=", conn.RemoteAddr().String()),
	)

	this.newAgent(conn)
}

// -----------------------------------------------------------------------------
// IWsConnManager 接口

// 收到1个新的 websocket 连接对象
func (this *AgentMgr) OnWsConn(wsconn *websocket.Conn) {
	defer log.Logger.Sync()

	// 参数效验
	if nil == wsconn {
		return
	}

	// 超过最大连接数
	if this.connNum.Load() >= this.maxConn {
		wsconn.Close()
		log.Logger.Warn(
			"[AgentMgr] 达到最大连接数，关闭新连接",
			log.Int32("当前连接数=", this.connNum.Load()),
		)
	}

	// 参数设置
	wsconn.PayloadType = websocket.BinaryFrame // 以二进制方式接受数据

	// 创建代理
	log.Logger.Debug(
		"[AgentMgr] 新 ws 连接",
		log.String("ip=", wsconn.RemoteAddr().String()),
	)

	this.newAgent(wsconn)
}

// -----------------------------------------------------------------------------
// IAgentManager 接口

// 设置消息通道
func (this *AgentMgr) SetClientPacketChan(ch chan *Packet) {
	if ch != nil {
		this.clientPkt = ch
	}
}

// 设置 server 消息通道
func (this *AgentMgr) SetServerPacketChan(ch chan *Packet) {
	if ch != nil {
		this.serverPkt = ch
	}
}

// 某个 Agent 停止
func (this *AgentMgr) OnAgentStop(a *Agent) {
	if nil == a {
		return
	}

	if this.stoping {
		return
	}

	id := a.GetId()
	if _, ok := this.agentMap.Load(id); ok {
		this.agentMap.Delete(id)
		this.connNum.Add(-1)

		log.Logger.Debug(
			"[AgentMgr] Agent 断开",
			log.Int32("当前连接数=", this.connNum.Load()),
		)

	}
}

// -----------------------------------------------------------------------------
// public

func (this *AgentMgr) Run() {
	// 启动心跳管理
	if this.heartbeat > 0 {
		go this.checkHeart()
	}
}

// 停止连接管理
func (this *AgentMgr) Stop() {
	this.stoping = true
	close(this.chDie)

	this.agentMap.Range(func(key, v interface{}) bool {
		if a, ok := v.(*Agent); ok {
			a.Stop()
		}

		this.agentMap.Delete(key)
		return true
	})
}

// 设置握手 key
func (this *AgentMgr) SetKey(k string) {
	if "" != k {
		this.key = k
	}
}

// 设置心跳 key
func (this *AgentMgr) SetHeartbeat(h time.Duration) {
	this.heartbeat = h
	this.heartSend = int64(h) / 2
	this.heartRecv = int64(h)
}

// 设置 handler
func (this *AgentMgr) SetHandler(h IHandler) {
	if nil != h {
		this.handler = h
	}
}

// 设置消息通道
func (this *AgentMgr) SetPacketChan(ch chan *Packet) {
	if ch != nil {
		this.packetChan = ch
	}
}

// 获取当前连接数
func (this *AgentMgr) GetConnNum() int32 {
	return this.connNum.Load()
}

// -----------------------------------------------------------------------------
// private

// 创建代理
func (this *AgentMgr) newAgent(conn net.Conn) {
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
	a.SetClientPacketChan(this.clientPkt)
	a.SetServerPacketChan(this.serverPkt)
	id := this.agentId.Add(1)
	a.SetId(id)
	a.SetMgr(this)

	this.agentMap.Store(id, a)
	this.connNum.Add(1)
	a.Run()
}

// 检测心跳
func (this *AgentMgr) checkHeart() {
	// 半程检测
	hb := this.heartbeat / 2
	ticker := time.NewTicker(hb)

	defer func() {
		log.Logger.Debug(
			"[AgentMgr] checkHeart 结束",
		)

		ticker.Stop()
	}()

	for {
		select {
		case <-ticker.C:
			this.check()
		case <-this.chDie:
			return
		}
	}
}

// 检查发送是否超时
func (this *AgentMgr) check() {
	t := time.Now()

	this.agentMap.Range(func(key, v interface{}) bool {
		if a, ok := v.(*Agent); ok {
			this.checkSendTime(a, t)
			this.checkRecvTime(a, t)
		}

		return true
	})
}

// 检查发送是否超时
func (this *AgentMgr) checkSendTime(a *Agent, t time.Time) {
	pass := t.UnixNano() - a.lastSend.Load()
	if pass >= this.heartSend {
		a.sendHeartbeat()
	}
}

// 检查接收是否超时
func (this *AgentMgr) checkRecvTime(a *Agent, t time.Time) {
	pass := t.UnixNano() - a.lastRecv.Load()
	if pass >= this.heartRecv {
		log.Logger.Debug(
			"[AgentMgr] agent 心跳超时，关闭该连接",
			log.String("agent", a.String()),
		)

		a.Stop()
	}
}
