package db

import (
	"context"

	"github.com/redis/go-redis/v9"
)

var RedisSession *redis.Client

func InitRedisSession() error {
	// Redis 用于 Session/路由
	RedisSession = redis.NewClient(&redis.Options{
		Addr:     "127.0.0.1:6379",
		Password: "",
		DB:       0,  // 单独的库
		PoolSize: 50, // 连接池大小
	})

	// 测试连接
	if err := RedisSession.Ping(context.Background()).Err(); err != nil {
		return err
	}
	return nil
}
