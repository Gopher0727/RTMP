package db

import (
	"context"
	"fmt"

	"github.com/redis/go-redis/v9"

	"github.com/Gopher0727/RTMP/config"
)

var RedisMessage *redis.Client

// InitRedisMessage 初始化Redis消息连接
func InitRedisMessage() error {
	cfg := config.GetConfig()
	messageConfig := cfg.Redis.Message

	RedisMessage = redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:%d", messageConfig.Host, messageConfig.Port),
		Password: messageConfig.Password,
		DB:       messageConfig.DB,
		PoolSize: messageConfig.PoolSize,
	})

	// 测试连接
	ctx := context.Background()
	if err := RedisMessage.Ping(ctx).Err(); err != nil {
		return fmt.Errorf("failed to connect to Redis message: %w", err)
	}

	return nil
}

// GetRedisMessage 获取Redis消息连接
func GetRedisMessage() *redis.Client {
	if RedisMessage == nil {
		panic("Redis message connection not initialized")
	}
	return RedisMessage
}

// CloseRedisMessage 关闭Redis消息连接
func CloseRedisMessage() error {
	if RedisMessage != nil {
		return RedisMessage.Close()
	}
	return nil
}
