package network

/*
单例模式
*/

// 全局唯一的网络实例，模拟网络功能
type network struct {
	sockets map[Endpoint]Socket
}

// 懒汉版单例模式
var instance = &network{sockets: make(map[Endpoint]Socket)}

func Instance() *network {
	return instance
}

func (n *network) Listen(endpoint Endpoint, socket Socket) error {
	if _, ok := n.sockets[endpoint]; ok {
		return ErrEndpointAlreadyListened
	}
	n.sockets[endpoint] = socket
	return nil
}

func (n *network) Disconnect(endpoint Endpoint) {
	delete(n.sockets, endpoint)
}

func (n *network) DisconnectAll() {
	n.sockets = make(map[Endpoint]Socket)
}

func (n *network) Send(packet *Packet) error {
	socket, ok := n.sockets[packet.Dest()]
	if !ok {
		return ErrConnectionRefuse
	}
	go socket.Receive(packet)
	return nil
}
