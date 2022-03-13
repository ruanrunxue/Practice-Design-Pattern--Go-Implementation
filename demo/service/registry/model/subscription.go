package model

import (
	"demo/db"
	"demo/network"
	"errors"
	"strconv"
	"strings"
)

// Subscription 订阅记录对象，记录服务之间的订阅信息，订阅方式支持以下2种：
// 1、按服务Id订阅，ID当对应的服务状态变更时（只能是目标服务只能有1个），通知订阅服务
// 2、按服务类型订阅，当对应类型当服务状态发生变更时（目标服务可以有多个），通知订阅服务
// 3、如果targetServiceId和targetServiceType同时存在时，按照服务ID订阅
// 4、如果targetServiceId和targetServiceType都不存在时，为无效订阅
type Subscription struct {
	Id            string
	SrcSvcId      string      // 源服务ID，订阅方
	TargetSvcId   string      // 目标服务ID，被订阅方
	TargetSvcType ServiceType // 目标服务类型
	NotifyUrl     string      // 订阅方接收通知请求的url，形式为http://ip:port/xxx/xxx/xxx
}

func NewSubscription(id string) *Subscription {
	return &Subscription{Id: id}
}

// NotifyEndpoint 返回被通知方的endpoint
func (s Subscription) NotifyEndpoint() (network.Endpoint, error) {
	url := strings.ReplaceAll(s.NotifyUrl, "http://", "")
	idx := strings.Index(url, "/")
	if idx == -1 {
		return network.Endpoint{}, errors.New("url invalid")
	}
	ipPort := url[0:idx]
	elems := strings.Split(ipPort, ":")
	port, err := strconv.Atoi(elems[1])
	if err != nil {
		return network.Endpoint{}, err
	}
	return network.EndpointOf(elems[0], port), nil
}

// NotifyUri 返回被通知方的uri，/xxxx/xxxx/xxx
func (s Subscription) NotifyUri() (string, error) {
	url := strings.ReplaceAll(s.NotifyUrl, "http://", "")
	idx := strings.Index(url, "/")
	if idx == -1 {
		return "", errors.New("url invalid")
	}
	return url[idx:], nil
}

// SubscriptionVisitor 订阅表遍历, 筛选符合targetSvcId和targetSvcType的订阅记录
type SubscriptionVisitor struct {
	targetSvcId   string
	targetSvcType ServiceType
}

func NewSubscriptionVisitor(targetSvcId string, targetSvcType ServiceType) *SubscriptionVisitor {
	return &SubscriptionVisitor{
		targetSvcId:   targetSvcId,
		targetSvcType: targetSvcType,
	}
}

func (s SubscriptionVisitor) Visit(table *db.Table) ([]interface{}, error) {
	var result []interface{}
	iter := table.Iterator()
	for iter.HasNext() {
		subscription := new(Subscription)
		if err := iter.Next(subscription); err != nil {
			return nil, err
		}
		// 先匹配ServiceId，如果一致则无须匹配ServiceType
		if subscription.TargetSvcId != "" && subscription.TargetSvcId == s.targetSvcId {
			result = append(result, subscription)
			continue
		}
		// ServiceId匹配不上，再匹配ServiceType
		if subscription.TargetSvcType != "" && subscription.TargetSvcType == s.targetSvcType {
			result = append(result, subscription)
		}
	}
	return result, nil
}
