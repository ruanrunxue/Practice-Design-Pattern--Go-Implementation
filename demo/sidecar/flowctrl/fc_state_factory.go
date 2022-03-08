package flowctrl

type stateFactory struct {
}

var factory = stateFactory{}

// rate <= 10：NormalState
// 10 < rate <= 50：MinorState
// rate > 50: MajorState
func (s stateFactory) create(rate uint64) state {
	if rate <= 10 {
		return newNormalState()
	} else if rate <= 50 {
		return newMinorState()
	} else {
		return newMajorState()
	}
}
