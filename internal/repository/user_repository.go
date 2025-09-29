package repository

import (
	"context"

	"gorm.io/gorm"

	"github.com/Gopher0727/RTMP/internal/model"
	"github.com/google/wire"
)

// IUserRepository 用户仓库接口
type IUserRepository interface {
	Create(ctx context.Context, user *model.User) error
	GetByID(ctx context.Context, id uint) (*model.User, error)
	GetByUsername(ctx context.Context, username string) (*model.User, error)
	GetByEmail(ctx context.Context, email string) (*model.User, error)
	UpdateStatus(ctx context.Context, id uint, status int, instanceID string) error
	List(ctx context.Context, page, size int) ([]*model.User, int64, error)
	IsOnline(ctx context.Context, id uint) (bool, string, error)
	GetOnlineUsers(ctx context.Context) ([]*model.User, error)
}

// UserRepository 用户仓库实现
type UserRepository struct {
	db *gorm.DB
}

// NewUserRepository 创建用户仓库
func NewUserRepository(db *gorm.DB) IUserRepository {
	return &UserRepository{
		db: db,
	}
}

// Create 创建用户
func (r *UserRepository) Create(ctx context.Context, user *model.User) error {
	return r.db.WithContext(ctx).Create(user).Error
}

// GetByID 根据ID获取用户
func (r *UserRepository) GetByID(ctx context.Context, id uint) (*model.User, error) {
	var user model.User
	if err := r.db.WithContext(ctx).First(&user, id).Error; err != nil {
		return nil, err
	}
	return &user, nil
}

// GetByUsername 根据用户名获取用户
func (r *UserRepository) GetByUsername(ctx context.Context, username string) (*model.User, error) {
	var user model.User
	if err := r.db.WithContext(ctx).Where("username = ?", username).First(&user).Error; err != nil {
		return nil, err
	}
	return &user, nil
}

// UpdateStatus 更新用户状态
func (r *UserRepository) UpdateStatus(ctx context.Context, id uint, status int, instanceID string) error {
	return r.db.WithContext(ctx).Model(&model.User{}).Where("id = ?", id).
		Updates(map[string]any{
			"status":      status,
			"instance_id": instanceID,
		}).Error
}

// GetByEmail 根据邮箱获取用户
func (r *UserRepository) GetByEmail(ctx context.Context, email string) (*model.User, error) {
	var user model.User
	if err := r.db.WithContext(ctx).Where("email = ?", email).First(&user).Error; err != nil {
		return nil, err
	}
	return &user, nil
}

// List 获取用户列表
func (r *UserRepository) List(ctx context.Context, page, size int) ([]*model.User, int64, error) {
	var users []*model.User
	var total int64

	offset := (page - 1) * size
	if err := r.db.WithContext(ctx).Model(&model.User{}).Count(&total).Error; err != nil {
		return nil, 0, err
	}

	if err := r.db.WithContext(ctx).Offset(offset).Limit(size).Find(&users).Error; err != nil {
		return nil, 0, err
	}

	return users, total, nil
}

// IsOnline 检查用户是否在线
func (r *UserRepository) IsOnline(ctx context.Context, id uint) (bool, string, error) {
	var user model.User
	if err := r.db.WithContext(ctx).Select("status, instance_id").Where("id = ?", id).First(&user).Error; err != nil {
		return false, "", err
	}
	return user.Status == model.UserStatusOnline, user.InstanceID, nil
}

// GetOnlineUsers 获取在线用户列表
func (r *UserRepository) GetOnlineUsers(ctx context.Context) ([]*model.User, error) {
	var users []*model.User
	if err := r.db.WithContext(ctx).Where("status = ?", model.UserStatusOnline).Find(&users).Error; err != nil {
		return nil, err
	}
	return users, nil
}

// UserRepositorySet 用户仓库依赖注入
var UserRepositorySet = wire.NewSet(NewUserRepository)
