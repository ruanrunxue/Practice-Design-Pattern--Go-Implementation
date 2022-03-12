package input

import (
	"demo/monitor/config"
	"demo/monitor/plugin"
	"reflect"
)

/*
策略模式
*/

// Plugin 输入插件
type Plugin interface {
	plugin.Plugin
	Input() (*plugin.Event, error)
}

// NewPlugin 输入插件工厂方法
func NewPlugin(config config.Input) (Plugin, error) {
	inputType, ok := Type[config.PluginType]
	if !ok {
		return nil, plugin.ErrUnknownPlugin
	}
	inputPlugin := reflect.New(inputType)
	ctx := reflect.ValueOf(config.Ctx)
	inputPlugin.MethodByName("SetContext").Call([]reflect.Value{ctx})
	return inputPlugin.Interface().(Plugin), nil
}
