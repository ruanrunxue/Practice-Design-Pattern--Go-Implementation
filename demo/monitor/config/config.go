package config

import (
	"demo/monitor/plugin"
)

type Type uint8

const (
	JsonType Type = iota
	YamlType
)

type item struct {
	Name       string         `json:"name" yaml:"name"`
	PluginType string         `json:"type" yaml:"type"`
	Ctx        plugin.Context `json:"context" yaml:"context"`
	loadConf   func(conf string, item interface{}) error
}

type Input item

func (i *Input) Load(conf string) error {
	return i.loadConf(conf, i)
}

type Filter item

func (f *Filter) Load(conf string) error {
	return f.loadConf(conf, f)
}

type Output item

func (o *Output) Load(conf string) error {
	return o.loadConf(conf, o)
}

type Pipeline struct {
	item    `yaml:",inline"` // yaml嵌套时需要加上,inline
	Input   Input            `json:"input" yaml:"input"`
	Filters []Filter         `json:"filters" yaml:"filters,flow"`
	Output  Output           `json:"output" yaml:"output"`
}

func (p *Pipeline) Load(conf string) error {
	return p.loadConf(conf, p)
}
