// /////////////////////////////////////////////////////////////////////////////
// entity

package ecs

import (
	"sync/atomic"
)

var (
	idInc uint64 // 最新的实体id编号
)

// /////////////////////////////////////////////////////////////////////////////
// Entity 对象

// 实体对象
type Entity struct {
	id       uint64    // 实体唯一id
	parent   *Entity   // 父实体
	children []*Entity // 子实体
}

// 新建一个 Entity 对象
func NewEntity() *Entity {
	e := Entity{
		id: atomic.AddUint64(&idInc, 1),
	}

	return &e
}

// 添加一个子实体
func (this *Entity) AddChild(child *Entity) {
	child.parent = this
	this.children = append(this.children, child)
}

// 删除一个子实体
func (this *Entity) DelChild(child *Entity) {
	index := -1
	for i, v := range this.children {
		if v.id == child.id {
			n = index
			break
		}
	}

	if index >= 0 {
		this.children = append(this.children[:index], this.children[index:]...)
	}
}
