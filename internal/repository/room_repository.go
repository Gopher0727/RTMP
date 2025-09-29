package repository

import (
	"context"

	"github.com/google/wire"
	"gorm.io/gorm"

	"github.com/Gopher0727/RTMP/internal/model"
)

// IRoomRepository 房间仓库接口
type IRoomRepository interface {
	Create(ctx context.Context, room *model.Room) error
	GetByID(ctx context.Context, id uint) (*model.Room, error)
	List(ctx context.Context, page, size int) ([]*model.Room, int64, error)
	AddMember(ctx context.Context, roomID, userID uint, role int) error
	RemoveMember(ctx context.Context, roomID, userID uint) error
	GetMembers(ctx context.Context, roomID uint) ([]*model.RoomMember, error)
	IsMember(ctx context.Context, roomID, userID uint) (bool, error)
	GetRoomUsers(ctx context.Context, roomID uint) ([]*model.User, error)
}

// RoomRepository 房间仓库实现
type RoomRepository struct {
	db *gorm.DB
}

// NewRoomRepository 创建房间仓库
func NewRoomRepository(db *gorm.DB) IRoomRepository {
	return &RoomRepository{
		db: db,
	}
}

// Create 创建房间
func (r *RoomRepository) Create(ctx context.Context, room *model.Room) error {
	return r.db.WithContext(ctx).Create(room).Error
}

// GetByID 根据ID获取房间
func (r *RoomRepository) GetByID(ctx context.Context, id uint) (*model.Room, error) {
	var room model.Room
	if err := r.db.WithContext(ctx).First(&room, id).Error; err != nil {
		return nil, err
	}
	return &room, nil
}

// List 获取房间列表
func (r *RoomRepository) List(ctx context.Context, page, size int) ([]*model.Room, int64, error) {
	var rooms []*model.Room
	var total int64

	offset := (page - 1) * size
	if err := r.db.WithContext(ctx).Model(&model.Room{}).Count(&total).Error; err != nil {
		return nil, 0, err
	}

	if err := r.db.WithContext(ctx).Offset(offset).Limit(size).Find(&rooms).Error; err != nil {
		return nil, 0, err
	}

	return rooms, total, nil
}

// AddMember 添加房间成员
func (r *RoomRepository) AddMember(ctx context.Context, roomID, userID uint, role int) error {
	member := &model.RoomMember{
		RoomID: roomID,
		UserID: userID,
		Role:   role,
	}
	return r.db.WithContext(ctx).Create(member).Error
}

// RemoveMember 移除房间成员
func (r *RoomRepository) RemoveMember(ctx context.Context, roomID, userID uint) error {
	return r.db.WithContext(ctx).
		Where("room_id = ? AND user_id = ?", roomID, userID).
		Delete(&model.RoomMember{}).Error
}

// GetMembers 获取房间成员
func (r *RoomRepository) GetMembers(ctx context.Context, roomID uint) ([]*model.RoomMember, error) {
	var members []*model.RoomMember
	if err := r.db.WithContext(ctx).Where("room_id = ?", roomID).Find(&members).Error; err != nil {
		return nil, err
	}
	return members, nil
}

// IsMember 检查用户是否是房间成员
func (r *RoomRepository) IsMember(ctx context.Context, roomID, userID uint) (bool, error) {
	var count int64
	err := r.db.WithContext(ctx).
		Model(&model.RoomMember{}).
		Where("room_id = ? AND user_id = ?", roomID, userID).
		Count(&count).Error
	return count > 0, err
}

// GetRoomUsers 获取房间内的所有用户
func (r *RoomRepository) GetRoomUsers(ctx context.Context, roomID uint) ([]*model.User, error) {
	var users []*model.User
	if err := r.db.WithContext(ctx).Table("users").
		Joins("JOIN room_members ON room_members.user_id = users.id").
		Where("room_members.room_id = ?", roomID).
		Find(&users).Error; err != nil {
		return nil, err
	}
	return users, nil
}

// RoomRepositorySet 房间仓库依赖注入
var RoomRepositorySet = wire.NewSet(NewRoomRepository)
