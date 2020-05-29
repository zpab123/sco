// /////////////////////////////////////////////////////////////////////////////
// 常量-接扣-types

package rpc

import (
	"github.com/zpab123/sco/discovery" // 服务发现
)

// rpc client 服务
type IClient interface {
	model.IComponent    // 接口继承：组件
	discovery.IListener // 接口继承：服务发现侦听
}

// /////////////////////////////////////////////////////////////////////////////
// 接口

// rpc server 服务
type IServer interface {
}

// rpc client 服务
type IClient interface {
	discovery.IListener // 接口继承：服务发现侦听
}
