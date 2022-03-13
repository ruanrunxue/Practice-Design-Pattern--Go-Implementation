package mediator

import (
	"demo/network/http"
)

/*
中介者模式
*/

// Mediator HTTP请求/响应转发中介
type Mediator interface {
	// Forward 转发请求，返回对端响应
	Forward(req *http.Request) *http.Response
}
