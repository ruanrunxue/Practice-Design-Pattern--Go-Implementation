package mediator

import (
	"demo/network"
	"demo/network/http"
	"demo/service/registry/model"
	"demo/sidecar"
	"errors"
	"strings"
)

// ServiceMediator 服务中介，根据mediator-uri，向Registry发现对端地址，转发请求
// 其中，mediator-uri的形式为/{serviceType}+ServiceUri
type ServiceMediator struct {
	registryEndpoint network.Endpoint
	localIp          string
	server           *http.Server
	sidecarFactory   sidecar.Factory
}

func NewServiceMediator(registryEndpoint network.Endpoint, localIp string, sidecarFactory sidecar.Factory) *ServiceMediator {
	return &ServiceMediator{
		registryEndpoint: registryEndpoint,
		localIp:          localIp,
		server:           http.NewServer(sidecarFactory.Create()).Listen(localIp, 80),
		sidecarFactory:   sidecarFactory,
	}
}

// Forward 转发请求，请求URL为 /{serviceType}+ServiceUri 的形式，如/serviceA/api/v1/task
func (s *ServiceMediator) Forward(req *http.Request) *http.Response {
	svcType := s.svcTypeOf(req.Uri())
	svcUri := s.svcUriOf(req.Uri())

	dest, err := s.discovery(svcType)
	if err != nil {
		return http.ResponseOfId(req.ReqId()).
			AddStatusCode(http.StatusInternalServerError).
			AddProblemDetails("discovery " + string(svcType) + " failed: " + err.Error())
	}
	forwardReq := req.Clone().AddUri(svcUri)
	client, err := http.NewClient(s.sidecarFactory.Create(), s.localIp)
	if err != nil {
		return http.ResponseOfId(req.ReqId()).
			AddStatusCode(http.StatusInternalServerError).
			AddProblemDetails("create http client failed: " + err.Error())
	}
	defer client.Close()
	resp, err := client.Send(dest, forwardReq)
	if err != nil {
		return http.ResponseOfId(req.ReqId()).
			AddStatusCode(http.StatusInternalServerError).
			AddProblemDetails("forward http req failed: " + err.Error())
	}
	return http.ResponseOfId(req.ReqId()).AddHeaders(resp.Headers()).AddStatusCode(resp.StatusCode()).
		AddProblemDetails(resp.ProblemDetails()).AddBody(resp.Body())
}

func (s *ServiceMediator) Run() error {
	return s.server.Put("/", s.Forward).
		Post("/", s.Forward).
		Get("/", s.Forward).
		Delete("/", s.Forward).
		Start()
}

func (s *ServiceMediator) Endpoint() network.Endpoint {
	return network.EndpointOf(s.localIp, 80)
}

func (s *ServiceMediator) Shutdown() error {
	s.server.Shutdown()
	return nil
}

func (s *ServiceMediator) svcTypeOf(mediatorUri http.Uri) model.ServiceType {
	elems := strings.Split(string(mediatorUri), "/")
	return model.ServiceType(elems[1])
}

func (s *ServiceMediator) svcUriOf(mediatorUri http.Uri) http.Uri {
	tmp := string(mediatorUri)[1:]
	idx := strings.Index(tmp, "/")
	return http.Uri(tmp[idx:])
}

// 根据serviceType进行服务发现
func (s *ServiceMediator) discovery(svcType model.ServiceType) (network.Endpoint, error) {
	client, err := http.NewClient(s.sidecarFactory.Create(), s.localIp)
	if err != nil {
		return network.Endpoint{}, err
	}
	defer client.Close()
	req := http.EmptyRequest().AddUri("/api/v1/service-profile").
		AddMethod(http.GET).AddQueryParam("service-type", string(svcType))
	resp, err := client.Send(s.registryEndpoint, req)
	if err != nil {
		return network.Endpoint{}, err
	}
	if !resp.IsSuccess() {
		return network.Endpoint{}, errors.New(resp.ProblemDetails())
	}
	return resp.Body().(*model.ServiceProfile).Endpoint, nil
}
