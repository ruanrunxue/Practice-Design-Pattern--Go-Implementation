package flowctrl

import (
	"sync/atomic"
	"time"
)

/*
状态模式
*/

// state 流控状态接口
type state interface {
	// tryAccept 判断当前是否处理请求
	tryAccept() bool
	// trySwitch 尝试切换到下一个状态，可能未满足条件而保持不变
	trySwitch() bool
	// setContext 设置流控上下文
	setContext(context *Context)
}

/*
模板方法模式
*/

// stateTemplate 流控状态模板
type stateTemplate struct {
	ctx *Context
	// 判断当前状态是否与other一致，由实际状态定义
	isSameTo func(other state) bool
}

// 模板方法，尝试切换到下个状态，其中isSameTo方法由子类实现
func (s *stateTemplate) trySwitch() bool {
	now := time.Now().Unix()
	interval := now - atomic.LoadInt64(&s.ctx.lastUpdateTimestamp)
	// 未到1s则不需要切换状态
	if interval < 1 {
		return false
	}
	atomic.StoreInt64(&s.ctx.lastUpdateTimestamp, now)
	reqCount := atomic.LoadUint64(&s.ctx.reqCount)
	atomic.CompareAndSwapUint64(&s.ctx.reqCount, reqCount, 0)
	rate := reqCount / uint64(interval)
	nextState := factory.create(rate)
	if s.isSameTo(nextState) {
		return false
	}
	s.ctx.switchTo(nextState)
	return true
}

func (s *stateTemplate) setContext(context *Context) {
	s.ctx = context
}
