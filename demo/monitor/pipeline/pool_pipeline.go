package pipeline

import (
	"demo/monitor/plugin"
	"fmt"
	"github.com/panjf2000/ants"
)

var pool, _ = ants.NewPool(5)

// PoolPipeline 每次启动使用goroutine时启动
type PoolPipeline struct {
	pipelineTemplate
}

func (p *PoolPipeline) SetContext(ctx plugin.Context) {
	p.run = func() {
		if err := pool.Submit(p.doRun); err != nil {
			fmt.Printf("PoolPipeine run error %s", err.Error())
		}
	}
}
