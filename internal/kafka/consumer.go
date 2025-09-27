package kafka

import (
	"context"
	"log"
	"sync"

	"github.com/IBM/sarama"

	"github.com/Gopher0727/RTMP/config"
	"github.com/Gopher0727/RTMP/internal/service"
)

// MessageConsumer Kafka消息消费者
type MessageConsumer struct {
	consumer       sarama.ConsumerGroup
	topics         []string
	messageService service.MessageService
	hubService     service.HubService
	wg             sync.WaitGroup
	ctx            context.Context
	cancel         context.CancelFunc
}

// ConsumerHandler 实现sarama.ConsumerGroupHandler接口
type ConsumerHandler struct {
	messageService service.MessageService
	hubService     service.HubService
}

// Setup 在消费者会话开始时调用
func (h ConsumerHandler) Setup(_ sarama.ConsumerGroupSession) error {
	return nil
}

// Cleanup 在消费者会话结束时调用
func (h ConsumerHandler) Cleanup(_ sarama.ConsumerGroupSession) error {
	return nil
}

// ConsumeClaim 处理消费的消息
func (h ConsumerHandler) ConsumeClaim(session sarama.ConsumerGroupSession, claim sarama.ConsumerGroupClaim) error {
	for message := range claim.Messages() {
		log.Printf("Received message: topic=%s, partition=%d, offset=%d, key=%s, value=%s\n",
			message.Topic, message.Partition, message.Offset, string(message.Key), string(message.Value))

		// 处理消息，根据消息类型分发到不同的服务
		// 这里简化处理，实际应该根据消息格式解析并处理

		// 标记消息已处理
		session.MarkMessage(message, "")
	}
	return nil
}

// NewMessageConsumer 创建新的消息消费者
func NewMessageConsumer(cfg *config.Config, messageService service.MessageService, hubService service.HubService) (*MessageConsumer, error) {
	config := sarama.NewConfig()
	config.Consumer.Return.Errors = true
	config.Consumer.Offsets.Initial = sarama.OffsetNewest

	// 创建消费者组
	consumer, err := sarama.NewConsumerGroup(cfg.Kafka.Brokers, cfg.Kafka.ConsumerGroup, config)
	if err != nil {
		return nil, err
	}

	ctx, cancel := context.WithCancel(context.Background())

	return &MessageConsumer{
		consumer:       consumer,
		topics:         cfg.Kafka.Topics,
		messageService: messageService,
		hubService:     hubService,
		ctx:            ctx,
		cancel:         cancel,
	}, nil
}

// Start 启动消费者
func (c *MessageConsumer) Start() {
	c.wg.Add(1)
	go func() {
		defer c.wg.Done()
		handler := ConsumerHandler{
			messageService: c.messageService,
			hubService:     c.hubService,
		}

		for {
			// 消费消息
			if err := c.consumer.Consume(c.ctx, c.topics, handler); err != nil {
				log.Printf("Error from consumer: %v", err)
			}

			// 检查上下文是否已取消
			if c.ctx.Err() != nil {
				return
			}
		}
	}()

	log.Println("Kafka consumer started")
}

// Stop 停止消费者
func (c *MessageConsumer) Stop() {
	c.cancel()
	c.wg.Wait()
	if err := c.consumer.Close(); err != nil {
		log.Printf("Error closing consumer: %v", err)
	}
	log.Println("Kafka consumer stopped")
}
