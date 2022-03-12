package output

import (
	"demo/monitor/plugin"
	"reflect"
)

// Type output插件类型
var Type = make(plugin.Types)

func init() {
	Type["memory_db"] = reflect.TypeOf(MemoryDbOutput{})
}
