// /////////////////////////////////////////////////////////////////////////////
// 代理对应于用户，用于存储原始连接信息

package network

import (
	"github.com/pkg/errors"         // 异常
	"github.com/zpab123/sco/scoerr" // 异常
	"github.com/zpab123/zaplog"     // 日志
)

// /////////////////////////////////////////////////////////////////////////////
// Agent

// 代理对应于用户，用于存储原始连接信息
type Agent struct {
	options *TAgentOpt // 配置参数
	scoConn *ScoConn   // sco 引擎连接对象
}

// 新建1个 Agent 对象
//
// 创建失败，返回 nil error
func NewAgent(socket ISocket, opt *TAgentOpt) (*Agent, error) {
	var err error
	// 参数效验
	if nil == socket {
		err = errors.New("创建 Agent 失败：参数 socket=nil")
		return nil, err
	}

	// 创建 ScoConn
	if nil == opt {
		opt = NewTAgentOpt()
	}
	sc := NewScoConn(socket, opt.ScoConnOpt)

	// Agent
	a := Agent{
		options: opt,
		scoConn: sc,
	}

	return &a, nil
}

// 启动 Agent
func (this *Agent) Run() {
	go this.sendLoop()
	// 接收循环，这里不能 go this.recvLoop()，会导致 websocket 连接直接断开
	this.recvLoop()
}

// 停止 Agent
func (this *Agent) Stop() {
	this.scoConn.packetSocket.SendPacket(nil)
	this.scoConn.Close()
}

// 接收线程
func (this *Agent) recvLoop() {
	defer func() {
		zaplog.Debugf("recvLoop 结束")
		this.Stop()
	}()

	for {
		pkt, err := this.scoConn.RecvPacket()
		if nil != pkt {
			zaplog.Debugf("收到消息mid=%d", pkt.GetMid())
			zaplog.Debugf("收到消息bodyLen=%d", pkt.GetBodyLen())
			zaplog.Debugf("收到消息body=%s", string(pkt.GetBody()))

			continue
		}

		// 错误处理
		if nil != err && !scoerr.IsTimeoutError(err) {
			if scoerr.IsConnectionError(err) {
				break
			} else {
				panic(err)
			}
		}
	}
}

// 发送线程
func (this *Agent) sendLoop() {
	defer func() {
		zaplog.Debugf("sendLoop 结束")
		this.Stop()
	}()

	var err error

	for {
		err = this.scoConn.Flush() // 刷新缓冲区
		if nil != err {
			break
		}
	}
}
