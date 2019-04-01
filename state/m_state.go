// /////////////////////////////////////////////////////////////////////////////
// 常量-接口-types

package state

// /////////////////////////////////////////////////////////////////////////////
// 常量

// 状态通用
const (
	C_INVALID  uint32 = iota // 无效状态
	C_INIT                   // 初始化状态
	C_RUNING                 // 正在启动中
	C_WORKING                // 工作状态
	C_CLOSEING               // 正在关闭中
	C_CLOSED                 // 关闭完成
	C_STOPING                // 正在停止中
	C_STOPED                 // 停止完成
)

// /////////////////////////////////////////////////////////////////////////////
// 接口

// 状态管理接口
type IStateManager interface {
	SetState(v uint32)                            // 设置状态
	GetState() uint32                             // 获取状态
	CompareAndSwap(oldv uint32, newv uint32) bool // 对比并交换状态
}
