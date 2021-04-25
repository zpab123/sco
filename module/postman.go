// /////////////////////////////////////////////////////////////////////////////
// 用于 msg 分发

package module

// /////////////////////////////////////////////////////////////////////////////
// Observer

// 消息投递员
type Postman struct {
	msgId  uint32        // 消息 id
	recver []*Subscriber // 订阅者集合
}

// 创建一个投递员 id=消息id
func NewPostman(id uint32) *Postman {
	pm := Postman{
		msgId:  id,
		recver: make([]*Subscriber, 0),
	}

	return &pm
}

// -----------------------------------------------------------------------------
// public

// 增加一个订阅者
func (this *Postman) AddSuber(suber *Subscriber) {
	if suber == nil || suber.MsgChan == nil {
		return
	}

	// 已经订阅
	for i, _ := range this.recver {
		if this.recver[i].SuberId == suber.SuberId {
			return
		}
	}

	// 未订阅
	this.recver = append(this.recver, suber)
}

// 删除一个订阅者
// id=订阅者id
func (this *Postman) DelSuber(suber *Subscriber) {
	for i, _ := range this.recver {
		if this.recver[i].SuberId == suber.SuberId {
			this.recver = append(this.recver[:i], this.recver[i+1:]...)
			return
		}
	}
}

// 分发消息
func (this *Postman) Dispath(msg Messge) {
	for i, _ := range this.recver {
		this.recver[i].MsgChan <- msg
	}
}

// -----------------------------------------------------------------------------
// private
