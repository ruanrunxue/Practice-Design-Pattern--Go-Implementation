package flowctrl

import "math/rand"

// majorState Major流控状态，随机流控50%的消息
type majorState struct {
	stateTemplate
}

func newMajorState() *majorState {
	ms := &majorState{}
	ms.stateTemplate.isSameTo = ms.isSameTo
	return ms
}

func (m *majorState) tryAccept() bool {
	val := rand.Uint32() % 100
	return val >= 50
}

func (m *majorState) isSameTo(other state) bool {
	_, ok := other.(*majorState)
	return ok
}
