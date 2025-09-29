package kafka

import "github.com/Gopher0727/RTMP/internal/model"

// SyncMessage 同步消息结构
type SyncMessage struct {
	Type      string      `json:"type"`
	SourceID  string      `json:"source_id"` // 消息来源实例ID
	Timestamp int64       `json:"timestamp"`
	Content   interface{} `json:"content"`
}

// MessagePayload 消息负载结构
type MessagePayload struct {
	Message *model.Message `json:"message"`
}

// StatusPayload 状态负载结构
type StatusPayload struct {
	UserID     uint   `json:"user_id"`
	Status     string `json:"status"`
	InstanceID string `json:"instance_id"`
}
