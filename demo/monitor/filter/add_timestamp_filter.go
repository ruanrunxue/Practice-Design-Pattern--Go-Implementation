package filter

import (
	"demo/monitor/model"
	"demo/monitor/plugin"
	"time"
)

// AddTimestampFilter 为MonitorRecord增加时间戳
type AddTimestampFilter struct {
}

func (a *AddTimestampFilter) Install() {
}

func (a *AddTimestampFilter) Uninstall() {
}

func (a *AddTimestampFilter) SetContext(ctx plugin.Context) {
}

func (a *AddTimestampFilter) Filter(event *plugin.Event) *plugin.Event {
	re, ok := event.Payload().(*model.MonitorRecord)
	if !ok {
		return event
	}
	re.Timestamp = time.Now().Unix()
	return plugin.NewEvent(re)
}
