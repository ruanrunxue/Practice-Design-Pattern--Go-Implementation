package network

import "strconv"

// Endpoint 值对象，其中ip和port属性为不可变，如果需要变更，需要整对象替换
type Endpoint struct {
	ip   string
	port int
}

// EndpointOf 静态工厂方法，用于实例化对象
func EndpointOf(ip string, port int) Endpoint {
	return Endpoint{
		ip:   ip,
		port: port,
	}
}

// EndpointOfDefaultPort 默认端口为80的工厂方法
func EndpointOfDefaultPort(ip string) Endpoint {
	return Endpoint{
		ip:   ip,
		port: 80,
	}
}

func (e Endpoint) Ip() string {
	return e.ip
}

func (e Endpoint) Port() int {
	return e.port
}

func (e Endpoint) String() string {
	return e.ip + ":" + strconv.Itoa(e.port)
}
