package config

import "testing"

func TestJsonFactory(t *testing.T) {
	str := "{\"name\":\"pipeline0\", \"type\":\"single_thread\", " +
		"\"input\":{\"name\":\"memory_mq_0\", \"type\":\"memory_mq\", \"context\":{\"topic\":\"test\"}}," +
		"\"output\":{\"name\":\"memory_db_0\", \"type\":\"memory_db\", \"context\":{\"tableName\":\"test\"}}," +
		"\"filters\":" + "[{\"name\":\"log_to_json_0\", \"type\":\"log_to_json\"}," +
		"{\"name\":\"add_timestamp_0\", \"type\":\"add_timestamp\"}," +
		"{\"name\":\"json_to_monitor_event_0\", \"type\":\"json_to_monitor_event\"}]" + "}"
	conf := NewJsonFactory().CreatePipelineConfig()
	err := conf.Load(str)
	if err != nil {
		t.Error(err)
	}
	if conf.Name != "pipeline0" || conf.Input.PluginType != "memory_mq" {
		t.Errorf("load config failed, want name pipeline0, got %s, "+
			"want input type memory_mq got %s", conf.Name, conf.Input.PluginType)
	}
	if len(conf.Filters) != 3 {
		t.Errorf("want filters len 3, got %d", len(conf.Filters))
	}
	if topic, ok := conf.Input.Ctx.GetString("topic"); ok {
		if topic != "test" {
			t.Errorf("want input topic test, got %s", topic)
		}
	} else {
		t.Errorf("load input context failed")
	}
}
