// /////////////////////////////////////////////////////////////////////////////
// 网络库

package network

// /////////////////////////////////////////////////////////////////////////////
// public api

// 根据名字创建1个新的 Acceptor
func NewAcceptor(name string, addr *TLaddr, mgr IConnManager) (aptor IAcceptor, err error) {
	switch name {
	case C_ACCEPTOR_NAME_TCP: // tcp
		//aptor, err = NewTcpAcceptor(addr, mgr)
	case C_ACCEPTOR_NAME_WS: // websocket
		aptor, err = NewWsAcceptor(addr.WsAddr, mgr)
	case C_ACCEPTOR_NAME_MUL: // tcp ws 混合模式
		//aptor, err = NewMulAcceptor(addr)
	case C_ACCEPTOR_NAME_COM: // tcp ws 组合模式
		//aptor, err = NewComAcceptor(addr, mgr)
	default:
	}

	return
}
