// /////////////////////////////////////////////////////////////////////////////
// 状态组件

package state

// /////////////////////////////////////////////////////////////////////////////
// 包 初始化

import (
	"github.com/zpab123/sco/syncutil" // 原子变量工具
)

// /////////////////////////////////////////////////////////////////////////////
// StateManager 对象

// 状态管理，可以安全地被多个线程访问
type State struct {
	state syncutil.AtomicUint32 // 当前状态
}

// 新建1个 State 对象
func NewState() *State {
	s := State{}

	return &s
}

// 设置状态
func (this *State) Set(v uint32) {
	this.state.Store(v)
}

// 获取状态
func (this *State) Get() uint32 {
	return this.state.Load()
}

// 对比并交换状态
func (this *State) CompareAndSwap(oldv uint32, newv uint32) bool {
	return this.state.CompareAndSwap(oldv, newv)
}
