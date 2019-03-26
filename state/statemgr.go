// /////////////////////////////////////////////////////////////////////////////
// 状态组件

package state

// /////////////////////////////////////////////////////////////////////////////
// 包 初始化

import (
	"github.com/zpab123/syncutil" // 原子变量工具
)

// /////////////////////////////////////////////////////////////////////////////
// StateManager 对象

// 状态管理，可以安全地被多个线程访问
type StateManager struct {
	state syncutil.AtomicUint32 // 当前状态
}

// 新建1个 StateManager 对象
func NewStateManager() *StateManager {
	sm := &StateManager{}

	return sm
}

// 设置状态
func (this *StateManager) SetState(v uint32) {
	this.state.Store(v)
}

// 获取状态
func (this *StateManager) GetState() uint32 {
	return this.state.Load()
}

// 交换状态
func (this *StateManager) SwapState(old uint32, newv uint32) bool {
	return this.state.CompareAndSwap(old, newv)
}
