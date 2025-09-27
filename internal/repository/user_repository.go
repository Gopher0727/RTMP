package repository

import (
	"context"

	"github.com/google/wire"
	"gorm.io/gorm"

	"github.com/Gopher0727/RTMP/internal/model"
)

// UserRepository 用户仓库接口
type UserRepository interface {
	Create(ctx context.Context, user *model.User) error
	GetByID(ctx context.Context, id uint) (*model.User, error)
	GetByUsername(ctx context.Context, username string) (*model.User, error)
	UpdateStatus(ctx context.Context, id uint, status int, instanceID string) error
	List(ctx context.Context, page, size int) ([]*model.User, int64, error)
}

// UserRepositoryImpl 用户仓库实现
type UserRepositoryImpl struct {
	db *gorm.DB
}

// NewUserRepository 创建用户仓库
func NewUserRepository(db *gorm.DB) UserRepository {
	return &UserRepositoryImpl{db: db}
}

// Create 创建用户
func (r *UserRepositoryImpl) Create(ctx context.Context, user *model.User) error {
	return r.db.WithContext(ctx).Create(user).Error
}

// GetByID 根据ID获取用户
func (r *UserRepositoryImpl) GetByID(ctx context.Context, id uint) (*model.User, error) {
	var user model.User
	if err := r.db.WithContext(ctx).First(&user, id).Error; err != nil {
		return nil, err
	}
	return &user, nil
}

// GetByUsername 根据用户名获取用户
func (r *UserRepositoryImpl) GetByUsername(ctx context.Context, username string) (*model.User, error) {
	var user model.User
	if err := r.db.WithContext(ctx).Where("username = ?", username).First(&user).Error; err != nil {
		return nil, err
	}
	return &user, nil
}

// UpdateStatus 更新用户状态
func (r *UserRepositoryImpl) UpdateStatus(ctx context.Context, id uint, status int, instanceID string) error {
	return r.db.WithContext(ctx).Model(&model.User{}).Where("id = ?", id).
		Updates(map[string]any{
			"status":      status,
			"instance_id": instanceID,
		}).Error
}

// List 获取用户列表
func (r *UserRepositoryImpl) List(ctx context.Context, page, size int) ([]*model.User, int64, error) {
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

// UserRepositorySet 用户仓库依赖注入
var UserRepositorySet = wire.NewSet(NewUserRepository)
