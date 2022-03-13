package registry

import (
	"demo/db"
	"demo/network"
	"demo/network/http"
	"demo/service/registry/model"
	"demo/sidecar"
	"reflect"
	"testing"
)

func TestRegistry(t *testing.T) {
	mdb := db.MemoryDbInstance()
	defer mdb.Clear()
	factory := sidecar.NewRawSocketFactory()
	registry := NewRegistry("192.168.0.1", mdb, factory)
	err := registry.Run()
	if err != nil {
		t.Fatal(err)
	}

	region := model.NewRegion("1")
	region.Name = "region-1"
	region.Country = "CN"
	profile := model.NewServiceProfileBuilder().WithId("svc1").WithType("svc").
		WithStatus(model.Normal).WithRegion(region).WithPriority(1).WithLoad(100).Build()
	rReq := http.EmptyRequest().AddUri("/api/v1/service-profile").AddMethod(http.PUT).
		AddBody(profile)
	client, _ := http.NewClient(network.DefaultSocket(), "192.168.0.2")
	defer client.Close()
	rResp, err := client.Send(registry.Endpoint(), rReq)
	if err != nil {
		t.Fatal(err)
	}
	if rResp.StatusCode() != http.StatusCreate {
		t.Fatalf("want StatusCreate got %v", rResp.StatusCode())
	}

	dReq := http.EmptyRequest().AddUri("/api/v1/service-profile").AddMethod(http.GET).
		AddQueryParam("service-id", "svc1")
	dResp, err := client.Send(registry.Endpoint(), dReq)
	if err != nil {
		t.Fatal(err)
	}
	if dResp.StatusCode() != http.StatusOk {
		t.Fatalf("want StatusOk got %v", dResp.StatusCode())
	}
	dProfile := dResp.Body().(*model.ServiceProfile)
	if !reflect.DeepEqual(profile, dProfile) {
		t.Fatalf("want %+v got %+v", profile, dProfile)
	}

	drReq := http.EmptyRequest().AddUri("/api/v1/service-profile").AddMethod(http.DELETE).
		AddHeader("service-id", "svc1")
	drResp, err := client.Send(registry.Endpoint(), drReq)
	if err != nil {
		t.Fatal(err)
	}
	if drResp.StatusCode() != http.StatusNoContent {
		t.Fatalf("want StatusNoContent got %v", drResp.StatusCode())
	}

	dReq2 := http.EmptyRequest().AddUri("/api/v1/service-profile").AddMethod(http.GET).
		AddQueryParam("service-id", "svc1")
	dResp2, err := client.Send(registry.Endpoint(), dReq2)
	if err != nil {
		t.Fatal(err)
	}
	if dResp2.StatusCode() != http.StatusNotFound {
		t.Fatalf("want StatusNotFound got %v", dResp2.StatusCode())
	}
}
