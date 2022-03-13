package shopping

import (
	"demo/network/http"
	"fmt"
)

// ShipmentService 发货服务
type ShipmentService struct {
	ServiceTemplate
}

func NewShipmentService() *ShipmentService {
	shipmentService := &ShipmentService{}
	shipmentService.svcType = "shipment-service"
	shipmentService.ServiceTemplate.startService = shipmentService.startService
	return shipmentService
}

func (s *ShipmentService) startService() error {
	s.server = http.NewServer(s.sidecarFactory.Create())
	return s.server.Listen(s.localIp, 80).
		Put("/api/v1/shipment", s.ship).
		Start()
}

// 发货成功
func (s *ShipmentService) ship(req *http.Request) *http.Response {
	fmt.Printf("shipment service %s ship good success\n", s.svcId)
	return http.ResponseOfId(req.ReqId()).AddStatusCode(http.StatusCreate)
}
