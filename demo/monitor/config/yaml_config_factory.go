package config

import "gopkg.in/yaml.v2"

// loadYaml 加载yaml配置
func loadYaml(conf string, item interface{}) error {
	return yaml.Unmarshal([]byte(conf), item)
}

// YamlFactory Yaml配置工厂
type YamlFactory struct {
}

func NewYamlFactory() *YamlFactory {
	return &YamlFactory{}
}

func (y YamlFactory) CreateInputConfig() Input {
	return Input{loadConf: loadYaml}
}

func (y YamlFactory) CreateFilterConfig() Filter {
	return Filter{loadConf: loadYaml}
}

func (y YamlFactory) CreateOutputConfig() Output {
	return Output{loadConf: loadYaml}
}

func (y YamlFactory) CreatePipelineConfig() Pipeline {
	pipeline := Pipeline{}
	pipeline.loadConf = loadYaml
	return pipeline
}
