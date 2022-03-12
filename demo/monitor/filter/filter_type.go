package filter

import (
	"demo/monitor/plugin"
	"reflect"
)

// Type filter插件类型
var Type = make(plugin.Types)

func init() {
	Type["extract_log"] = reflect.TypeOf(ExtractLogFilter{})
	Type["add_timestamp"] = reflect.TypeOf(AddTimestampFilter{})
}
