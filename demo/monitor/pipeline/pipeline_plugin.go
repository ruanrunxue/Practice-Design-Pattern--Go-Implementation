package pipeline

import (
	"demo/monitor/config"
	"demo/monitor/filter"
	"demo/monitor/input"
	"demo/monitor/output"
	"demo/monitor/plugin"
	"fmt"
	"reflect"
	"sync/atomic"
)

/*
桥接模式
*/

// Plugin pipeline由input、filter、output三种插件组成，定义了一个数据处理流程
// 数据流向为 input -> filter -> output
type Plugin interface {
	plugin.Plugin
	SetInput(input input.Plugin)
	SetFilters(filters []filter.Plugin)
	SetOutput(output output.Plugin)
}

// NewPlugin Pipeline工厂方法
func NewPlugin(config config.Pipeline) (Plugin, error) {
	pipelineType, ok := Type[config.PluginType]
	if !ok {
		return nil, plugin.ErrUnknownPlugin
	}
	pipelinePlugin := reflect.New(pipelineType)

	pipelinePlugin.MethodByName("SetContext").Call([]reflect.Value{reflect.ValueOf(config.Ctx)})
	// 设置input插件
	inputPlugin, err := input.NewPlugin(config.Input)
	if err != nil {
		return nil, err
	}
	pipelinePlugin.MethodByName("SetInput").Call([]reflect.Value{reflect.ValueOf(inputPlugin)})
	// 设置filter插件
	var filterPlugins []filter.Plugin
	for _, fc := range config.Filters {
		filterPlugin, err := filter.NewPlugin(fc)
		if err != nil {
			return nil, err
		}
		filterPlugins = append(filterPlugins, filterPlugin)
	}
	pipelinePlugin.MethodByName("SetFilters").Call([]reflect.Value{reflect.ValueOf(filterPlugins)})
	// 设置output插件
	outputPlugin, err := output.NewPlugin(config.Output)
	if err != nil {
		return nil, err
	}
	pipelinePlugin.MethodByName("SetOutput").Call([]reflect.Value{reflect.ValueOf(outputPlugin)})

	return pipelinePlugin.Interface().(Plugin), nil
}

type pipelineTemplate struct {
	input   input.Plugin
	filters []filter.Plugin
	output  output.Plugin
	isClose uint32
	run     func()
}

func (p *pipelineTemplate) Install() {
	p.output.Install()
	for _, f := range p.filters {
		f.Install()
	}
	p.input.Install()
	p.run()
}

func (p *pipelineTemplate) Uninstall() {
	p.input.Uninstall()
	for _, f := range p.filters {
		f.Install()
	}
	p.output.Uninstall()
	atomic.StoreUint32(&p.isClose, 1)
}

func (p *pipelineTemplate) SetInput(input input.Plugin) {
	p.input = input
}

func (p *pipelineTemplate) SetFilters(filters []filter.Plugin) {
	p.filters = filters
}

func (p *pipelineTemplate) SetOutput(output output.Plugin) {
	p.output = output
}

func (p *pipelineTemplate) doRun() {
	for atomic.LoadUint32(&p.isClose) != 1 {
		event, err := p.input.Input()
		if err != nil {
			fmt.Printf("pipeline input err %s\n", err.Error())
			atomic.StoreUint32(&p.isClose, 1)
			break
		}
		for _, f := range p.filters {
			event = f.Filter(event)
		}
		if err = p.output.Output(event); err != nil {
			fmt.Printf("pipeline output err %s\n", err.Error())
			atomic.StoreUint32(&p.isClose, 1)
			break
		}
	}
}
