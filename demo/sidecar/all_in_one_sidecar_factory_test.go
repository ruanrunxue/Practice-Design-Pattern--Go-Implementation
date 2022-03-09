package sidecar

import (
	"demo/mq"
	"demo/network"
	"demo/network/http"
	"strings"
	"testing"
)

func TestAllInOneSidecar(t *testing.T) {
	factory := NewAllInOneFactory(mq.MemoryMqInstance())
	server := http.NewServer(factory.Create()).Listen("192.168.0.1", 80).
		Get("/hello", func(req *http.Request) *http.Response {
			return http.ResponseOfId(req.ReqId()).AddStatusCode(http.StatusNoContent)
		})
	if err := server.Start(); err != nil {
		t.Error(err)
	}

	client, err := http.NewClient(network.DefaultSocket(), "192.168.0.2")
	if err != nil {
		t.Error(err)
	}
	req := http.EmptyRequest().AddMethod(http.GET).AddUri("/hello")
	resp, err := client.Send(network.EndpointOf("192.168.0.1", 80), req)
	if err != nil {
		t.Error(err)
	}
	if resp.StatusCode() != http.StatusNoContent {
		t.Error(resp.StatusCode())
	}

	msg, err := mq.MemoryMqInstance().Consume("access_log.topic")
	if !strings.Contains(msg.Payload(), "[192.168.0.1:80][RECV_REQ]receive http request from 192.168.0.2:") {
		t.Error("req access log error: " + msg.Payload())
	}
	msg, err = mq.MemoryMqInstance().Consume("access_log.topic")
	if !strings.Contains(msg.Payload(), "[192.168.0.1:80][SEND_RESP]send http response to 192.168.0.2:") {
		t.Error("resp access log error: " + msg.Payload())
	}
	client.Close()
	server.Shutdown()

}
