package flowctrl

// normalState 正常状态，无须流控
type normalState struct {
	stateTemplate
}

func newNormalState() *normalState {
	ns := &normalState{}
	ns.stateTemplate.isSameTo = ns.isSameTo
	return ns
}

func (n *normalState) tryAccept() bool {
	return true
}

func (n *normalState) isSameTo(other state) bool {
	_, ok := other.(*normalState)
	return ok
}
