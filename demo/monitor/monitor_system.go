package monitor

import (
	"demo/monitor/config"
	"demo/monitor/pipeline"
	"demo/monitor/plugin"
	"fmt"
)

type System struct {
	plugins       map[string]plugin.Plugin
	configFactory config.Factory
}

func NewSystem(configFactory config.Factory) *System {
	return &System{
		plugins:       make(map[string]plugin.Plugin),
		configFactory: configFactory,
	}
}

func (s *System) LoadConf(conf string) error {
	pipelineConf := s.configFactory.CreatePipelineConfig()
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
