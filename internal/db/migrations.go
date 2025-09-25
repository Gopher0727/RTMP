package db

// Migrate 执行数据库迁移（自动迁移模型）。
// 该函数依赖于全局 DB 已通过 InitMySQL 初始化。
func Migrate() error {
	if DB == nil {
		return nil
	}
	return DB.AutoMigrate(&User{})
}
