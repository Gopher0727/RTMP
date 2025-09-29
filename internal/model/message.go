package model

import (
	"time"

	"gorm.io/gorm"
)

// MessageType 消息类型
type MessageType string

const (
	MessageTypeText    MessageType = "text"    // 文本消息
	MessageTypeSystem  MessageType = "system"  // 系统消息
	MessageTypeNotify  MessageType = "notify"  // 通知消息
	MessageTypeWarning MessageType = "warning" // 警告消息
)

// MessageTarget 消息目标类型
type MessageTarget string

const (
	MessageTargetUser MessageTarget = "user" // 发送给用户
	MessageTargetRoom MessageTarget = "room" // 发送给房间
	MessageTargetAll  MessageTarget = "all"  // 发送给所有人
)

// Message 消息模型
type Message struct {
	ID         uint           `gorm:"primarykey" json:"id"`
	Content    string         `gorm:"type:text;not null" json:"content"`
	Type       string         `gorm:"size:20;not null" json:"type"` // 使用string类型以支持"user"和"room"
	TargetType MessageTarget  `gorm:"size:20;not null" json:"target_type"`
	TargetID   uint           `gorm:"not null" json:"target_id"`    // 目标ID（用户ID或房间ID）
	SenderID   uint           `json:"sender_id"`                    // 发送者ID，0表示系统
	SenderName string         `gorm:"size:50" json:"sender_name"`   // 发送者名称
	ReceiverID uint           `json:"receiver_id"`                  // 接收者ID，私聊时使用
	RoomID     uint           `json:"room_id"`                      // 房间ID，房间消息时使用
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
