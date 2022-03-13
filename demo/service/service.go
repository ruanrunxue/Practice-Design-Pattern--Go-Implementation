package service

import "demo/network"

type Service interface {
	// Run 运行服务
	Run() error
	// Endpoint 返回服务对外提供服务的endpoint
	Endpoint() network.Endpoint
	// Shutdown 停止服务
	Shutdown() error
}
