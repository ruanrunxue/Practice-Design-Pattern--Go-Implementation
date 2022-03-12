package config

import (
	"testing"
)

func TestYamlFactory(t *testing.T) {
	str := "name: pipeline_0\ntype: single_thread\ninput:\n  name: input_0\n  type: memory_mq\n  context:\n    topic: monitor_0\nfilters:\n  - name: filter_0\n    type: log_to_json\n  - name: filter_1\n    type: add_timestamp\n  - name: filter_2\n    type: json_to_monitor_event\noutput:\n  name: output_0\n  type: memory_db\n  context:\n    tableName: monitor_event_0"
	conf := NewYamlFactory().CreatePipelineConfig()
	err := conf.Load(str)
	if err != nil {
		t.Error(err)
	}
	if conf.Name != "pipeline_0" || conf.Input.PluginType != "memory_mq" {
		t.Errorf("load config failed, want name pipeline0, got %s, "+
			"want input type memory_mq got %s", conf.Name, conf.Input.PluginType)
	}
	if len(conf.Filters) != 3 {
		t.Errorf("want filters len 3, got %d", len(conf.Filters))
	}
	if topic, ok := conf.Input.Ctx.GetString("topic"); ok {
		if topic != "monitor_0" {
			t.Errorf("want input topic monitor_0, got %s", topic)
		}
	} else {
		t.Errorf("load input context failed")
	}
}
