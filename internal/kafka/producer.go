package kafka

import (
	"encoding/json"
	"log"
	"sync"

	"github.com/IBM/sarama"

	"github.com/Gopher0727/RTMP/config"
)

// MessageProducer Kafka消息生产者
type MessageProducer struct {
	producer sarama.SyncProducer
	topics   map[string]string
	mu       sync.Mutex
}

// MessagePayload 消息负载
type MessagePayload struct {
	Type    string `json:"type"`
	Content any    `json:"content"`
}

// NewMessageProducer 创建新的消息生产者
func NewMessageProducer(cfg *config.Config) (*MessageProducer, error) {
	config := sarama.NewConfig()
	config.Producer.RequiredAcks = sarama.WaitForAll
	config.Producer.Retry.Max = 5
	config.Producer.Return.Successes = true

	producer, err := sarama.NewSyncProducer(cfg.Kafka.Brokers, config)
	if err != nil {
		return nil, err
	}

	// 初始化主题映射
	topics := make(map[string]string)
	for _, topic := range cfg.Kafka.Topics {
		topics[topic] = topic
	}

	return &MessageProducer{
		producer: producer,
		topics:   topics,
	}, nil
}

// SendMessage 发送消息到指定主题
func (p *MessageProducer) SendMessage(topic string, key string, payload any) error {
	p.mu.Lock()
	defer p.mu.Unlock()

	// 序列化消息
	jsonData, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	// 创建消息
	msg := &sarama.ProducerMessage{
		Topic: topic,
		Key:   sarama.StringEncoder(key),
		Value: sarama.ByteEncoder(jsonData),
	}

	// 发送消息
	partition, offset, err := p.producer.SendMessage(msg)
	if err != nil {
		return err
	}

	log.Printf("Message sent to topic %s, partition %d, offset %d", topic, partition, offset)
	return nil
}

// SendUserMessage 发送用户消息
func (p *MessageProducer) SendUserMessage(userID string, payload any) error {
	return p.SendMessage(p.topics["user_messages"], userID, payload)
}

// SendRoomMessage 发送房间消息
func (p *MessageProducer) SendRoomMessage(roomID string, payload any) error {
	return p.SendMessage(p.topics["room_messages"], roomID, payload)
}

// SendSystemMessage 发送系统消息
func (p *MessageProducer) SendSystemMessage(payload any) error {
	return p.SendMessage(p.topics["system_messages"], "system", payload)
}

// Close 关闭生产者
func (p *MessageProducer) Close() error {
	return p.producer.Close()
}
