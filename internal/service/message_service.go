package service

import (
	"context"

	"github.com/google/wire"

	"github.com/Gopher0727/RTMP/internal/model"
	"github.com/Gopher0727/RTMP/internal/repository"
)

// IMessageService 消息服务接口
type IMessageService interface {
	SendMessage(ctx context.Context, message *model.Message) error
	GetUserMessages(ctx context.Context, userID uint, page, size int) ([]*model.Message, int64, error)
	GetRoomMessages(ctx context.Context, roomID uint, page, size int) ([]*model.Message, int64, error)
	MarkAsRead(ctx context.Context, messageIDs []uint) error
}

// MessageService 消息服务实现
type MessageService struct {
	messageRepo repository.MessageRepository
	roomRepo    repository.RoomRepository
}

// NewMessageService 创建消息服务
func NewMessageService(messageRepo repository.MessageRepository, roomRepo repository.RoomRepository) *MessageService {
	return &MessageService{
		messageRepo: messageRepo,
		roomRepo:    roomRepo,
	}
}

// SendMessage 发送消息
func (s *MessageService) SendMessage(ctx context.Context, message *model.Message) error {
	// 如果是房间消息，验证发送者是否是房间成员
	if message.TargetType == model.MessageTargetRoom {
		isMember, err := s.roomRepo.IsMember(ctx, message.TargetID, message.SenderID)
		if err != nil {
			return err
		}
		if !isMember {
			return ErrNotRoomMember
		}
	}

	return s.messageRepo.Create(ctx, message)
}

// GetUserMessages 获取用户消息
func (s *MessageService) GetUserMessages(ctx context.Context, userID uint, page, size int) ([]*model.Message, int64, error) {
	return s.messageRepo.GetUserMessages(ctx, userID, page, size)
}

// GetRoomMessages 获取房间消息
func (s *MessageService) GetRoomMessages(ctx context.Context, roomID uint, page, size int) ([]*model.Message, int64, error) {
	return s.messageRepo.GetRoomMessages(ctx, roomID, page, size)
}

// MarkAsRead 标记消息为已读
func (s *MessageService) MarkAsRead(ctx context.Context, messageIDs []uint) error {
	return s.messageRepo.MarkAsRead(ctx, messageIDs)
}

// MessageServiceSet 消息服务依赖注入
var MessageServiceSet = wire.NewSet(NewMessageService)
