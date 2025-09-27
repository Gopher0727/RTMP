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
func SetupRouter(r *gin.Engine) {
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
		v1.POST("/auth/register", api.Register)
		v1.POST("/auth/login", api.Login)

		// 需要认证的路由
		auth := v1.Group("/")
		auth.Use(middleware.JWTAuth())
		{
			// 用户相关
			auth.GET("/users", api.GetUsers)
			auth.GET("/users/:id", api.GetUser)
			auth.PUT("/users/:id", api.UpdateUser)

			// 消息相关
			auth.POST("/messages", api.SendMessage)
			auth.GET("/messages", api.GetMessages)
			auth.GET("/messages/:id", api.GetMessage)
			auth.PUT("/messages/:id/read", api.MarkMessageAsRead)

			// 房间相关
			auth.POST("/rooms", api.CreateRoom)
			auth.GET("/rooms", api.GetRooms)
			auth.GET("/rooms/:id", api.GetRoom)
			auth.POST("/rooms/:id/members", api.AddRoomMember)
			auth.DELETE("/rooms/:id/members/:userId", api.RemoveRoomMember)
			auth.GET("/rooms/:id/members", api.GetRoomMembers)

			// WebSocket连接
			auth.GET("/ws", api.WebSocketHandler)

			// HTTP长轮询
			auth.GET("/poll", api.LongPollingHandler)

			// 在线用户
			auth.GET("/online", api.GetOnlineUsers)
		}
	}
}
