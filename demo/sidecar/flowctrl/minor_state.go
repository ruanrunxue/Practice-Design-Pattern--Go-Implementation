package flowctrl

import "math/rand"

// minorState Minor 流控状态，随机流控20%的消息
type minorState struct {
	stateTemplate
}

func newMinorState() *minorState {
	ms := &minorState{}
	ms.stateTemplate.isSameTo = ms.isSameTo
	return ms
}

func (m *minorState) tryAccept() bool {
	val := rand.Uint32() % 100
	return val >= 20
}

func (m *minorState) isSameTo(other state) bool {
	_, ok := other.(*minorState)
	return ok
}
