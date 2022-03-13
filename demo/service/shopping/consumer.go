package shopping

import (
	"demo/network"
	"demo/network/http"
	"fmt"
)

// Consumer 商城消费者
type Consumer struct {
	name           string
	localIp        string
	centerEndpoint network.Endpoint
}

func NewConsumer(name string) *Consumer {
	return &Consumer{name: name}
}

func (c *Consumer) UsePhone(localIp string) *Consumer {
	c.localIp = localIp
	return c
}

func (c *Consumer) LoginShoppingCenter(endpoint network.Endpoint) *Consumer {
	c.centerEndpoint = endpoint
	return c
}

func (c *Consumer) Buy(goods string) {
	client, err := http.NewClient(network.DefaultSocket(), c.localIp)
	if err != nil {
		fmt.Printf("%s buy %s failed: %s\n", c.name, goods, err.Error())
		return
	}
	defer client.Close()
	req := http.EmptyRequest().AddUri("/shopping-center/api/v1/good").
		AddHeader("user", c.name).AddHeader("goods", goods).AddMethod(http.POST)
	_, err = client.Send(c.centerEndpoint, req)
	if err != nil {
		fmt.Printf("%s buy %s failed: %s\n", c.name, goods, err.Error())
		return
	}
}
