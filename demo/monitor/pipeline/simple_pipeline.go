package pipeline

import "demo/monitor/plugin"

// SimplePipeline 简单Pipeline实现，每次运行时新启一个goroutine
type SimplePipeline struct {
	pipelineTemplate
}

func (s *SimplePipeline) SetContext(ctx plugin.Context) {
	s.run = func() {
		go func() {
			s.doRun()
		}()
	}
}
