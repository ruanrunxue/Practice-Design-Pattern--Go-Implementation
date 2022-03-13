package shopping

import (
	"demo/network"
	"demo/network/http"
	"demo/sidecar"
	"fmt"
)

/*
外观模式
*/

type Center struct {
	server                  *http.Server
	localIp                 string
	sidecarFactory          sidecar.Factory
	serviceMediatorEndpoint network.Endpoint
}

func NewCenter(localIp string, factory sidecar.Factory, mediatorEndpoint network.Endpoint) *Center {
	return &Center{
		server:                  http.NewServer(factory.Create()).Listen(localIp, 80),
		localIp:                 localIp,
		sidecarFactory:          factory,
		serviceMediatorEndpoint: mediatorEndpoint,
	}
}

func (c *Center) Run() error {
	return c.server.Post("/shopping-center/api/v1/good", c.buy).Start()
}

func (c *Center) Endpoint() network.Endpoint {
	return network.EndpointOf(c.localIp, 80)
}

func (c *Center) Shutdown() error {
	c.server.Shutdown()
	return nil
}

func (c *Center) buy(req *http.Request) *http.Response {
	goods, ok := req.Header("goods")
	if !ok {
		return http.ResponseOfId(req.ReqId()).AddStatusCode(http.StatusBadRequest).
			AddProblemDetails("goods is empty")
	}
	user, ok := req.Header("user")
	if !ok {
		return http.ResponseOfId(req.ReqId()).AddStatusCode(http.StatusBadRequest).
			AddProblemDetails("user is empty")
	}
	fmt.Printf("\nuser %s start to buy good %s.\n", user, goods)

	client, err := http.NewClient(c.sidecarFactory.Create(), c.localIp)
	if err != nil {
		return http.ResponseOfId(req.ReqId()).AddStatusCode(http.StatusInternalServerError).
			AddProblemDetails(err.Error())
	}
	defer client.Close()

	fmt.Println("\nshopping center send create order request to order service.")
	orderReq := http.EmptyRequest().AddUri("/order-service/api/v1/order").AddMethod(http.PUT)
	orderResp, err := client.Send(c.serviceMediatorEndpoint, orderReq)
	if err != nil {
		return http.ResponseOfId(req.ReqId()).AddStatusCode(http.StatusInternalServerError).
			AddProblemDetails(err.Error())
	}
	fmt.Printf("shopping center receive response from order service, status code %v.\n", orderResp.StatusCode())

	fmt.Println("\nshopping center send check stock request to stock service.")
	stockReq := http.EmptyRequest().AddUri("/stock-service/api/v1/stock").AddMethod(http.GET)
	stockResp, err := client.Send(c.serviceMediatorEndpoint, stockReq)
	if err != nil {
		return http.ResponseOfId(req.ReqId()).AddStatusCode(http.StatusInternalServerError).
			AddProblemDetails(err.Error())
	}
	fmt.Printf("shopping center receive response from stock service, status code %v.\n", stockResp.StatusCode())

	fmt.Println("\nshopping center send payment request to payment service.")
	paymentReq := http.EmptyRequest().AddUri("/payment-service/api/v1/payment").AddMethod(http.POST)
	paymentResp, err := client.Send(c.serviceMediatorEndpoint, paymentReq)
	if err != nil {
		return http.ResponseOfId(req.ReqId()).AddStatusCode(http.StatusInternalServerError).
			AddProblemDetails(err.Error())
	}
	fmt.Printf("shopping center receive response from payment service, status code %v.\n", paymentResp.StatusCode())

	fmt.Println("\nshopping center send shipment request to shipment service.")
	shipmentReq := http.EmptyRequest().AddUri("/shipment-service/api/v1/shipment").AddMethod(http.PUT)
	shipmentResp, err := client.Send(c.serviceMediatorEndpoint, shipmentReq)
	if err != nil {
		return http.ResponseOfId(req.ReqId()).AddStatusCode(http.StatusInternalServerError).
			AddProblemDetails(err.Error())
	}
	fmt.Printf("shopping center receive response from shipment service, status code %v.\n", shipmentResp.StatusCode())

	fmt.Printf("\nuser %s buy goods %s success.\n", user, goods)
	return http.ResponseOfId(req.ReqId()).AddStatusCode(http.StatusOk)
}
