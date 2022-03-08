package flowctrl

import "testing"

func TestFcContext(t *testing.T) {
	context := NewContext()
	for i := 0; i < 20; i++ {
		context.TryAccept()
	}
	// 时间快进一秒
	context.lastUpdateTimestamp -= 1
	context.TryAccept()
	_, ok := context.curState.(*minorState)
	if !ok {
		t.Error("switch to minorState failed")
	}
	for i := 0; i < 100; i++ {
		context.TryAccept()
	}
	// 时间快进一秒
	context.lastUpdateTimestamp -= 1
	context.TryAccept()
	_, ok = context.curState.(*majorState)
	if !ok {
		t.Error("switch to majorState failed")
	}

	for i := 0; i < 6; i++ {
		context.TryAccept()
	}
	// 时间快进一秒
	context.lastUpdateTimestamp -= 1
	context.TryAccept()
	_, ok = context.curState.(*normalState)
	if !ok {
		t.Error("switch to normalState failed")
	}
}
