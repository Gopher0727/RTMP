//go:build wireinject
// +build wireinject

package internal

import (
	"github.com/google/wire"
	"gorm.io/gorm"

	"github.com/Gopher0727/RTMP/config"
	"github.com/Gopher0727/RTMP/internal/api"
	"github.com/Gopher0727/RTMP/internal/repository"
	"github.com/Gopher0727/RTMP/internal/service"
)

// InitApp 初始化应用依赖
func InitApp(db *gorm.DB) (*App, error) {
	wire.Build(
		// 仓库层
		repository.UserRepositorySet,
		repository.MessageRepositorySet,
		repository.RoomRepositorySet,

		// 服务层
		service.UserServiceSet,
		service.MessageServiceSet,
		service.RoomServiceSet,
		service.HubServiceSet,

		// API处理器层
		api.AuthHandlerSet,
		api.UserHandlerSet,
		api.MessageHandlerSet,
		api.RoomHandlerSet,

		// 配置
		wire.Value(&config.Config{}),

		// 应用
		NewApp,
	)
	return nil, nil
}

// App 应用结构体
type App struct {
	// 服务层
	UserService    service.UserService
	MessageService service.MessageService
	RoomService    service.RoomService
	HubService     service.HubService

	// API处理器层
	AuthHandler    *api.AuthHandler
	UserHandler    *api.UserHandler
	MessageHandler *api.MessageHandler
	RoomHandler    *api.RoomHandler

	// 配置
	Config *config.Config
}

// NewApp 创建应用
func NewApp(
	userService service.UserService,
	messageService service.MessageService,
	roomService service.RoomService,
	hubService service.HubService,
	authHandler *api.AuthHandler,
	userHandler *api.UserHandler,
	messageHandler *api.MessageHandler,
	roomHandler *api.RoomHandler,
	config *config.Config,
) *App {
	return &App{
		UserService:    userService,
		MessageService: messageService,
		RoomService:    roomService,
		HubService:     hubService,
		AuthHandler:    authHandler,
		UserHandler:    userHandler,
		MessageHandler: messageHandler,
		RoomHandler:    roomHandler,
		Config:         config,
	}
}
