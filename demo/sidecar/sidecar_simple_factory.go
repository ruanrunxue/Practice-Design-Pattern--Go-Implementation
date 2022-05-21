package sidecar

import (
	"demo/mq"
	"demo/network"
)

// Type sidecar 类型
type Type uint8

const (
	Raw Type = iota
	AllInOne
)

type SimpleFactory struct {
	producer mq.Producible
}

func (s SimpleFactory) Create(sidecarType Type) network.Socket {
	switch sidecarType {
	case Raw:
		return network.DefaultSocket()
	case AllInOne:
		return NewAccessLogSidecar(NewFlowCtrlSidecar(network.DefaultSocket()), s.producer)
	default:
		return nil
	}
}
