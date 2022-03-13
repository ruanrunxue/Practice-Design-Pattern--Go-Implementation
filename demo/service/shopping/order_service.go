package shopping

import (
	"demo/network/http"
	"fmt"
)

// OrderService 订单服务
type OrderService struct {
	ServiceTemplate
}

func NewOrderService() *OrderService {
	orderService := &OrderService{}
	orderService.svcType = "order-service"
	orderService.ServiceTemplate.startService = orderService.startService
	return orderService
}

func (o *OrderService) startService() error {
	o.server = http.NewServer(o.sidecarFactory.Create())
	return o.server.Listen(o.localIp, 80).
		Put("/api/v1/order", o.createOrder).
		Start()
}

func (o *OrderService) createOrder(req *http.Request) *http.Response {
	fmt.Printf("order service %s create order success\n", o.svcId)
	return http.ResponseOfId(req.ReqId()).AddStatusCode(http.StatusCreate)
}
