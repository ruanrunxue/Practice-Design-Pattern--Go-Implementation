package record

import "sync/atomic"

// Type 监控记录类型
type Type uint8

const (
	RecvReq  Type = iota // 接收请求
	RecvResp             // 接收响应
	SendReq              // 发送请求
	SendResp             // 发送响应
)

// id生成器
var recordId int32 = 0

// MonitorRecord 监控记录
type MonitorRecord struct {
	Id        int
	ServiceId string
	Type      Type
	Timestamp int64
}

func NewMonitoryRecord(serviceId string, recordType Type, timestamp int64) *MonitorRecord {
	return &MonitorRecord{
		Id:        int(atomic.AddInt32(&recordId, 1)),
		ServiceId: serviceId,
		Type:      recordType,
		Timestamp: timestamp,
	}
}
