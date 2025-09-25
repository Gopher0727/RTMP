package api

import (
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/Gopher0727/RTMP/internal/ws"
)

// LoginRequest 为登录请求的简单结构体（示例）
type LoginRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

// LoginResponse 是登录返回的示例结构（包含模拟 token）
type LoginResponse struct {
	Token string `json:"token"`
	User  any    `json:"user"`
}

// loginHandler 是一个简单的模拟登录实现：
// - 验证请求体
// - 返回一个伪造的 token（实际项目请替换为 JWT 或其它安全方案）
func loginHandler(c *gin.Context) {
	var req LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// 这里可以调用真实的用户校验逻辑（DB 查询、密码比对等）
	// 目前示例直接返回一个伪 token
	token := fmt.Sprintf("token-%s-%d", req.Username, time.Now().Unix())
	c.JSON(http.StatusOK, LoginResponse{Token: token, User: gin.H{"id": req.Username}})
}

// RegisterRoutes 将 api 下的 HTTP 路由统一注册到 gin 引擎。
// 传入 hub 以便部分 API（如消息推送）能够把消息发到 WebSocket Hub。
func RegisterRoutes(r *gin.Engine, hub *ws.Hub) {
	api := r.Group("/api")

	// auth
	api.POST("/auth/login", loginHandler)

	// health
	api.GET("/health", HealthHandler)

	// messages: 发送消息到 hub（可广播/单播/房间）
	api.POST("/messages", SendMessageHandler(hub))

	// push: 兼容的推送接口（别名）
	api.POST("/push", PushHandler(hub))

	// user management
	api.POST("/users", CreateUserHandler)
	api.GET("/users/:id", GetUserHandler)
}
