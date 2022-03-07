package http

import (
	"demo/network"
	"errors"
	"math/rand"
	"time"
)

type Client struct {
	socket        network.Socket
	localEndpoint network.Endpoint
	respChan      chan *Response // 用于同步阻塞等待Http响应
}

func NewClient(socket network.Socket, ip string) (*Client, error) {
	// 随机端口，从10000 ～ 19999
	endpoint := network.EndpointOf(ip, int(rand.Uint32()%10000+10000))
	client := &Client{
		socket:        socket,
		localEndpoint: endpoint,
		respChan:      make(chan *Response),
	}
	client.socket.AddListener(client)
	if err := client.socket.Listen(endpoint); err != nil {
		return nil, err
	}
	return client, nil
}

func (c *Client) Close() {
	c.socket.Close(c.localEndpoint)
	close(c.respChan)
}

func (c *Client) Send(dest network.Endpoint, req *Request) (*Response, error) {
	packet := network.NewPacket(c.localEndpoint, dest, req)
	err := c.socket.Send(packet)
	if err != nil {
		return nil, err
	}
	// 发送请求后同步阻塞等待响应
	select {
	case resp, ok := <-c.respChan:
		if ok {
			return resp, nil
		}
		errResp := ResponseOfId(req.ReqId()).AddStatusCode(StatusInternalServerError).
			AddProblemDetails("connection is break")
		return errResp, nil
	case <-time.After(time.Second * time.Duration(3)):
		// 超时时间为3s
		resp := ResponseOfId(req.ReqId()).AddStatusCode(StatusGatewayTimeout).
			AddProblemDetails("http server response timeout")
		return resp, nil
	}
}

func (c *Client) Handle(packet *network.Packet) error {
	resp, ok := packet.Payload().(*Response)
	if !ok {
		return errors.New("invalid packet, not http response")
	}
	c.respChan <- resp
	return nil
}
