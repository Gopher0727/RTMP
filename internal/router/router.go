package router

import (
	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"

	_ "github.com/Gopher0727/RTMP/docs" // 导入swagger文档
	"github.com/Gopher0727/RTMP/internal/api"
	"github.com/Gopher0727/RTMP/internal/middleware"
)

// SetupRouter 设置路由
func SetupRouter(r *gin.Engine, authHandler *api.AuthHandler, userHandler *api.UserHandler,
	messageHandler *api.MessageHandler, roomHandler *api.RoomHandler, hubHandler *api.HubHandler) {
	// 全局中间件
	r.Use(middleware.RequestID())
	r.Use(middleware.Logger())
	r.Use(middleware.CORS())
	r.Use(middleware.RateLimit())

	// 健康检查
	r.GET("/health", api.HealthHandler)

	// Swagger文档
	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	// API v1 路由组
	v1 := r.Group("/api/v1")
	{
		// 公开路由
		authGroup := v1.Group("/auth")
		{
			authGroup.POST("/register", authHandler.Register)
			authGroup.POST("/login", authHandler.Login)
			authGroup.POST("/logout", authHandler.Logout)
			authGroup.POST("/refresh", middleware.JWTAuth(), authHandler.RefreshToken)
		}

		// 需要认证的路由
		auth := v1.Group("/")
		auth.Use(middleware.JWTAuth())
		{
			// 用户相关
			auth.GET("/users", userHandler.ListUsers)
			auth.GET("/users/:id", userHandler.GetUser)
			auth.GET("/users/me", userHandler.GetCurrentUser)
			auth.PUT("/users/:id/status", userHandler.UpdateUserStatus)

			// 消息相关
			auth.POST("/messages", messageHandler.SendMessage)
			auth.GET("/messages/user/:user_id", messageHandler.GetUserMessages)
			auth.GET("/messages/room/:room_id", messageHandler.GetRoomMessages)
			auth.PUT("/messages/read", messageHandler.MarkAsRead)

			// 房间相关
			auth.POST("/rooms", roomHandler.CreateRoom)
			auth.GET("/rooms", roomHandler.ListRooms)
			auth.GET("/rooms/:id", roomHandler.GetRoom)
			auth.POST("/rooms/:id/members", roomHandler.AddMember)
			auth.DELETE("/rooms/:id/members/:user_id", roomHandler.RemoveMember)
			auth.GET("/rooms/:id/members", roomHandler.GetMembers)

			// WebSocket连接
			auth.GET("/ws", hubHandler.WebSocketHandler)

			// HTTP长轮询
			auth.GET("/poll", hubHandler.LongPollingHandler)

			// 在线用户
			auth.GET("/online", hubHandler.GetOnlineUsers)
		}
	}
}
