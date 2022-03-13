package shopping

import (
	"demo/network/http"
	"fmt"
)

// PaymentService 支付服务
type PaymentService struct {
	ServiceTemplate
}

func NewPaymentService() *PaymentService {
	paymentService := &PaymentService{}
	paymentService.svcType = "payment-service"
	paymentService.ServiceTemplate.startService = paymentService.startService
	return paymentService
}

func (p *PaymentService) startService() error {
	p.server = http.NewServer(p.sidecarFactory.Create())
	return p.server.Listen(p.localIp, 80).
		Post("/api/v1/payment", p.deduct).
		Start()
}

// 支付成功
func (p *PaymentService) deduct(req *http.Request) *http.Response {
	fmt.Printf("payment service %s deduct money success\n", p.svcId)
	return http.ResponseOfId(req.ReqId()).AddStatusCode(http.StatusOk)
}
