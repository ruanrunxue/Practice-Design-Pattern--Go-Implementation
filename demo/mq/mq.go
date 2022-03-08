package mq

// Consumable 消费接口，从消息队列中消费数据
type Consumable interface {
	Consume(topic Topic) (*Message, error)
}

// Producible 生产接口，向消息队列生产消费数据
type Producible interface {
	Produce(message *Message) error
}

// Mq 消息队列接口，继承了Consumable和Producible，同时又consume和produce两种行为
type Mq interface {
	Consumable
	Producible
}

type Topic string

// Message 消息队列中消息定义
type Message struct {
	topic   Topic
	payload string
}

func NewMessage(topic Topic, payload string) *Message {
	return &Message{
		topic:   topic,
		payload: payload,
	}
}

func (m Message) Topic() Topic {
	return m.topic
}

func (m Message) Payload() string {
	return m.payload
}
