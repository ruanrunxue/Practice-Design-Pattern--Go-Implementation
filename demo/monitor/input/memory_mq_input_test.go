package input

import (
	"demo/monitor/config"
	"demo/monitor/plugin"
	"demo/mq"
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
