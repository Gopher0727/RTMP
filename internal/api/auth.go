package api

import (
	"context"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/wire"

	"github.com/Gopher0727/RTMP/config"
	"github.com/Gopher0727/RTMP/internal/middleware"
	"github.com/Gopher0727/RTMP/internal/service"
	"github.com/Gopher0727/RTMP/internal/utils"
)

// AuthHandler 认证处理器
type AuthHandler struct {
	userService service.IUserService
}

// NewAuthHandler 创建认证处理器
func NewAuthHandler(userService service.IUserService) *AuthHandler {
	return &AuthHandler{
		userService: userService,
	}
}

// RegisterRequest 注册请求
type RegisterRequest struct {
	Username string `json:"username" binding:"required,min=3,max=20" example:"user123"`
	Password string `json:"password" binding:"required,min=6,max=20" example:"password123"`
	Email    string `json:"email" binding:"required,email" example:"user@example.com"`
	Nickname string `json:"nickname" binding:"max=50" example:"用户昵称"`
}

// LoginRequest 登录请求
type LoginRequest struct {
	Username string `json:"username" binding:"required" example:"user123"`
	Password string `json:"password" binding:"required" example:"password123"`
}

// TokenResponse 令牌响应
type TokenResponse struct {
	Token string    `json:"token" example:"eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."`
	User  *UserInfo `json:"user"`
}

// UserInfo 用户信息
type UserInfo struct {
	ID       uint   `json:"id"`
	Username string `json:"username"`
	Email    string `json:"email"`
	Nickname string `json:"nickname"`
	Avatar   string `json:"avatar"`
}

// Register godoc
// @Summary 用户注册
// @Description 注册新用户账号
// @Tags auth
// @Accept json
// @Produce json
// @Param request body RegisterRequest true "注册信息"
// @Success 201 {object} utils.Response{data=TokenResponse}
// @Failure 400 {object} utils.Response
// @Failure 409 {object} utils.Response
// @Failure 500 {object} utils.Response
// @Router /api/v1/auth/register [post]
func (h *AuthHandler) Register(c *gin.Context) {
	var req RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ResponseBadRequest(c, "参数错误: "+err.Error())
		return
	}

	ctx := context.Background()
	user, err := h.userService.Register(ctx, req.Username, req.Password)
	if err != nil {
		if err == service.ErrUserAlreadyExists {
			c.JSON(http.StatusConflict, utils.Response{
				Code:    http.StatusConflict,
				Message: "用户已存在",
				Data:    nil,
			})
			return
		}
		utils.ResponseInternalError(c, "注册失败")
		return
	}

	// 生成JWT令牌
	token, err := middleware.GenerateToken(user.Username, config.GetJWTConfig())
	if err != nil {
		utils.ResponseInternalError(c, "生成令牌失败")
		return
	}

	userInfo := &UserInfo{
		ID:       user.ID,
		Username: user.Username,
		Email:    user.Email,
		Nickname: user.Nickname,
		Avatar:   user.Avatar,
	}

	resp := &TokenResponse{
		Token: token,
		User:  userInfo,
	}

	c.JSON(http.StatusCreated, utils.Response{
		Code:    http.StatusCreated,
		Message: "注册成功",
		Data:    resp,
	})
}

// Login godoc
// @Summary 用户登录
// @Description 用户登录获取访问令牌
// @Tags auth
// @Accept json
// @Produce json
// @Param request body LoginRequest true "登录信息"
// @Success 200 {object} utils.Response{data=TokenResponse}
// @Failure 400 {object} utils.Response
// @Failure 401 {object} utils.Response
// @Failure 404 {object} utils.Response
// @Failure 500 {object} utils.Response
// @Router /api/v1/auth/login [post]
func (h *AuthHandler) Login(c *gin.Context) {
	var req LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ResponseBadRequest(c, "参数错误: "+err.Error())
		return
	}

	ctx := context.Background()
	user, err := h.userService.Login(ctx, req.Username, req.Password)
	if err != nil {
		if err == service.ErrUserNotFound {
			utils.ResponseNotFound(c, "用户不存在")
			return
		}
		if err == service.ErrInvalidPassword {
			utils.ResponseUnauthorized(c, "密码错误")
			return
		}
		utils.ResponseInternalError(c, "登录失败")
		return
	}

	// 生成JWT令牌
	token, err := middleware.GenerateToken(user.Username, config.GetJWTConfig())
	if err != nil {
		utils.ResponseInternalError(c, "生成令牌失败")
		return
	}

	userInfo := &UserInfo{
		ID:       user.ID,
		Username: user.Username,
		Email:    user.Email,
		Nickname: user.Nickname,
		Avatar:   user.Avatar,
	}

	resp := &TokenResponse{
		Token: token,
		User:  userInfo,
	}

	utils.ResponseSuccess(c, resp)
}

// Logout godoc
// @Summary 用户登出
// @Description 用户登出（客户端删除令牌）
// @Tags auth
// @Accept json
// @Produce json
// @Success 200 {object} utils.Response
// @Security BearerAuth
// @Router /api/v1/auth/logout [post]
func (h *AuthHandler) Logout(c *gin.Context) {
	// JWT是无状态的，登出主要由客户端处理（删除令牌）
	// 这里可以添加令牌黑名单逻辑（如果需要的话）
	utils.ResponseSuccess(c, gin.H{"message": "登出成功"})
}

// RefreshToken godoc
// @Summary 刷新令牌
// @Description 刷新访问令牌
// @Tags auth
// @Accept json
// @Produce json
// @Success 200 {object} utils.Response{data=TokenResponse}
// @Failure 401 {object} utils.Response
// @Failure 500 {object} utils.Response
// @Security BearerAuth
// @Router /api/v1/auth/refresh [post]
func (h *AuthHandler) RefreshToken(c *gin.Context) {
	// 从JWT中获取用户信息
	username, exists := c.Get("username")
	if !exists {
		utils.ResponseUnauthorized(c, "未授权")
		return
	}

	// 生成新的JWT令牌
	token, err := middleware.GenerateToken(username.(string), config.GetJWTConfig())
	if err != nil {
		utils.ResponseInternalError(c, "生成令牌失败")
		return
	}

	// 获取用户信息
	ctx := context.Background()
	user, err := h.userService.GetUserByID(ctx, 1) // 这里需要从JWT中获取用户ID
	if err != nil {
		utils.ResponseInternalError(c, "获取用户信息失败")
		return
	}

	userInfo := &UserInfo{
		ID:       user.ID,
		Username: user.Username,
		Email:    user.Email,
		Nickname: user.Nickname,
		Avatar:   user.Avatar,
	}

	resp := &TokenResponse{
		Token: token,
		User:  userInfo,
	}

	utils.ResponseSuccess(c, resp)
}

// AuthHandlerSet 认证处理器依赖注入
var AuthHandlerSet = wire.NewSet(NewAuthHandler)
