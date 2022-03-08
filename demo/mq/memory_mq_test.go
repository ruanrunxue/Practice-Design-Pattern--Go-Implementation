package mq

import "testing"

func TestMemoryMq(t *testing.T) {
	msg := NewMessage("test", "hello world")
	err := MemoryMqInstance().Produce(msg)
	if err != nil {
		t.Error(err)
	}
	result, err := MemoryMqInstance().Consume("test")
	if err != nil {
		t.Error(err)
	}
	if result.Payload() != "hello world" {
		t.Error("payload not equals")
	}
	MemoryMqInstance().Clear()
}
