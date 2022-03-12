package input

import (
	"demo/monitor/plugin"
	"demo/mq"
)

type MemoryMqInput struct {
	topic    mq.Topic
	consumer mq.Consumable
}

func (m *MemoryMqInput) Install() {
	m.consumer = mq.MemoryMqInstance()
}

func (m *MemoryMqInput) Uninstall() {
}

func (m *MemoryMqInput) SetContext(ctx plugin.Context) {
	if topic, ok := ctx.GetString("topic"); ok {
		m.topic = mq.Topic(topic)
	}
}

func (m *MemoryMqInput) Input() (*plugin.Event, error) {
	msg, err := m.consumer.Consume(m.topic)
	if err != nil {
		return nil, err
	}
	event := plugin.NewEvent(msg.Payload())
	event.AddHeader("topic", string(m.topic))
	return event, nil
}
