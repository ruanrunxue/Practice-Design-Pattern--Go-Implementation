package sidecar

import (
	"demo/mq"
	"demo/network"
	"demo/network/http"
	"fmt"
)

// AccessLogSidecar HTTP access log修饰器，拦截socket接收和发送报文，上报access log到Mq上，供监控系统统计分析
type AccessLogSidecar struct {
	socket   network.Socket
	producer mq.Producible
	topic    mq.Topic
}

func NewAccessLogSidecar(socket network.Socket, producer mq.Producible) *AccessLogSidecar {
	return &AccessLogSidecar{
		socket:   socket,
		producer: producer,
		topic:    "access_log.topic",
	}
}

func (a *AccessLogSidecar) Listen(endpoint network.Endpoint) error {
	return network.Instance().Listen(endpoint, a)
}

func (a *AccessLogSidecar) Close(endpoint network.Endpoint) {
	a.socket.Close(endpoint)
}

func (a *AccessLogSidecar) Send(packet *network.Packet) error {
	if _, ok := packet.Payload().(*http.Request); ok {
		accessLog := fmt.Sprintf("[%s][SEND_REQ]send http request to %s", packet.Src(), packet.Dest())
		message := mq.NewMessage(a.topic, accessLog)
		a.producer.Produce(message)
	}
	if _, ok := packet.Payload().(*http.Response); ok {
		accessLog := fmt.Sprintf("[%s][SEND_RESP]send http response to %s", packet.Src(), packet.Dest())
		message := mq.NewMessage(a.topic, accessLog)
		a.producer.Produce(message)
	}
	return a.socket.Send(packet)
}

func (a *AccessLogSidecar) Receive(packet *network.Packet) {
	if _, ok := packet.Payload().(*http.Request); ok {
		accessLog := fmt.Sprintf("[%s][RECV_REQ]receive http request from %s", packet.Dest(), packet.Src())
		message := mq.NewMessage(a.topic, accessLog)
		a.producer.Produce(message)
	}
	if _, ok := packet.Payload().(*http.Response); ok {
		accessLog := fmt.Sprintf("[%s][RECV_RESP]receive http response from %s", packet.Dest(), packet.Src())
		message := mq.NewMessage(a.topic, accessLog)
		a.producer.Produce(message)
	}
	a.socket.Receive(packet)
}

func (a *AccessLogSidecar) AddListener(listener network.SocketListener) {
	a.socket.AddListener(listener)
}
