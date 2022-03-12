package monitor

import (
	"demo/monitor/config"
	"demo/monitor/pipeline"
	"demo/monitor/plugin"
	"errors"
	"fmt"
)

type System struct {
	plugins map[string]plugin.Plugin
}

func NewSystem() *System {
	return &System{
		plugins: make(map[string]plugin.Plugin),
	}
}

func (s *System) LoadConf(conf string, confType config.Type) error {
	var configFactory config.Factory
	switch confType {
	case config.JsonType:
		configFactory = config.NewJsonFactory()
	case config.YamlType:
		configFactory = config.NewYamlFactory()
	default:
		return errors.New("unknown config type")
	}
	pipelineConf := configFactory.CreatePipelineConfig()
	if err := pipelineConf.Load(conf); err != nil {
		return err
	}
	pipelinePlugin, err := pipeline.NewPlugin(pipelineConf)
	if err != nil {
		return err
	}
	s.plugins[pipelineConf.Name] = pipelinePlugin
	return nil
}

func (s *System) Start() {
	for name, plugin := range s.plugins {
		plugin.Install()
		fmt.Printf("plugin %s install success\n", name)
	}
}

func (s *System) Shutdown() {
	for name, plugin := range s.plugins {
		plugin.Uninstall()
		fmt.Printf("plugin %s uninstall success\n", name)
	}
}
