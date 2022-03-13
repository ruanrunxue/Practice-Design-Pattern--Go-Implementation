package mq

/**
 * 接口隔离原则（ISP）：一个模块不应该强迫客户程序依赖它们不想使用的接口，模块间的关系应该建立在最小的接口集上。
 * 实现ISP的关键是将大接口拆分成小接口，而拆分的关键就是接口粒度的把握。接口隔离可以减少模块间耦合，提升系统稳定性。
 * 但是过度地细化和拆分接口，也会导致系统的接口数量的上涨，从而产生更大的维护成本。接口的粒度需要根据具体的业务场景来定，可以参考单一职责原则，将那些为同一类客户端程序提供服务的接口合并在一起。
 * 例子：
 * 根据消息队列的模型，拆分成Consumable, Producible两个接口，由Mq继承它们
 * 生产者依赖Producible，消费者依赖Consumable，符合ISP
 */

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
