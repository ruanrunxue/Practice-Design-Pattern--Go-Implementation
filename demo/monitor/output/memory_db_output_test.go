package output

import (
	"demo/monitor/config"
	"demo/monitor/model"
	"demo/monitor/plugin"
	"testing"
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
	mrecord := model.NewMonitoryRecord()
	mrecord.Endpoint = "service1"
	event := plugin.NewEvent(mrecord)
	mo.Output(event)

	result := new(model.MonitorRecord)
	mo.db.Query(mo.tableName, 1, result)
	if result.Endpoint != "service1" {
		t.Errorf("want service1 got %s", result.Endpoint)
	}
	mo.db.DeleteTable(mo.tableName)
	mo.Uninstall()
}
