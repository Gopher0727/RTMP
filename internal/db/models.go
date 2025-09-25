package db

import (
	"fmt"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	"github.com/Gopher0727/RTMP/config"
)

// DB 为包级全局数据库连接实例，供其他包使用
// （注意：在 init 之前为 nil）
var DB *gorm.DB

// User 是数据库中的用户模型示例
type User struct {
	ID       uint   `gorm:"primaryKey" json:"id"`
	Username string `gorm:"uniqueIndex;size:64" json:"username"`
	Nick     string `gorm:"size:128" json:"nick"`
}

// InitMySQL 使用给定配置连接 MySQL，并设置全局变量 DB
func InitMySQL(cfg config.MySQLConfig) error {
	// 构造 MySQL DSN: user:password@tcp(host:port)/dbname?charset=utf8mb4&parseTime=True&loc=Local
	// 由于 config 没有 dbname 字段，使用默认 db 名称 rtmp
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=utf8mb4&parseTime=True&loc=Local",
		cfg.UserName, cfg.Password, cfg.Host, cfg.Port, "rtmp")

	gdb, err := gorm.Open(mysql.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
	})
	if err != nil {
		return err
	}

	DB = gdb
	return nil
}

// CloseDB 可在程序退出时调用以关闭底层数据库连接（若需要）
func CloseDB() error {
	if DB == nil {
		return nil
	}
	sqlDB, err := DB.DB()
	if err != nil {
		return err
	}
	return sqlDB.Close()
}
