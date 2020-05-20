// /////////////////////////////////////////////////////////////////////////////
// cso 错误工具

package scoerr

import (
	"io"
	"net"

	"github.com/pkg/errors" // 错误检查工具
)

// /////////////////////////////////////////////////////////////////////////////
// 数据和接口

// 超时检查接口
type timeoutError interface {
	Timeout() bool // Is it a timeout error
}

// /////////////////////////////////////////////////////////////////////////////
// net 相关

// err 是否是1个超时错误
func IsTimeoutError(err error) bool {
	if err == nil {
		return false
	}

	err = errors.Cause(err)
	ne, ok := err.(timeoutError)

	return ok && ne.Timeout()
}

// io 错误检查
func IsEOFOrNetReadError(err error) bool {
	if err == io.EOF {
		return true
	}
	ne, ok := err.(*net.OpError)

	return ok && ne.Op == "read"
}

// 是否是网络连接相关错误
func IsConnectionError(_err interface{}) bool {
	// 非错误类
	err, ok := _err.(error)
	if !ok {
		return false
	}

	// EOF 错误（比如网络断开）
	err = errors.Cause(err)
	if err == io.EOF {
		return true
	}

	// 非 net.Error 错误
	neterr, ok := err.(net.Error)
	if !ok {
		return false
	}

	// net.Error 错误 但属于 io 超时
	if neterr.Timeout() {
		return false
	}

	return true
}
