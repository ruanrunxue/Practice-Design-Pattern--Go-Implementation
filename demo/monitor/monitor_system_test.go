package monitor

import (
	"demo/db"
	"demo/monitor/config"
	"demo/monitor/model"
	"demo/mq"
	"testing"
	"time"
)

func TestMonitorSystem(t *testing.T) {
	system := NewSystem()
	conf := "name: pipeline_0\ntype: simple\ninput:\n  name: input_0\n  type: memory_mq\n  context:\n    topic: access_log.topic\nfilters:\n  - name: filter_0\n    type: extract_log\n  - name: filter_1\n    type: add_timestamp\noutput:\n  name: output_0\n  type: memory_db\n  context:\n    tableName: monitor_record_0"
	err := system.LoadConf(conf, config.YamlType)
	if err != nil {
		t.Error(err)
	}
	system.Start()
	time.Sleep(100 * time.Millisecond)
	log := "[192.168.1.1:8088][recv_req]receive request from address 192.168.1.91:80 success"
	msg := mq.NewMessage("access_log.topic", log)
	mq.MemoryMqInstance().Produce(msg)
	time.Sleep(100 * time.Millisecond)
	result := new(model.MonitorRecord)
	db.MemoryDbInstance().Query("monitor_record_0", 1, result)
	if result.Endpoint != "192.168.1.1:8088" {
		t.Errorf("want 192.168.1.1:8088 got %s", result.Endpoint)
	}
	system.Shutdown()
	db.MemoryDbInstance().Clear()
	mq.MemoryMqInstance().Clear()
}
