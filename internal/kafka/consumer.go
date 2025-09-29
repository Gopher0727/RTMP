package kafka

import (
	"context"
	"encoding/json"
	"log"
	"sync"

	"github.com/IBM/sarama"

	"github.com/Gopher0727/RTMP/config"
)

// MessageConsumer Kafka消息消费者
type MessageConsumer struct {
	consumer   sarama.Consumer
	instanceID string
	topics     map[string]bool
	mu         sync.Mutex
	ctx        context.Context
	cancel     context.CancelFunc
	wg         sync.WaitGroup
	handlers   map[string]func(*SyncMessage)
}

// NewMessageConsumer 创建新的消息消费者
func NewMessageConsumer(cfg *config.Config) (*MessageConsumer, error) {
	config := sarama.NewConfig()
	config.Consumer.Return.Errors = true

	consumer, err := sarama.NewConsumer(cfg.Kafka.Brokers, config)
	if err != nil {
		return nil, err
	}

	// 初始化主题映射
	topics := make(map[string]bool)
	for _, topic := range cfg.Kafka.Topics {
		topics[topic] = true
	}

	// 生成实例ID
	instanceID := generateInstanceID()

	ctx, cancel := context.WithCancel(context.Background())

	return &MessageConsumer{
		consumer:   consumer,
		instanceID: instanceID,
		topics:     topics,
		ctx:        ctx,
		cancel:     cancel,
		handlers:   make(map[string]func(*SyncMessage)),
	}, nil
}

// RegisterHandler 注册消息处理器
func (c *MessageConsumer) RegisterHandler(msgType string, handler func(*SyncMessage)) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.handlers[msgType] = handler
}

// Start 启动消息消费
func (c *MessageConsumer) Start() error {
	// 订阅所有主题
	for topic := range c.topics {
		c.wg.Add(1)
		go c.consumeTopic(topic)
	}
	return nil
}

// consumeTopic 消费指定主题的消息
func (c *MessageConsumer) consumeTopic(topic string) {
	defer c.wg.Done()

	// 获取所有分区
	partitions, err := c.consumer.Partitions(topic)
	if err != nil {
		log.Printf("Failed to get partitions for topic %s: %v", topic, err)
		return
	}

	// 为每个分区创建消费者
	for _, partition := range partitions {
		// 跳过当前实例产生的消息
		go func(p int32) {
			pc, err := c.consumer.ConsumePartition(topic, p, sarama.OffsetNewest)
			if err != nil {
				log.Printf("Failed to start consumer for partition %d: %v", p, err)
				return
			}
			defer pc.Close()

			for {
				select {
				case msg := <-pc.Messages():
					// 解析消息
					var syncMsg SyncMessage
					if err := json.Unmarshal(msg.Value, &syncMsg); err != nil {
						log.Printf("Failed to unmarshal message: %v", err)
						continue
					}

					// 跳过自己产生的消息
					if syncMsg.SourceID == c.instanceID {
						continue
					}

					// 调用相应的处理器
					c.mu.Lock()
					handler, exists := c.handlers[syncMsg.Type]
					c.mu.Unlock()

					if exists {
						go handler(&syncMsg)
					}

				case err := <-pc.Errors():
					log.Printf("Error consuming message: %v", err)

				case <-c.ctx.Done():
					return
				}
			}
		}(partition)
	}
}

// Stop 停止消息消费
func (c *MessageConsumer) Stop() error {
	// 取消上下文
	c.cancel()
	// 等待所有goroutine结束
	c.wg.Wait()
	// 关闭消费者
	return c.consumer.Close()
}

// GetInstanceID 获取实例ID
func (c *MessageConsumer) GetInstanceID() string {
	return c.instanceID
}
