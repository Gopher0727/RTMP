package repository

import (
	"context"

	"github.com/google/wire"
	"gorm.io/gorm"

	"github.com/Gopher0727/RTMP/internal/model"
)

// IMessageRepository 消息仓库接口
type IMessageRepository interface {
	Create(ctx context.Context, message *model.Message) error
	GetByID(ctx context.Context, id uint) (*model.Message, error)
	GetUserMessages(ctx context.Context, userID uint, page, size int) ([]*model.Message, int64, error)
	GetRoomMessages(ctx context.Context, roomID uint, page, size int) ([]*model.Message, int64, error)
	MarkAsRead(ctx context.Context, messageIDs []uint) error
}

// MessageRepository 消息仓库实现
type MessageRepository struct {
	db *gorm.DB
}

// NewMessageRepository 创建消息仓库
func NewMessageRepository(db *gorm.DB) IMessageRepository {
	return &MessageRepository{
		db: db,
	}
}

// Create 创建消息
func (r *MessageRepository) Create(ctx context.Context, message *model.Message) error {
	return r.db.WithContext(ctx).Create(message).Error
}

// GetByID 根据ID获取消息
func (r *MessageRepository) GetByID(ctx context.Context, id uint) (*model.Message, error) {
	var message model.Message
	if err := r.db.WithContext(ctx).First(&message, id).Error; err != nil {
		return nil, err
	}
	return &message, nil
}

// GetUserMessages 获取用户消息
func (r *MessageRepository) GetUserMessages(ctx context.Context, userID uint, page, size int) ([]*model.Message, int64, error) {
	var messages []*model.Message
	var total int64

	offset := (page - 1) * size
	query := r.db.WithContext(ctx).Model(&model.Message{}).
		Where("target_type = ? AND target_id = ?", model.MessageTargetUser, userID)

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	if err := query.Order("created_at DESC").Offset(offset).Limit(size).Find(&messages).Error; err != nil {
		return nil, 0, err
	}

	return messages, total, nil
}

// GetRoomMessages 获取房间消息
func (r *MessageRepository) GetRoomMessages(ctx context.Context, roomID uint, page, size int) ([]*model.Message, int64, error) {
	var messages []*model.Message
	var total int64

	offset := (page - 1) * size
	query := r.db.WithContext(ctx).Model(&model.Message{}).
		Where("target_type = ? AND target_id = ?", model.MessageTargetRoom, roomID)

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	if err := query.Order("created_at DESC").Offset(offset).Limit(size).Find(&messages).Error; err != nil {
		return nil, 0, err
	}

	return messages, total, nil
}

// MarkAsRead 标记消息为已读
func (r *MessageRepository) MarkAsRead(ctx context.Context, messageIDs []uint) error {
	return r.db.WithContext(ctx).Model(&model.Message{}).
		Where("id IN ?", messageIDs).
		Update("is_read", true).Error
}

// MessageRepositorySet 消息仓库依赖注入
var MessageRepositorySet = wire.NewSet(NewMessageRepository)
