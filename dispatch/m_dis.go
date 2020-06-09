// /////////////////////////////////////////////////////////////////////////////
// 常量-接口-types

package dispatch

import (
	"github.com/zpab123/sco/network"
)

// /////////////////////////////////////////////////////////////////////////////
// 接口

// 调度接口
type IDispatcher interface {
	OnRemotePacket(a *network.Agent, pkt *network.Packet)
}
