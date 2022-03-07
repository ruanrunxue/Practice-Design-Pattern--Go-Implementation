package http

import (
	"demo/network"
	"testing"
)

func TestHttpServer(t *testing.T) {
	server := NewServer(network.DefaultSocket()).Listen("192.168.0.1", 80).
		Get("/hello", func(req *Request) *Response {
			return ResponseOfId(req.ReqId()).AddStatusCode(StatusNoContent)
		})
	if err := server.Start(); err != nil {
		t.Error(err)
	}

	client, err := NewClient(network.DefaultSocket(), "192.168.0.2")
	if err != nil {
		t.Error(err)
	}
	req := EmptyRequest().AddMethod(GET).AddUri("/hello")
	resp, err := client.Send(network.EndpointOf("192.168.0.1", 80), req)
	if err != nil {
		t.Error(err)
	}
	if resp.StatusCode() != StatusNoContent {
		t.Error(resp.StatusCode())
	}

	client.Close()
	server.Shutdown()
}
