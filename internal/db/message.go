package db

import (
	"context"

	"github.com/redis/go-redis/v9"
)

var RedisMessage *redis.Client

func InitRedisMessage() error {
	// Redis 用于消息队列
	RedisMessage = redis.NewClient(&redis.Options{
		Addr:     "127.0.0.1:6380",
		Password: "",
		DB:       0,   // 可以是不同实例，也可以是同实例不同 DB
		PoolSize: 100, // 消息多，连接池大一些
	})

	// 测试连接
	if err := RedisMessage.Ping(context.Background()).Err(); err != nil {
		return err
	}
	return nil
}
