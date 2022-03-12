package config

/*
抽象工厂模式
*/

// Factory 配置抽象工厂接口
type Factory interface {
	CreateInputConfig() Input
	CreateFilterConfig() Filter
	CreateOutputConfig() Output
	CreatePipelineConfig() Pipeline
}
