package shopping

import (
	"demo/network"
	"demo/network/http"
	"demo/service/registry/model"
	"demo/sidecar"
	"fmt"
)

// ServiceTemplate 商城应用服务模板，启动时向注册中心注册
type ServiceTemplate struct {
	server           *http.Server
	localIp          string
	sidecarFactory   sidecar.Factory
	svcId            string
	svcType          model.ServiceType
	region           *model.Region
	priority         int
	load             int
	registryEndpoint network.Endpoint
	startService     func() error // 子类实现
}

func (s *ServiceTemplate) Run() error {
	client, err := http.NewClient(s.sidecarFactory.Create(), s.localIp)
	if err != nil {
		return err
	}
	profile := model.NewServiceProfileBuilder().WithId(s.svcId).WithType(s.svcType).
		WithStatus(model.Normal).WithEndpoint(s.localIp, 80).WithRegion(s.region).
		WithPriority(1).WithLoad(100).Build()
	req := http.EmptyRequest().AddUri("/api/v1/service-profile").AddMethod(http.PUT).
		AddBody(profile)
	resp, err := client.Send(s.registryEndpoint, req)
	if err != nil {
		return err
	}
	if !resp.IsSuccess() {
		return fmt.Errorf("register to registry failed: %+v", resp)
	}
	// 注册成功后对外提供服务
	return s.startService()
}

func (s *ServiceTemplate) Endpoint() network.Endpoint {
	return network.EndpointOf(s.localIp, 80)
}

func (s *ServiceTemplate) Shutdown() error {
	s.server.Shutdown()
	return nil
}

func (s *ServiceTemplate) WithLocalIp(localIp string) *ServiceTemplate {
	s.localIp = localIp
	return s
}

func (s *ServiceTemplate) WithSidecarFactory(sidecarFactory sidecar.Factory) *ServiceTemplate {
	s.sidecarFactory = sidecarFactory
	return s
}

func (s *ServiceTemplate) WithSvcId(svcId string) *ServiceTemplate {
	s.svcId = svcId
	return s
}

func (s *ServiceTemplate) WithRegion(id string, name string, country string) *ServiceTemplate {
	region := model.NewRegion(id)
	region.Name = name
	region.Country = country
	s.region = region
	return s
}

func (s *ServiceTemplate) WithPriority(priority int) *ServiceTemplate {
	s.priority = priority
	return s
}

func (s *ServiceTemplate) WithLoad(load int) *ServiceTemplate {
	s.load = load
	return s
}

func (s *ServiceTemplate) WithRegistryEndpoint(registryEndpoint network.Endpoint) *ServiceTemplate {
	s.registryEndpoint = registryEndpoint
	return s
}
