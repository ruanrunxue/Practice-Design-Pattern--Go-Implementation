package input

import (
	"demo/monitor/config"
	"demo/monitor/plugin"
	"demo/mq"
	"demo/network"
	"testing"
)

func TestMemoryMqInputPlugin_New(t *testing.T) {
	ctx := plugin.EmptyContext()
	ctx.Add("topic", "test")
	conf := config.Input{
		Name:       "input0",
		PluginType: "memory_mq",
		Ctx:        ctx,
	}
	inputPlugin, err := NewPlugin(conf)
	if err != nil {
		t.Error(err)
	}
	mi, ok := inputPlugin.(*MemoryMqInput)
	if !ok {
		t.Errorf("want *MemoryMqInput, got %T", mi)
	}

	mi.Install()
	msg := mq.NewMessage("test", "hello")
	mq.MemoryMqInstance().Produce(msg)
	event, _ := mi.Input()
	if event.Payload().(string) != "hello" {
		t.Errorf("want hello, got %v", event.Payload())
	}
	mi.Uninstall()
	mq.MemoryMqInstance().Clear()
}

func TestSocketInputPlugin(t *testing.T) {
	ctx := plugin.EmptyContext()
	ctx.Add("ip", "192.168.0.1")
	ctx.Add("port", "80")
	conf := config.Input{
		Name:       "input0",
		PluginType: "socket",
		Ctx:        ctx,
	}
	inputPlugin, err := NewPlugin(conf)
	if err != nil {
		t.Error(err)
	}
	si, ok := inputPlugin.(*SocketInput)
	if !ok {
		t.Errorf("want *MemoryMqInput, got %T", si)
	}
	if si.endpoint.String() != "192.168.0.1:80" {
		t.Errorf("want 192.168.0.1:80, got %s", si.endpoint.String())
	}

	si.Install()
	packet := network.NewPacket(network.EndpointOf("192.168.0.2", 8088),
		network.EndpointOf("192.168.0.1", 80), "hello")
	network.Instance().Send(packet)
	event, _ := si.Input()
	if event.Payload().(string) != "hello" {
		t.Errorf("want hello, got %v", event.Payload())
	}
	si.Uninstall()
}
