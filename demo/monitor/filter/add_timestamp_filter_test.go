package filter

import (
	"demo/monitor/config"
	"demo/monitor/plugin"
	"demo/monitor/record"
	"testing"
)

func TestAddTimestampFilter(t *testing.T) {
	conf := config.Filter{
		Name:       "filter0",
		PluginType: "add_timestamp",
		Ctx:        plugin.EmptyContext(),
	}
	filterPlugin, err := NewPlugin(conf)
	if err != nil {
		t.Error(err)
	}
	filterPlugin.Install()
	re := record.NewMonitoryRecord()
	re.Endpoint = "192.168.0.1:80"
	re.Type = record.RecvResp
	event := plugin.NewEvent(re)
	event = filterPlugin.Filter(event)
	if event.Payload().(*record.MonitorRecord).Timestamp == 0 {
		t.Error("timestamp add failed")
	}
}
