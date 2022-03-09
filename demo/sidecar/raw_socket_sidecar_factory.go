package sidecar

import (
	"demo/network"
)

// RawSocketFactory 只具备socket功能的sidecar
type RawSocketFactory struct {
}

func (r RawSocketFactory) Create() network.Socket {
	return network.DefaultSocket()
}
