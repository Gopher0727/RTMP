package kafka

import (
	"encoding/hex"
	"log"
	"math/rand"
	"time"

	"github.com/Gopher0727/RTMP/config"
	"github.com/Gopher0727/RTMP/internal/service"
)

var (
	producer *MessageProducer
	consumer *MessageConsumer
)

// InitKafka 初始化Kafka生产者和消费者
func InitKafka(cfg *config.Config) error {
	var err error

	// 初始化生产者
	producer, err = NewMessageProducer(cfg)
	if err != nil {
		return err
	}

	log.Println("Kafka producer initialized")
	return nil
}

// InitConsumer 初始化消费者（需要在服务初始化后调用）
func InitConsumer(cfg *config.Config, messageService service.IMessageService, hubService service.IHubService) error {
	var err error

	// 初始化消费者
	consumer, err = NewMessageConsumer(cfg)
	if err != nil {
		return err
	}

	// 启动消费者
	consumer.Start()

	log.Println("Kafka consumer initialized and started")
	return nil
}

// GetProducer 获取Kafka生产者实例
func GetProducer() *MessageProducer {
	return producer
}

// GetConsumer 获取Kafka消费者实例
func GetConsumer() *MessageConsumer {
	return consumer
}

// CloseKafka 关闭Kafka连接
func CloseKafka() {
	if producer != nil {
		if err := producer.Close(); err != nil {
			log.Printf("Error closing Kafka producer: %v", err)
		}
	}

	if consumer != nil {
		consumer.Stop()
	}

	log.Println("Kafka connections closed")
}

// generateInstanceID 生成唯一的实例ID
func generateInstanceID() string {
	b := make([]byte, 8)
	_, err := rand.Read(b)
	if err != nil {
		// 如果随机数生成失败，使用时间戳和进程ID作为备用
		return time.Now().Format("20060102150405")
	}
	return "instance-" + hex.EncodeToString(b)
}
