package db

import (
	"fmt"

	"gorm.io/gorm"

	"github.com/Gopher0727/RTMP/config"
	"github.com/Gopher0727/RTMP/internal/model"
)

var MySQL *gorm.DB

// InitMySQL 初始化MySQL数据库连接
func InitMySQL() error {
	var err error
	MySQL, err = config.InitMySQL()
	if err != nil {
		return fmt.Errorf("failed to initialize MySQL: %w", err)
	}

	// 自动迁移数据库表
	if err := AutoMigrate(); err != nil {
		return fmt.Errorf("failed to auto migrate: %w", err)
	}

	return nil
}

// AutoMigrate 自动迁移数据库表
func AutoMigrate() error {
	return MySQL.AutoMigrate(
		&model.User{},
		&model.Message{},
		&model.Room{},
		&model.RoomMember{},
	)
}

// GetDB 获取数据库连接
func GetDB() *gorm.DB {
	if MySQL == nil {
		panic("MySQL connection not initialized")
	}
	return MySQL
}

// Close 关闭数据库连接
func Close() error {
	if MySQL != nil {
		sqlDB, err := MySQL.DB()
		if err != nil {
			return err
		}
		return sqlDB.Close()
	}
	return nil
}
