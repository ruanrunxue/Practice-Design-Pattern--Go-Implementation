package filter

import (
	"demo/monitor/config"
	"demo/monitor/plugin"
	"demo/monitor/record"
	"testing"
)

func TestExtractLogFilter(t *testing.T) {
	conf := config.Filter{
		Name:       "filter0",
		PluginType: "extract_log",
		Ctx:        plugin.EmptyContext(),
	}
	filterPlugin, err := NewPlugin(conf)
	if err != nil {
		t.Error(err)
	}
	filterPlugin.Install()
	log := "[192.168.1.1:8088][recv_req]receive request from address 192.168.1.91:80 success"
	event := plugin.NewEvent(log)
	event = filterPlugin.Filter(event)
	re, ok := event.Payload().(*record.MonitorRecord)
	if !ok {
		t.Errorf("want *record.MonitorRecord got %T", event.Payload())
	}
	if re.Endpoint != "192.168.1.1:8088" || re.Type != "recv_req" {
		t.Errorf("want 192.168.1.1:8088 got %s, want recv_req got %s", re.Endpoint, re.Type)
	}
}
