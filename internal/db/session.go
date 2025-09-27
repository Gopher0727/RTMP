package db

import (
	"context"
	"fmt"

	"github.com/redis/go-redis/v9"

	"github.com/Gopher0727/RTMP/config"
)

var RedisSession *redis.Client

// InitRedisSession 初始化Redis会话连接
func InitRedisSession() error {
	cfg := config.GetConfig()
	sessionConfig := cfg.Redis.Session

	RedisSession = redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:%d", sessionConfig.Host, sessionConfig.Port),
		Password: sessionConfig.Password,
		DB:       sessionConfig.DB,
		PoolSize: sessionConfig.PoolSize,
	})

	// 测试连接
	ctx := context.Background()
	if err := RedisSession.Ping(ctx).Err(); err != nil {
		return fmt.Errorf("failed to connect to Redis session: %w", err)
	}

	return nil
}

// GetRedisSession 获取Redis会话连接
func GetRedisSession() *redis.Client {
	if RedisSession == nil {
		panic("Redis session connection not initialized")
	}
	return RedisSession
}

// CloseRedisSession 关闭Redis会话连接
func CloseRedisSession() error {
	if RedisSession != nil {
		return RedisSession.Close()
	}
	return nil
}
