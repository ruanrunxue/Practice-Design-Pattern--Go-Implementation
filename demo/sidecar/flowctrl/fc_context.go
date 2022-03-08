package flowctrl

import (
	"sync/atomic"
	"time"
)

// Context 流控状态上下文，根据每秒处理请求速率进行流控
type Context struct {
	// reqCount 当前处理的请求个数，切换状态后更新
	reqCount uint64
	// lastUpdateTimestamp 上一次更新的时间戳，每秒更新一次
	lastUpdateTimestamp int64
	// 当前所处的状态
	curState state
}

func NewContext() *Context {
	ctx := &Context{
		reqCount:            0,
		lastUpdateTimestamp: time.Now().Unix(),
	}
	ctx.switchTo(newNormalState())
	return ctx
}

// TryAccept 判断是否应该接收请求
func (c *Context) TryAccept() bool {
	atomic.AddUint64(&c.reqCount, 1)
	c.curState.trySwitch()
	return c.curState.tryAccept()
}

func (c *Context) switchTo(nextState state) {
	nextState.setContext(c)
	c.curState = nextState
}
