package main

import (
	"demo/db"
	"demo/monitor"
	"demo/monitor/config"
	"demo/mq"
	"demo/service/mediator"
	"demo/service/registry"
	"demo/service/shopping"
	"demo/sidecar"
	"io/ioutil"
)

func main() {
	mdb := db.MemoryDbInstance()
	mmq := mq.MemoryMqInstance()
	sidecarFactory := sidecar.NewAllInOneFactory(mmq)

	// 启动监控系统
	monitorSys := monitor.NewSystem(config.NewYamlFactory())
	conf, _ := ioutil.ReadFile("monitor_pipeline.yaml")
	monitorSys.LoadConf(string(conf))
	monitorSys.Start()
	defer monitorSys.Shutdown()

	// 启动注册中心
	registryCenter := registry.NewRegistry("192.168.0.1", mdb, sidecarFactory)
	registryCenter.Run()
	defer registryCenter.Shutdown()

	// 启动服务中介
	mediatorSvc := mediator.NewServiceMediator(registryCenter.Endpoint(), "192.168.0.2", sidecarFactory)
	mediatorSvc.Run()

	// 启动订单服务
	orderSvc := shopping.NewOrderService().WithRegion("1", "region-1", "CN").
		WithLocalIp("192.168.0.3").WithRegistryEndpoint(registryCenter.Endpoint()).
		WithSvcId("order-0").WithPriority(1).WithLoad(100).WithSidecarFactory(sidecarFactory)
	orderSvc.Run()

	// 启动库存服务
	stockSvc := shopping.NewStockService().WithRegion("1", "region-1", "CN").
		WithLocalIp("192.168.0.4").WithRegistryEndpoint(registryCenter.Endpoint()).
		WithSvcId("stock-0").WithPriority(1).WithLoad(100).WithSidecarFactory(sidecarFactory)
	stockSvc.Run()

	// 启动支付服务
	paymentSvc := shopping.NewPaymentService().WithRegion("1", "region-1", "CN").
		WithLocalIp("192.168.0.5").WithRegistryEndpoint(registryCenter.Endpoint()).
		WithSvcId("payment-0").WithPriority(1).WithLoad(100).WithSidecarFactory(sidecarFactory)
	paymentSvc.Run()

	// 启动发货服务
	shipmentSvc := shopping.NewShipmentService().WithRegion("1", "region-1", "CN").
		WithLocalIp("192.168.0.6").WithRegistryEndpoint(registryCenter.Endpoint()).
		WithSvcId("shipment-0").WithPriority(1).WithLoad(100).WithSidecarFactory(sidecarFactory)
	shipmentSvc.Run()

	// 启动在线商城
	shoppingCenter := shopping.NewCenter("192.168.0.7", sidecarFactory, mediatorSvc.Endpoint())
	shoppingCenter.Run()

	// 消费者从在线商城上购买商品
	shopping.NewConsumer("paul").UsePhone("192.168.0.8").
		LoginShoppingCenter(shoppingCenter.Endpoint()).
		Buy("iphone13")

	console := db.NewConsole(mdb)
	console.Start()
}
