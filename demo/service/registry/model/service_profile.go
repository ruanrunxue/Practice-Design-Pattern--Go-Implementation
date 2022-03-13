package model

import (
	"demo/db"
	"demo/network"
)

// ServiceStatus 服务状态
type ServiceStatus uint8

const (
	Normal ServiceStatus = iota
	Fault
	Unknown
)

type ServiceType string

/*
建造者模式
*/

// ServiceProfile 服务档案，其中服务ID唯一标识一个服务实例，一种服务类型可以有多个服务实例
type ServiceProfile struct {
	Id       string           // 服务ID
	Type     ServiceType      // 服务类型
	Status   ServiceStatus    // 服务状态
	Endpoint network.Endpoint // 服务Endpoint
	Region   *Region          // 服务所属region
	Priority int              // 服务优先级，范围0～100，值越低，优先级越高
	Load     int              // 服务负载，负载越高表示服务处理的业务压力越大
}

func NewServiceProfileBuilder() *serviceProfileBuild {
	return &serviceProfileBuild{profile: &ServiceProfile{}}
}

func (s *ServiceProfile) ToTableRecord() *ServiceProfileRecord {
	return &ServiceProfileRecord{
		Id:       s.Id,
		Type:     s.Type,
		Status:   s.Status,
		Ip:       s.Endpoint.Ip(),
		Port:     s.Endpoint.Port(),
		RegionId: s.Region.Id,
		Priority: s.Priority,
		Load:     s.Load,
	}
}

func (s *ServiceProfile) Clone() Cloneable {
	sp := *s
	return &sp
}

type serviceProfileBuild struct {
	profile *ServiceProfile
}

func (s *serviceProfileBuild) WithId(id string) *serviceProfileBuild {
	s.profile.Id = id
	return s
}

func (s *serviceProfileBuild) WithType(serviceType ServiceType) *serviceProfileBuild {
	s.profile.Type = serviceType
	return s
}

func (s *serviceProfileBuild) WithStatus(status ServiceStatus) *serviceProfileBuild {
	s.profile.Status = status
	return s
}

func (s *serviceProfileBuild) WithEndpoint(ip string, port int) *serviceProfileBuild {
	s.profile.Endpoint = network.EndpointOf(ip, port)
	return s
}

func (s *serviceProfileBuild) WithRegion(region *Region) *serviceProfileBuild {
	s.profile.Region = region
	return s
}

func (s *serviceProfileBuild) WithPriority(priority int) *serviceProfileBuild {
	s.profile.Priority = priority
	return s
}

func (s *serviceProfileBuild) WithLoad(load int) *serviceProfileBuild {
	s.profile.Load = load
	return s
}

func (s *serviceProfileBuild) Build() *ServiceProfile {
	return s.profile
}

// ServiceProfileRecord 存储在数据库里的类型
type ServiceProfileRecord struct {
	Id       string        // 服务ID
	Type     ServiceType   // 服务类型
	Status   ServiceStatus // 服务状态
	Ip       string        // 服务IP
	Port     int           // 服务端口
	RegionId string        // 服务所属regionId
	Priority int           // 服务优先级，范围0～100，值越低，优先级越高
	Load     int           // 服务负载，负载越高表示服务处理的业务压力越大
}

func (s *ServiceProfileRecord) ToServiceProfile() *ServiceProfile {
	return NewServiceProfileBuilder().WithId(s.Id).WithRegion(NewRegion(s.RegionId)).
		WithEndpoint(s.Ip, s.Port).WithStatus(s.Status).WithType(s.Type).
		WithPriority(s.Priority).WithLoad(s.Load).Build()
}

// ServiceProfileVisitor profile表遍历, 筛选符合ServiceId和ServiceType的记录
type ServiceProfileVisitor struct {
	svcId   string
	svcType ServiceType
}

func NewServiceProfileVisitor(svcId string, svcType ServiceType) *ServiceProfileVisitor {
	return &ServiceProfileVisitor{
		svcId:   svcId,
		svcType: svcType,
	}
}

func (s *ServiceProfileVisitor) Visit(table *db.Table) ([]interface{}, error) {
	var result []interface{}
	iter := table.Iterator()
	for iter.HasNext() {
		profile := new(ServiceProfileRecord)
		if err := iter.Next(profile); err != nil {
			return nil, err
		}
		// 先匹配ServiceId，如果一致则无须匹配ServiceType
		if profile.Id != "" && profile.Id == s.svcId {
			result = append(result, profile)
			continue
		}
		// ServiceId匹配不上，再匹配ServiceType
		if profile.Type != "" && profile.Type == s.svcType {
			result = append(result, profile)
		}
	}
	return result, nil
}
