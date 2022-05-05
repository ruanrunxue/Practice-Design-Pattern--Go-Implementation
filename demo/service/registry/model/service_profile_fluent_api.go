package model

import "demo/network"

type (
	// fluentServiceProfileBuilder 实现 Fluent API 风格的建造者模式，限定属性构建的顺序
	fluentServiceProfileBuilder struct {
		profile *ServiceProfile
	}
	// 一系列 fluent interface，通过返回值指定下一个构建的属性
	idBuilder interface {
		WithId(id string) typeBuilder
	}
	typeBuilder interface {
		WithType(svcType ServiceType) statusBuilder
	}
	statusBuilder interface {
		WithStatus(status ServiceStatus) endpointBuilder
	}
	endpointBuilder interface {
		WithEndpoint(ip string, port int) regionBuilder
	}
	regionBuilder interface {
		WithRegion(regionId, regionName, regionCountry string) priorityBuilder
	}
	priorityBuilder interface {
		WithPriority(priority int) loadBuilder
	}
	loadBuilder interface {
		WithLoad(load int) endBuilder
	}
	// 最后一个builder返回完成构建的ServiceProfile
	endBuilder interface {
		Build() *ServiceProfile
	}
)

// NewFluentServiceProfileBuilder builder工厂方法
func NewFluentServiceProfileBuilder() idBuilder {
	return &fluentServiceProfileBuilder{profile: &ServiceProfile{}}
}

func (f *fluentServiceProfileBuilder) WithId(id string) typeBuilder {
	f.profile.Id = id
	return f
}

func (f *fluentServiceProfileBuilder) WithType(svcType ServiceType) statusBuilder {
	f.profile.Type = svcType
	return f
}

func (f *fluentServiceProfileBuilder) WithStatus(status ServiceStatus) endpointBuilder {
	f.profile.Status = status
	return f
}

func (f *fluentServiceProfileBuilder) WithEndpoint(ip string, port int) regionBuilder {
	f.profile.Endpoint = network.EndpointOf(ip, port)
	return f
}

func (f *fluentServiceProfileBuilder) WithRegion(regionId, regionName, regionCountry string) priorityBuilder {
	f.profile.Region = &Region{
		Id:      regionId,
		Name:    regionName,
		Country: regionCountry,
	}
	return f
}

func (f *fluentServiceProfileBuilder) WithPriority(priority int) loadBuilder {
	f.profile.Priority = priority
	return f
}

func (f *fluentServiceProfileBuilder) WithLoad(load int) endBuilder {
	f.profile.Load = load
	return f
}

func (f *fluentServiceProfileBuilder) Build() *ServiceProfile {
	return f.profile
}
