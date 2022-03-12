package pipeline

import (
	"demo/monitor/plugin"
	"reflect"
)

// Type input插件类型
var Type = make(plugin.Types)

func init() {
	Type["simple"] = reflect.TypeOf(SimplePipeline{})
	Type["pool"] = reflect.TypeOf(PoolPipeline{})
}
