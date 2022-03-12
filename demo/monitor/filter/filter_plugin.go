package filter

import (
	"demo/monitor/config"
	"demo/monitor/plugin"
	"reflect"
)

// Plugin 过滤插件
type Plugin interface {
	plugin.Plugin
	Filter(event *plugin.Event) *plugin.Event
}

// NewPlugin 过滤插件工厂方法
func NewPlugin(config config.Filter) (Plugin, error) {
	filterType, ok := Type[config.PluginType]
	if !ok {
		return nil, plugin.ErrUnknownPlugin
	}
	filterPlugin := reflect.New(filterType)
	ctx := reflect.ValueOf(config.Ctx)
	filterPlugin.MethodByName("SetContext").Call([]reflect.Value{ctx})
	return filterPlugin.Interface().(Plugin), nil
}
