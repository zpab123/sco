package network

// /////////////////////////////////////////////////////////////////////////////
// ErrRecvAgain

// 用于重新接收 packet
type ErrRecvAgain struct{}

func (err ErrRecvAgain) Error() string {
	e := "packet 尚未完整，请继续接收"

	return e
}

func (err ErrRecvAgain) Temporary() bool {
	return true
}

func (err ErrRecvAgain) Timeout() bool {
	return true
}

// /////////////////////////////////////////////////////////////////////////////
// ErrRecvAgain

// body 长度错误
type ErrBodyLen struct{}
