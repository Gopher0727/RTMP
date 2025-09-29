//go:build wireinject
// +build wireinject

package internal

import (
	"github.com/google/wire"
	"gorm.io/gorm"

	"github.com/Gopher0727/RTMP/config"
	"github.com/Gopher0727/RTMP/internal/api"
	"github.com/Gopher0727/RTMP/internal/kafka"
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
		api.HubHandlerSet,

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
	UserService    service.IUserService
	MessageService service.IMessageService
	RoomService    service.IRoomService
	HubService     service.IHubService

	// API处理器层
	AuthHandler    *api.AuthHandler
	UserHandler    *api.UserHandler
	MessageHandler *api.MessageHandler
	RoomHandler    *api.RoomHandler
	HubHandler     *api.HubHandler

	// 配置
	Config *config.Config
}

// NewApp 创建应用
func NewApp(
	userService service.IUserService,
	messageService service.IMessageService,
	roomService service.IRoomService,
	hubService service.IHubService,
	authHandler *api.AuthHandler,
	userHandler *api.UserHandler,
	messageHandler *api.MessageHandler,
	roomHandler *api.RoomHandler,
	hubHandler *api.HubHandler,
	config *config.Config,
) *App {
	// 初始化Kafka生产者
	if err := kafka.InitKafka(config); err != nil {
		// todo
		panic("Failed to initialize Kafka producer: " + err.Error())
	}

	// 初始化Kafka消费者（在服务初始化后）
	if err := kafka.InitConsumer(config, messageService, hubService); err != nil {
		// todo
		panic("Failed to initialize Kafka consumer: " + err.Error())
	}

	return &App{
		UserService:    userService,
		MessageService: messageService,
		RoomService:    roomService,
		HubService:     hubService,
		AuthHandler:    authHandler,
		UserHandler:    userHandler,
		MessageHandler: messageHandler,
		RoomHandler:    roomHandler,
		HubHandler:     hubHandler,
		Config:         config,
	}
}
