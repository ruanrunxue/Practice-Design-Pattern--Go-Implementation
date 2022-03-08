package mq

import (
	"errors"
	"sync"
)

/*
懒汉式单例模式，通过sync.Once实现
*/

var once = &sync.Once{}
var memoryMqInstance *memoryMq

// memoryMq 内存消息队列，通过channel模式
type memoryMq struct {
	queues sync.Map // key为Topic，value为chan *Message，每个topic单独一个队列
}

func MemoryMqInstance() *memoryMq {
	once.Do(func() {
		memoryMqInstance = &memoryMq{queues: sync.Map{}}
	})
	return memoryMqInstance
}

func (m *memoryMq) Clear() {
	m.queues = sync.Map{}
}

func (m *memoryMq) Consume(topic Topic) (*Message, error) {
	record, ok := m.queues.Load(topic)
	if !ok {
		q := make(chan *Message, 10000)
		m.queues.Store(topic, q)
		record = q
	}
	queue, ok := record.(chan *Message)
	if !ok {
		return nil, errors.New("record's type is not chan *Message")
	}
	return <-queue, nil
}

func (m *memoryMq) Produce(message *Message) error {
	record, ok := m.queues.Load(message.Topic())
	if !ok {
		q := make(chan *Message, 10000)
		m.queues.Store(message.Topic(), q)
		record = q
	}
	queue, ok := record.(chan *Message)
	if !ok {
		return errors.New("record's type is not chan *Message")
	}
	queue <- message
	return nil
}
