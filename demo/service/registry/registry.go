package registry

import (
	"demo/db"
	"demo/network"
	"demo/network/http"
	"demo/service/registry/model"
	"demo/sidecar"
	"reflect"
)

const (
	regionTable       = "regions"
	profileTable      = "profiles"
	subscriptionTable = "subscriptions"
)

// Registry 服务注册中心
type Registry struct {
	db            db.Db
	server        *http.Server
	localIp       string
	svcManagement *svcManagement
	svcDiscovery  *svcDiscovery
}

func NewRegistry(localIp string, db db.Db, factory sidecar.Factory) *Registry {
	return &Registry{
		db:            db,
		server:        http.NewServer(factory.Create()).Listen(localIp, 80),
		localIp:       localIp,
		svcManagement: newSvcManagement(localIp, db, factory),
		svcDiscovery:  newSvcDiscovery(db),
	}
}

func (r *Registry) Run() error {
	if err := r.db.CreateTableIfNotExist(db.NewTable(regionTable).WithType(reflect.TypeOf(new(model.Region)))); err != nil {
		return err
	}
	if err := r.db.CreateTableIfNotExist(db.NewTable(profileTable).WithType(reflect.TypeOf(new(model.ServiceProfileRecord)))); err != nil {
		return err
	}
	if err := r.db.CreateTableIfNotExist(db.NewTable(subscriptionTable).WithType(reflect.TypeOf(new(model.Subscription)))); err != nil {
		return err
	}

	return r.server.Put("/api/v1/service-profile", r.svcManagement.register).
		Post("/api/v1/service-profile", r.svcManagement.update).
		Delete("/api/v1/service-profile", r.svcManagement.deregister).
		Get("/api/v1/service-profile", r.svcDiscovery.discovery).
		Put("/api/v1/subscription", r.svcManagement.subscribe).
		Delete("/api/v1/subscription", r.svcManagement.unsubscribe).
		Start()
}

func (r *Registry) Endpoint() network.Endpoint {
	return network.EndpointOf(r.localIp, 80)
}

func (r *Registry) Shutdown() error {
	r.server.Shutdown()
	return nil
}
