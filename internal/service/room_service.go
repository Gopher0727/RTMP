package service

import (
	"context"

	"github.com/google/wire"

	"github.com/Gopher0727/RTMP/internal/model"
	"github.com/Gopher0727/RTMP/internal/repository"
)

// IRoomService 房间服务接口
type IRoomService interface {
	CreateRoom(ctx context.Context, room *model.Room) error
	GetRoomByID(ctx context.Context, id uint) (*model.Room, error)
	ListRooms(ctx context.Context, page, size int) ([]*model.Room, int64, error)
	AddMember(ctx context.Context, roomID, userID uint, role string) error
	RemoveMember(ctx context.Context, roomID, userID uint) error
	GetMembers(ctx context.Context, roomID uint) ([]*model.RoomMember, error)
	IsMember(ctx context.Context, roomID, userID uint) (bool, error)
}

// RoomService 房间服务实现
type RoomService struct {
	roomRepo repository.RoomRepository
}

// NewRoomService 创建房间服务
func NewRoomService(roomRepo repository.RoomRepository) *RoomService {
	return &RoomService{
		roomRepo: roomRepo,
	}
}

// CreateRoom 创建房间
func (s *RoomService) CreateRoom(ctx context.Context, room *model.Room) error {
	return s.roomRepo.Create(ctx, room)
}

// GetRoomByID 根据ID获取房间
func (s *RoomService) GetRoomByID(ctx context.Context, id uint) (*model.Room, error) {
	return s.roomRepo.GetByID(ctx, id)
}

// ListRooms 获取房间列表
func (s *RoomService) ListRooms(ctx context.Context, page, size int) ([]*model.Room, int64, error) {
	return s.roomRepo.List(ctx, page, size)
}

// AddMember 添加房间成员
func (s *RoomService) AddMember(ctx context.Context, roomID, userID uint, role int) error {
	// 检查房间是否存在
	_, err := s.roomRepo.GetByID(ctx, roomID)
	if err != nil {
		return ErrRoomNotFound
	}

	return s.roomRepo.AddMember(ctx, roomID, userID, role)
}

// RemoveMember 移除房间成员
func (s *RoomService) RemoveMember(ctx context.Context, roomID, userID uint) error {
	return s.roomRepo.RemoveMember(ctx, roomID, userID)
}

// GetMembers 获取房间成员
func (s *RoomService) GetMembers(ctx context.Context, roomID uint) ([]*model.RoomMember, error) {
	// 检查房间是否存在
	_, err := s.roomRepo.GetByID(ctx, roomID)
	if err != nil {
		return nil, ErrRoomNotFound
	}

	return s.roomRepo.GetMembers(ctx, roomID)
}

// IsMember 检查用户是否是房间成员
func (s *RoomService) IsMember(ctx context.Context, roomID, userID uint) (bool, error) {
	return s.roomRepo.IsMember(ctx, roomID, userID)
}

// RoomServiceSet 房间服务依赖注入
var RoomServiceSet = wire.NewSet(NewRoomService)
