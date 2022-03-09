package sidecar

import (
	"demo/network"
	"demo/network/http"
	"demo/sidecar/flowctrl"
)

/*
装饰者模式
*/

// FlowCtrlSidecar HTTP接收端流控功能装饰器，自动拦截Socket接收报文，实现流控功能
type FlowCtrlSidecar struct {
	socket network.Socket
	ctx    *flowctrl.Context
}

func NewFlowCtrlSidecar(socket network.Socket) *FlowCtrlSidecar {
	return &FlowCtrlSidecar{
		socket: socket,
		ctx:    flowctrl.NewContext(),
	}
}

func (f *FlowCtrlSidecar) Listen(endpoint network.Endpoint) error {
	return network.Instance().Listen(endpoint, f)
}

func (f *FlowCtrlSidecar) Close(endpoint network.Endpoint) {
	f.socket.Close(endpoint)
}

func (f *FlowCtrlSidecar) Send(packet *network.Packet) error {
	return f.socket.Send(packet)
}

func (f *FlowCtrlSidecar) Receive(packet *network.Packet) {
	httpReq, ok := packet.Payload().(*http.Request)
	// 如果不是HTTP请求，则不做流控处理
	if !ok {
		f.socket.Receive(packet)
		return
	}
	// 流控后返回429 Too Many Request响应
	if !f.ctx.TryAccept() {
		httpResp := http.ResponseOfId(httpReq.ReqId()).
			AddStatusCode(http.StatusTooManyRequest).
			AddProblemDetails("enter flow ctrl state")
		f.socket.Send(network.NewPacket(packet.Dest(), packet.Src(), httpResp))
		return
	}
	f.socket.Receive(packet)
}

func (f *FlowCtrlSidecar) AddListener(listener network.SocketListener) {
	f.socket.AddListener(listener)
}
