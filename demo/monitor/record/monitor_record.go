package record

import "sync/atomic"

// Type 监控记录类型
type Type string

const (
	RecvReq  Type = "recv_req"  // 接收请求
	RecvResp Type = "recv_resp" // 接收响应
	SendReq  Type = "send_req"  // 发送请求
	SendResp Type = "send_resp" // 发送响应
)

// id生成器
var recordId int32 = 0

// MonitorRecord 监控记录
type MonitorRecord struct {
	Id        int
	Endpoint  string
	Type      Type
	Timestamp int64
}

func NewMonitoryRecord() *MonitorRecord {
	return &MonitorRecord{
		Id: int(atomic.AddInt32(&recordId, 1)),
	}
}
