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

/**
 * 单一职责原则（SRP）： 一个模块应该有且只有一个导致其变化的原因
 * SRP是聚合和拆分的一个平衡，太过聚合会导致牵一发动全身，拆分过细又会提升复杂性。
 * 要从用户的视角来把握拆分的度，把面向不同用户的功能拆分开。如果实在无法判断/预测，那就等变化发生时再拆分，避免过度的设计。
 */

// Registry 服务注册中心，将服务管理和服务发现拆分至SvcManagement和SvcDiscovery，符合单一职责原则
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
