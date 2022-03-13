package shopping

import (
	"demo/network/http"
	"fmt"
)

// StockService 订单服务
type StockService struct {
	ServiceTemplate
}

func NewStockService() *StockService {
	stockService := &StockService{}
	stockService.svcType = "stock-service"
	stockService.ServiceTemplate.startService = stockService.startService
	return stockService
}

func (s *StockService) startService() error {
	s.server = http.NewServer(s.sidecarFactory.Create())
	return s.server.Listen(s.localIp, 80).
		Get("/api/v1/stock", s.checkStock).
		Start()
}

// 检查库存成功
func (s *StockService) checkStock(req *http.Request) *http.Response {
	fmt.Printf("stock service %s check stock success\n", s.svcId)
	return http.ResponseOfId(req.ReqId()).AddStatusCode(http.StatusOk)
}
