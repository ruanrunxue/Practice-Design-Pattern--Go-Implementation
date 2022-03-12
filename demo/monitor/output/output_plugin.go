package output

import (
	"demo/monitor/config"
	"demo/monitor/plugin"
	"reflect"
)

// Plugin 输出插件
type Plugin interface {
	plugin.Plugin
	Output(event *plugin.Event) error
}

// NewPlugin 输出插件工厂方法
func NewPlugin(config config.Output) (Plugin, error) {
	outputType, ok := Type[config.PluginType]
	if !ok {
		return nil, plugin.ErrUnknownPlugin
	}
	outputPlugin := reflect.New(outputType)
	ctx := reflect.ValueOf(config.Ctx)
	outputPlugin.MethodByName("SetContext").Call([]reflect.Value{ctx})
	return outputPlugin.Interface().(Plugin), nil
}
