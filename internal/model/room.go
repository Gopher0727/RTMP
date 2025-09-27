package model

import (
	"time"

	"gorm.io/gorm"
)

// Room 房间模型
type Room struct {
	ID          uint           `gorm:"primarykey" json:"id"`
	Name        string         `gorm:"size:50;not null" json:"name"`
	Description string         `gorm:"size:255" json:"description"`
	CreatorID   uint           `gorm:"not null" json:"creator_id"`          // 创建者ID
	InstanceID  string         `gorm:"size:50;not null" json:"instance_id"` // 房间所属实例ID
	IsPrivate   bool           `gorm:"default:false" json:"is_private"`     // 是否为私有房间
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
	DeletedAt   gorm.DeletedAt `gorm:"index" json:"-"`
}

// RoomMember 房间成员关系
type RoomMember struct {
	ID        uint           `gorm:"primarykey" json:"id"`
	RoomID    uint           `gorm:"not null;index:idx_room_user" json:"room_id"`
	UserID    uint           `gorm:"not null;index:idx_room_user" json:"user_id"`
	Role      int            `gorm:"default:0" json:"role"` // 0:普通成员 1:管理员 2:创建者
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}

// TableName 指定表名
func (Room) TableName() string {
	return "rooms"
}

// TableName 指定表名
func (RoomMember) TableName() string {
	return "room_members"
}
