package sidecar

import (
	"demo/mq"
	"demo/network"
)

// FactoryFunc Sidecar工厂方法定义
type FactoryFunc func() network.Socket

func RawSocketFactoryFunc() FactoryFunc {
	return func() network.Socket {
		return network.DefaultSocket()
	}
}

func AllInOneFactoryFunc(producer mq.Producible) FactoryFunc {
	return func() network.Socket {
		return NewAccessLogSidecar(NewFlowCtrlSidecar(network.DefaultSocket()), producer)
	}
}
