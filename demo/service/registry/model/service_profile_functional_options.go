package model

import "demo/network"

// ServiceProfileOption 定义构建ServiceProfile的函数类型，以*ServiceProfile作为入参
type ServiceProfileOption func(profile *ServiceProfile)

func NewServiceProfile(svcId string, svcType ServiceType, options ...ServiceProfileOption) *ServiceProfile {
	// 除了Id和Type之外的字段提供默认值
	profile := &ServiceProfile{
		Id:       svcId,
		Type:     svcType,
		Status:   Normal,
		Endpoint: network.EndpointOf("192.168.0.1", 80),
		Region:   &Region{Id: "region1", Name: "beijing", Country: "China"},
		Priority: 1,
		Load:     100,
	}
	// 通过ServiceProfileOption来修改字段
	for _, option := range options {
		option(profile)
	}
	return profile
}

func Status(status ServiceStatus) ServiceProfileOption {
	return func(profile *ServiceProfile) {
		profile.Status = status
	}
}

func Endpoint(ip string, port int) ServiceProfileOption {
	return func(profile *ServiceProfile) {
		profile.Endpoint = network.EndpointOf(ip, port)
	}
}

func SvcRegion(svcId, svcName, svcCountry string) ServiceProfileOption {
	return func(profile *ServiceProfile) {
		profile.Region = &Region{
			Id:      svcId,
			Name:    svcName,
			Country: svcCountry,
		}
	}
}

func Priority(priority int) ServiceProfileOption {
	return func(profile *ServiceProfile) {
		profile.Priority = priority
	}
}

func Load(load int) ServiceProfileOption {
	return func(profile *ServiceProfile) {
		profile.Load = load
	}
}
