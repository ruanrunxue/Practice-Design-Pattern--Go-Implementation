package config

import (
	"encoding/json"
)

// loadJson 加载json配置
func loadJson(conf string, item interface{}) error {
	return json.Unmarshal([]byte(conf), item)
}

// JsonFactory Json配置工厂
type JsonFactory struct {
}

func NewJsonFactory() *JsonFactory {
	return &JsonFactory{}
}

// CreateInputConfig 例子 {"name":"input1", "type":"memory_mq", "context":{"topic":"monitor",...}}
func (j JsonFactory) CreateInputConfig() Input {
	return Input{loadConf: loadJson}
}

// CreateFilterConfig 例子 [{"name":"filter1", "type":"to_json"},{"name":"filter2", "type":"add_timestamp"},...]
func (j JsonFactory) CreateFilterConfig() Filter {
	return Filter{loadConf: loadJson}
}

// CreateOutputConfig 例子 {"name":"output1", "type":"memory_db", "context":{"tableName":"test",...}}
func (j JsonFactory) CreateOutputConfig() Output {
	return Output{loadConf: loadJson}
}

// CreatePipelineConfig 例子 {"name":"pipline1", "type":"simple", "input":{...}, "filter":{...}, "output":{...}}
func (j JsonFactory) CreatePipelineConfig() Pipeline {
	pipeline := Pipeline{}
	pipeline.loadConf = loadJson
	return pipeline
}
