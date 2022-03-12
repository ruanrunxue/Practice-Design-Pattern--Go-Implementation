package output

import (
	"demo/monitor/config"
	"demo/monitor/plugin"
	"demo/monitor/record"
	"testing"
	"time"
)

func TestMemoryDbOutput(t *testing.T) {
	ctx := plugin.EmptyContext()
	ctx.Add("tableName", "test")
	conf := config.Output{
		Name:       "output0",
		PluginType: "memory_db",
		Ctx:        ctx,
	}
	outputPlugin, err := NewPlugin(conf)
	if err != nil {
		t.Error(err)
	}
	mo, ok := outputPlugin.(*MemoryDbOutput)
	if !ok {
		t.Errorf("want *MemoryDbOutput, got %T", mo)
	}

	mo.Install()
	mrecord := record.NewMonitoryRecord("service1", record.RecvResp, time.Now().Unix())
	event := plugin.NewEvent(mrecord)
	mo.Output(event)

	result := new(record.MonitorRecord)
	mo.db.Query(mo.tableName, 1, result)
	if result.ServiceId != "service1" {
		t.Errorf("want service1 got %s", result.ServiceId)
	}
	mo.db.DeleteTable(mo.tableName)
	mo.Uninstall()
}
