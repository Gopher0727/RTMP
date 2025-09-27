package model

import (
	"time"

	"gorm.io/gorm"
)

const (
	UserStatusOffline = 0
	UserStatusOnline  = 1
)

// User 用户模型
type User struct {
	ID         uint           `gorm:"primarykey" json:"id"`
	Username   string         `gorm:"size:50;not null;uniqueIndex" json:"username"`
	Password   string         `gorm:"size:100;not null" json:"-"`
	Email      string         `gorm:"size:100;not null;uniqueIndex" json:"email"`
	Nickname   string         `gorm:"size:50" json:"nickname"`
	Avatar     string         `gorm:"size:255" json:"avatar"`
	Status     int            `gorm:"default:0" json:"status"`    // 0:离线 1:在线
	InstanceID string         `gorm:"size:50" json:"instance_id"` // 用户所在实例ID
	CreatedAt  time.Time      `json:"created_at"`
	UpdatedAt  time.Time      `json:"updated_at"`
	DeletedAt  gorm.DeletedAt `gorm:"index" json:"-"`
}

// TableName 指定表名
func (User) TableName() string {
	return "users"
}
