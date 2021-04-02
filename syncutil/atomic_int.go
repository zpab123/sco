// /////////////////////////////////////////////////////////////////////////////
// int 和 bool 类型的原子变量

package syncutil

import "sync/atomic"

// /////////////////////////////////////////////////////////////////////////////
// Int32 原子变量

// int32 类型的 原子变量
type AtomicInt32 int32

// 原子性的获取 *AtomicInt32 的值。
func (this *AtomicInt32) Load() int32 {
	self := (*int32)(this)

	return atomic.LoadInt32(self)
}

// 原子性的将 v 的值保存到 *AtomicInt32
func (this *AtomicInt32) Store(v int32) {
	self := (*int32)(this)

	atomic.StoreInt32(self, v)
}

// 原子性的将 delta 的值添加到 *AtomicInt32 并返回新值。
func (this *AtomicInt32) Add(delta int32) int32 {
	self := (*int32)(this)

	return atomic.AddInt32(self, delta)
}

// 原子性的将 v 的值 与 原来的值 交换，返回原来的值
func (this *AtomicInt32) Swap(v int32) int32 {
	self := (*int32)(this)

	return atomic.SwapInt32(self, v)
}

// 将 AtomicInt32 的值与 old 进行比较，如果相等。则将 AtomicInt32 的值替换为 newv，并返回 true。否则返回 false
func (this *AtomicInt32) CompareAndSwap(old int32, newv int32) bool {
	self := (*int32)(this)

	return atomic.CompareAndSwapInt32(self, old, newv)
}

// /////////////////////////////////////////////////////////////////////////////
// Uint32 原子变量

// Uint32 原子变量
type AtomicUint32 uint32

// 原子性的获取 *AtomicUint32 的值。
func (this *AtomicUint32) Load() uint32 {
	self := (*uint32)(this)

	return atomic.LoadUint32(self)
}

// 原子性的将 v 的值保存到 *AtomicUint32
func (this *AtomicUint32) Store(v uint32) {
	self := (*uint32)(this)

	atomic.StoreUint32(self, v)
}

// 原子性的将 delta 的值添加到 *AtomicUint32 并返回新值。
func (this *AtomicUint32) Add(delta uint32) uint32 {
	self := (*uint32)(this)

	return atomic.AddUint32(self, delta)
}

// 原子性的将 v 的值 与 原来的值 交换，返回原来的值
func (this *AtomicUint32) Swap(v uint32) uint32 {
	self := (*uint32)(this)

	return atomic.SwapUint32(self, v)
}

// 将 AtomicUint32 的值与 old 进行比较，如果相等。则将 AtomicUint32 的值替换为 newv，并返回 true。否则返回 false
func (this *AtomicUint32) CompareAndSwap(old uint32, newv uint32) bool {
	self := (*uint32)(this)

	return atomic.CompareAndSwapUint32(self, old, newv)
}

// /////////////////////////////////////////////////////////////////////////////
// int64 原子变量

// int64 原子变量
type AtomicInt64 int64

// 原子性的获取 *AtomicInt64 的值。
func (this *AtomicInt64) Load() int64 {
	self := (*int64)(this)

	return atomic.LoadInt64(self)
}

// 原子性的将 v 的值保存到 *AtomicInt64
func (this *AtomicInt64) Store(v int64) {
	self := (*int64)(this)

	atomic.StoreInt64(self, v)
}

// 原子性的将 delta 的值添加到 *AtomicInt64 并返回新值。
func (this *AtomicInt64) Add(delta int64) int64 {
	self := (*int64)(this)

	return atomic.AddInt64(self, delta)
}

// 将 v 与 原来的值交换，并返回原来的值
func (this *AtomicInt64) Swap(v int64) int64 {
	self := (*int64)(this)

	return atomic.SwapInt64(self, v)
}

// 将 AtomicInt64 的值与 old 进行比较，如果相等。则将 AtomicInt64 的值替换为 newv，并返回 true。否则返回 false
func (this *AtomicInt64) CompareAndSwap(old int64, newv int64) bool {
	self := (*int64)(this)

	return atomic.CompareAndSwapInt64(self, old, newv)
}

// /////////////////////////////////////////////////////////////////////////////
// uint64 原子变量

// uint64 原子变量
type AtomicUint64 uint64

// 原子性的获取 *AtomicUint64 的值。
func (this *AtomicUint64) Load() uint64 {
	self := (*uint64)(this)

	return atomic.LoadUint64(self)
}

// 原子性的将 v 的值保存到 *AtomicUint64
func (this *AtomicUint64) Store(v uint64) {
	self := (*uint64)(this)

	atomic.StoreUint64(self, v)
}

// 原子性的将 delta 的值添加到 *AtomicUint64 并返回新值。
func (this *AtomicUint64) Add(delta uint64) uint64 {
	self := (*uint64)(this)

	return atomic.AddUint64(self, delta)
}

// 将 v 与 原来的值交换，并返回原来的值
func (this *AtomicUint64) Swap(v uint64) uint64 {
	self := (*uint64)(this)

	return atomic.SwapUint64(self, v)
}

// 将 AtomicUint64 的值与 old 进行比较，如果相等。则将 AtomicUint64 的值替换为 newv，并返回 true。否则返回 false
func (this *AtomicUint64) CompareAndSwap(old uint64, newv uint64) bool {
	self := (*uint64)(this)

	return atomic.CompareAndSwapUint64(self, old, newv)
}

// /////////////////////////////////////////////////////////////////////////////
// int 原子变量

// int 类型的 原子变量 （用 int64 表示的）
type AtomicInt int64

// 原子性的获取 *AtomicInt 的值。
func (this *AtomicInt) Load() int {
	return int(atomic.LoadInt64((*int64)(this)))
}

// 原子性的将 v 的值保存到 *AtomicInt
func (this *AtomicInt) Store(v int) {
	atomic.StoreInt64((*int64)(this), int64(v))
}

// 原子性的将 delta 的值添加到 *AtomicInt 并返回新值。
func (this *AtomicInt) Add(delta int) {
	atomic.AddInt64((*int64)(this), int64(delta))
}

// /////////////////////////////////////////////////////////////////////////////
// bool 原子变量

// bool 类型的 原子变量
type AtomicBool AtomicInt32

// 原子性的将 v 的值保存到 *AtomicBool
func (ab *AtomicBool) Store(v bool) {
	if v {
		(*AtomicInt32)(ab).Store(1)
	} else {
		(*AtomicInt32)(ab).Store(0)
	}
}

// 原子性的获取 *AtomicBool 的值。
func (ab *AtomicBool) Load() bool {
	return (*AtomicInt32)(ab).Load() != 0
}

// 将 v 与 原来的值交换，并返回原来的值
func (ab *AtomicBool) Swap(v bool) bool {
	// 原来的值
	old := ((*AtomicInt32)(ab).Load() != 0)

	// 交换
	if v {
		(*AtomicInt32)(ab).Store(0)
	} else {
		(*AtomicInt32)(ab).Store(1)
	}

	// 返回原来的值
	return old
}
