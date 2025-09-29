package kafka

import (
	"encoding/json"
	"log"
	"strconv"
	"sync"
	"time"

	"github.com/IBM/sarama"

	"github.com/Gopher0727/RTMP/config"
	"github.com/Gopher0727/RTMP/internal/model"
	"github.com/Gopher0727/RTMP/internal/service"
)

// MessageProducer Kafka消息生产者
type MessageProducer struct {
	producer   sarama.SyncProducer
	instanceID string
	topics     map[string]string
	mu         sync.Mutex
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

	// 生成实例ID
	instanceID := generateInstanceID()

	return &MessageProducer{
		producer:   producer,
		instanceID: instanceID,
		topics:     topics,
	}, nil
}

// SendMessage 发送消息到指定主题
func (p *MessageProducer) SendMessage(topic string, key string, value []byte) error {
	p.mu.Lock()
	defer p.mu.Unlock()

	// 创建消息
	msg := &sarama.ProducerMessage{
		Topic: topic,
		Key:   sarama.StringEncoder(key),
		Value: sarama.ByteEncoder(value),
	}

	// 发送消息
	partition, offset, err := p.producer.SendMessage(msg)
	if err != nil {
		return err
	}

	log.Printf("Message sent to topic=%s, partition=%d, offset=%d", topic, partition, offset)
	return nil
}

// SendUserMessage 发送用户消息
func (p *MessageProducer) SendUserMessage(userID uint, message *model.Message) error {
	// 创建符合SyncMessage格式的消息
	syncMsg := SyncMessage{
		Type:      "user_message",
		SourceID:  p.instanceID,
		Timestamp: time.Now().Unix(),
		Content:   message,
	}

	// 序列化消息
	jsonPayload, err := json.Marshal(syncMsg)
	if err != nil {
		return err
	}
	return p.SendMessage(p.topics["user_messages"], strconv.FormatUint(uint64(userID), 10), jsonPayload)
}

// SendRoomMessage 发送房间消息
func (p *MessageProducer) SendRoomMessage(roomID uint, message *model.Message) error {
	// 创建符合SyncMessage格式的消息
	syncMsg := SyncMessage{
		Type:      "room_message",
		SourceID:  p.instanceID,
		Timestamp: time.Now().Unix(),
		Content:   message,
	}

	// 序列化消息
	jsonPayload, err := json.Marshal(syncMsg)
	if err != nil {
		return err
	}
	return p.SendMessage(p.topics["room_messages"], strconv.FormatUint(uint64(roomID), 10), jsonPayload)
}

// SendSystemMessage 发送系统消息
func (p *MessageProducer) SendSystemMessage(payload any) error {
	// 创建符合SyncMessage格式的消息
	syncMsg := SyncMessage{
		Type:      "system_message",
		SourceID:  p.instanceID,
		Timestamp: time.Now().Unix(),
		Content:   payload,
	}

	// 序列化消息
	jsonPayload, err := json.Marshal(syncMsg)
	if err != nil {
		return err
	}
	return p.SendMessage(p.topics["system_messages"], "system", jsonPayload)
}

// SendInstanceSyncMessage 发送实例同步消息
func (p *MessageProducer) SendInstanceSyncMessage(msgType string, payload any) error {
	// 创建符合SyncMessage格式的消息
	syncMsg := SyncMessage{
		Type:      msgType,
		SourceID:  p.instanceID,
		Timestamp: time.Now().Unix(),
		Content:   payload,
	}

	// 序列化消息
	jsonPayload, err := json.Marshal(syncMsg)
	if err != nil {
		return err
	}

	return p.SendMessage(p.topics["instance_sync"], "sync", jsonPayload)
}

// SendStatusUpdate 发送状态更新消息
func (p *MessageProducer) SendStatusUpdate(userID uint, status int) error {
	statusPayload := StatusPayload{
		UserID:     userID,
		Status:     strconv.Itoa(status),
		InstanceID: p.instanceID,
	}

	// 创建符合SyncMessage格式的消息
	syncMsg := SyncMessage{
		Type:      "status_update",
		SourceID:  p.instanceID,
		Timestamp: time.Now().Unix(),
		Content:   statusPayload,
	}

	// 序列化消息
	jsonPayload, err := json.Marshal(syncMsg)
	if err != nil {
		return err
	}

	return p.SendMessage(p.topics["online_status"], "status", jsonPayload)
}

// Close 关闭生产者
func (p *MessageProducer) Close() error {
	return p.producer.Close()
}

// GetInstanceID 获取实例ID
func (p *MessageProducer) GetInstanceID() string {
	return p.instanceID
}

// 实现MessageNotifier接口
var _ service.MessageNotifier = (*MessageProducer)(nil)
