package input

import (
	"demo/monitor/plugin"
	"reflect"
)

// Type input插件类型
var Type = make(plugin.Types)

func init() {
	Type["memory_mq"] = reflect.TypeOf(MemoryMqInput{})
	Type["socket"] = reflect.TypeOf(SocketInput{})
}
