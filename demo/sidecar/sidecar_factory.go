package sidecar

import "demo/network"

/*
工厂模式
*/

// Factory Sidecar工厂接口
type Factory interface {
	Create() network.Socket
}
