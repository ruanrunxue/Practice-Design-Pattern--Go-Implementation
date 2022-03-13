package model

type NotifyType uint8

const (
	Register NotifyType = iota
	Update
	Deregister
)

// Notification 服务状态变更后通知给订阅者的消息
type Notification struct {
	SubscriptionId string
	Type           NotifyType
	// 注册通知时，为新注册的profile；变更通知时，为变更后的profile；去注册通知时，为之前的profile
	Profile *ServiceProfile
}

func NewNotification(id string) *Notification {
	return &Notification{SubscriptionId: id}
}
