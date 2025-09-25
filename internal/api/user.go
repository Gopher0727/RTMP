package api

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

// CreateUserRequest 是创建用户的简单请求结构（示例）
type CreateUserRequest struct {
	Username string `json:"username" binding:"required"`
	Nick     string `json:"nick,omitempty"`
}

// UserResponse 是返回给客户端的用户信息结构（示例）
type UserResponse struct {
	ID       string `json:"id"`
	Username string `json:"username"`
	Nick     string `json:"nick,omitempty"`
	Created  int64  `json:"created"`
}

// CreateUserHandler 处理用户创建请求（示例实现）。
// - 验证请求体
// - 返回一个模拟的用户对象（在真实项目中应写入数据库并返回实际 ID）
func CreateUserHandler(c *gin.Context) {
	var req CreateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// 生成示例 ID（真实项目请使用可靠的 ID 生成策略或数据库自增）
	id := req.Username
	if id == "" {
		id = "user-" + time.Now().Format("20060102150405")
	}

	resp := UserResponse{
		ID:       id,
		Username: req.Username,
		Nick:     req.Nick,
		Created:  time.Now().Unix(),
	}
	c.JSON(http.StatusCreated, resp)
}

// GetUserHandler 返回指定用户的示例数据。
// 在真实项目中应从数据库查询用户并返回。
func GetUserHandler(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "missing user id"})
		return
	}

	// 返回示例用户（真实项目应查询数据库）
	resp := UserResponse{
		ID:       id,
		Username: id,
		Nick:     "demo",
		Created:  time.Now().Unix(),
	}
	c.JSON(http.StatusOK, resp)
}
