package network

import "sync"

/*
单例模式
*/

// 全局唯一的网络实例，模拟网络功能
type network struct {
	sockets sync.Map
}

// 懒汉版单例模式
var instance = &network{sockets: sync.Map{}}

func Instance() *network {
	return instance
}

func (n *network) Listen(endpoint Endpoint, socket Socket) error {
	if _, ok := n.sockets.Load(endpoint); ok {
		return ErrEndpointAlreadyListened
	}
	n.sockets.Store(endpoint, socket)
	return nil
}

func (n *network) Disconnect(endpoint Endpoint) {
	n.sockets.Delete(endpoint)
}

func (n *network) DisconnectAll() {
	n.sockets = sync.Map{}
}

func (n *network) Send(packet *Packet) error {
	record, rOk := n.sockets.Load(packet.Dest())
	socket, sOk := record.(Socket)
	if !rOk || !sOk {
		return ErrConnectionRefuse
	}
	go socket.Receive(packet)
	return nil
}
