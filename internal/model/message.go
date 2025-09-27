package model

import (
	"time"

	"gorm.io/gorm"
)

// MessageType 消息类型
type MessageType int

const (
	MessageTypeText    MessageType = 1 // 文本消息
	MessageTypeSystem  MessageType = 2 // 系统消息
	MessageTypeNotify  MessageType = 3 // 通知消息
	MessageTypeWarning MessageType = 4 // 警告消息
)

// MessageTarget 消息目标类型
type MessageTarget int

const (
	MessageTargetUser MessageTarget = 1 // 发送给用户
	MessageTargetRoom MessageTarget = 2 // 发送给房间
	MessageTargetAll  MessageTarget = 3 // 发送给所有人
)

// Message 消息模型
type Message struct {
	ID         uint           `gorm:"primarykey" json:"id"`
	Content    string         `gorm:"type:text;not null" json:"content"`
	Type       MessageType    `gorm:"not null" json:"type"`
	TargetType MessageTarget  `gorm:"not null" json:"target_type"`
	TargetID   uint           `gorm:"not null" json:"target_id"`    // 目标ID（用户ID或房间ID）
	SenderID   uint           `json:"sender_id"`                    // 发送者ID，0表示系统
	SenderName string         `gorm:"size:50" json:"sender_name"`   // 发送者名称
	InstanceID string         `gorm:"size:50" json:"instance_id"`   // 消息所属实例ID
	IsRead     bool           `gorm:"default:false" json:"is_read"` // 是否已读
	CreatedAt  time.Time      `json:"created_at"`
	UpdatedAt  time.Time      `json:"updated_at"`
	DeletedAt  gorm.DeletedAt `gorm:"index" json:"-"`
}

// TableName 指定表名
func (Message) TableName() string {
	return "messages"
}
