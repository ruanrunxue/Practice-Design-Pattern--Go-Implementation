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

/**
 * 里氏替换原则（LSP）：子类型必须能够替换掉它们的基类型，也即基类中的所有性质，在子类中仍能成立。
 * 设计出符合LSP的软件的要点就是，根据该软件的使用者行为作出的合理假设，以此来审视它是否具备有效性和正确性。
 * 要想设计出符合LSP的模型所需要遵循的一些约束：
 * 1、基类应该设计为一个抽象类（不能直接实例化，只能被继承）。
 * 2、子类应该实现基类的抽象接口，而不是重写基类已经实现的具体方法。
 * 3、子类可以新增功能，但不能改变基类的功能。
 * 4、子类不能新增约束，包括抛出基类没有声明的异常。
 * 例子：
 * pipeline.NewPlugin中的入参没有使用plugin.Config作为入参类型，符合LSP。否则就需要转型，破坏了LSP
 */

/*
桥接模式
*/

// Plugin pipeline由input、filter、output三种插件组成，定义了一个数据处理流程
// 数据流向为 input -> filter -> output
type Plugin interface {
	plugin.Plugin
	SetInput(input input.Plugin)
	SetFilter(filter filter.Plugin)
	SetOutput(output output.Plugin)
}

/*
 * 开闭原则（OCP）：一个软件系统应该具备良好的可扩展性，新增功能应当通过扩展的方式实现，而不是在已有的代码基础上修改
 * 根据具体的业务场景识别出那些最有可能变化的点，然后分离出去，抽象成稳定的接口。
 * 后续新增功能时，通过扩展接口，而不是修改已有代码实现
 * 例子：
 * pipeline.Plugin将输入、过滤、输出三个独立变化点，分离到三个接口input.Plugin、filter.Plugin、output.Plugin上，符合OCP
 */

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
	filterChain := filter.NewChain(filterPlugins)
	pipelinePlugin.MethodByName("SetFilter").Call([]reflect.Value{reflect.ValueOf(filterChain)})
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
	filter  filter.Plugin
	output  output.Plugin
	isClose uint32
	run     func()
}

func (p *pipelineTemplate) Install() {
	p.output.Install()
	p.filter.Install()
	p.input.Install()
	p.run()
}

func (p *pipelineTemplate) Uninstall() {
	p.input.Uninstall()
	p.filter.Uninstall()
	p.output.Uninstall()
	atomic.StoreUint32(&p.isClose, 1)
}

func (p *pipelineTemplate) SetInput(input input.Plugin) {
	p.input = input
}

func (p *pipelineTemplate) SetFilter(filter filter.Plugin) {
	p.filter = filter
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
		event = p.filter.Filter(event)
		if err = p.output.Output(event); err != nil {
			fmt.Printf("pipeline output err %s\n", err.Error())
			atomic.StoreUint32(&p.isClose, 1)
			break
		}
	}
}
