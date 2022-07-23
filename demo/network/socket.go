package network

/*
观察者模式
*/

// SocketListener Socket报文监听者
type SocketListener interface {
	Handle(packet *Packet) error
}

// Socket 网络通信Socket接口
type Socket interface {
	// Listen 在endpoint指向地址上起监听
	Listen(endpoint Endpoint) error
	// Close 关闭监听
	Close(endpoint Endpoint)
	// Send 发送网络报文
	Send(packet *Packet) error
	// Receive 接收网络报文
	Receive(packet *Packet)
	// AddListener 增加网络报文监听者
	AddListener(listener SocketListener)
}

// socketImpl Socket的默认实现
type socketImpl struct {
	listeners []SocketListener
}

func DefaultSocket() *socketImpl {
	return &socketImpl{}
}

func (s *socketImpl) Listen(endpoint Endpoint) error {
	return Instance().Listen(endpoint, s)
}

func (s *socketImpl) Close(endpoint Endpoint) {
	Instance().Disconnect(endpoint)
}

func (s *socketImpl) Send(packet *Packet) error {
	return Instance().Send(packet)
}

func (s *socketImpl) Receive(packet *Packet) {
	for _, listener := range s.listeners {
		listener.Handle(packet)
	}
}

func (s *socketImpl) AddListener(listener SocketListener) {
	s.listeners = append(s.listeners, listener)
}
