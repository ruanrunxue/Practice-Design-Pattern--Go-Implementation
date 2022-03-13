package mediator

import (
	"demo/db"
	"demo/network/http"
	"demo/service/registry"
	"demo/service/registry/model"
	"demo/sidecar"
	"testing"
)

func TestServiceMediator(t *testing.T) {
	mdb := db.MemoryDbInstance()
	factory := sidecar.NewRawSocketFactory()

	registryCenter := registry.NewRegistry("192.168.0.1", mdb, factory)
	registryCenter.Run()
	defer registryCenter.Shutdown()

	mediator := NewServiceMediator(registryCenter.Endpoint(), "192.168.0.2", factory)
	mediator.Run()
	defer mediator.Shutdown()

	server := http.NewServer(factory.Create()).Listen("192.168.0.3", 80)
	server.Get("/hello", func(req *http.Request) *http.Response {
		return http.ResponseOfId(req.ReqId()).AddStatusCode(http.StatusNoContent)
	}).Start()
	defer server.Shutdown()
	// 模拟注册
	region := model.NewRegion("1")
	region.Name = "region-1"
	region.Country = "CN"
	profile := model.NewServiceProfileBuilder().WithId("svc1").WithType("svc").
		WithStatus(model.Normal).WithRegion(region).WithEndpoint("192.168.0.3", 80).
		WithPriority(1).WithLoad(100).Build()
	registerReq := http.EmptyRequest().AddUri("/api/v1/service-profile").AddMethod(http.PUT).
		AddBody(profile)
	client1, _ := http.NewClient(factory.Create(), "192.168.0.3")
	defer client1.Close()
	resp, err := client1.Send(registryCenter.Endpoint(), registerReq)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode() != http.StatusCreate {
		t.Fatalf("want http.StatusCreate got %v", resp.StatusCode())
	}

	client2, _ := http.NewClient(factory.Create(), "192.168.0.4")
	req := http.EmptyRequest().AddUri("/svc/hello").AddMethod(http.GET)
	resp2, err := client2.Send(mediator.Endpoint(), req)
	if err != nil {
		t.Fatal(err)
	}
	if resp2.StatusCode() != http.StatusNoContent {
		t.Fatalf("want http.StatusNoContent got %v", resp.StatusCode())
	}
}
