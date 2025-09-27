package service

import (
	"context"
	"errors"

	"github.com/google/wire"
	"gorm.io/gorm"

	"github.com/Gopher0727/RTMP/internal/model"
	"github.com/Gopher0727/RTMP/internal/repository"
	"github.com/Gopher0727/RTMP/internal/utils"
)

// IUserService 用户服务接口
type IUserService interface {
	Register(ctx context.Context, username, password string) (*model.User, error)
	Login(ctx context.Context, username, password string) (*model.User, error)
	GetUserByID(ctx context.Context, id uint) (*model.User, error)
	UpdateUserStatus(ctx context.Context, id uint, status int, instanceID string) error
	ListUsers(ctx context.Context, page, size int) ([]*model.User, int64, error)
}

// UserServiceImp 用户服务实现
type UserServiceImp struct {
	userRepo repository.IUserRepository
}

// NewUserService 创建用户服务
func NewUserService(userRepo repository.IUserRepository) IUserService {
	return &UserServiceImp{
		userRepo: userRepo,
	}
}

// Register 用户注册
func (s *UserServiceImp) Register(ctx context.Context, username, password string) (*model.User, error) {
	// 检查用户是否已存在
	existUser, err := s.userRepo.GetByUsername(ctx, username)
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}
	if existUser != nil {
		return nil, ErrUserAlreadyExists
	}

	// 对密码进行哈希处理
	hashedPassword, err := utils.HashPassword(password)
	if err != nil {
		return nil, err
	}

	// 创建新用户
	user := &model.User{
		Username: username,
		Password: hashedPassword,
		Status:   model.UserStatusOffline,
	}

	if err := s.userRepo.Create(ctx, user); err != nil {
		return nil, err
	}

	return user, nil
}

// Login 用户登录
func (s *UserServiceImp) Login(ctx context.Context, username, password string) (*model.User, error) {
	// 获取用户
	user, err := s.userRepo.GetByUsername(ctx, username)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrUserNotFound
		}
		return nil, err
	}

	// 验证密码
	if !utils.CheckPasswordHash(password, user.Password) {
		return nil, ErrInvalidPassword
	}

	return user, nil
}

// GetUserByID 根据ID获取用户
func (s *UserServiceImp) GetUserByID(ctx context.Context, id uint) (*model.User, error) {
	return s.userRepo.GetByID(ctx, id)
}

// UpdateUserStatus 更新用户状态
func (s *UserServiceImp) UpdateUserStatus(ctx context.Context, id uint, status int, instanceID string) error {
	return s.userRepo.UpdateStatus(ctx, id, status, instanceID)
}

// ListUsers 获取用户列表
func (s *UserServiceImp) ListUsers(ctx context.Context, page, size int) ([]*model.User, int64, error) {
	return s.userRepo.List(ctx, page, size)
}

// UserServiceSet 用户服务依赖注入
var UserServiceSet = wire.NewSet(NewUserService)
