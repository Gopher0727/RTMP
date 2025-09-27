package api

import (
	"context"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/google/wire"

	"github.com/Gopher0727/RTMP/internal/service"
	"github.com/Gopher0727/RTMP/internal/utils"
)

// UserHandler 用户处理器
type UserHandler struct {
	userService service.IUserService
}

// NewUserHandler 创建用户处理器
func NewUserHandler(userService service.IUserService) *UserHandler {
	return &UserHandler{
		userService: userService,
	}
}

// GetUserRequest 获取用户请求
type GetUserRequest struct {
	ID uint `uri:"id" binding:"required"`
}

// GetUserResponse 获取用户响应
type GetUserResponse struct {
	ID       uint   `json:"id"`
	Username string `json:"username"`
	Email    string `json:"email"`
	Nickname string `json:"nickname"`
	Avatar   string `json:"avatar"`
	Status   int    `json:"status"`
}

// ListUsersRequest 获取用户列表请求
type ListUsersRequest struct {
	Page int `form:"page,default=1" binding:"min=1"`
	Size int `form:"size,default=10" binding:"min=1,max=100"`
}

// ListUsersResponse 获取用户列表响应
type ListUsersResponse struct {
	Users []*GetUserResponse `json:"users"`
	Total int64              `json:"total"`
	Page  int                `json:"page"`
	Size  int                `json:"size"`
}

// GetUser godoc
// @Summary 获取用户信息
// @Description 根据用户ID获取用户信息
// @Tags users
// @Accept json
// @Produce json
// @Param id path int true "用户ID"
// @Success 200 {object} utils.Response{data=GetUserResponse}
// @Failure 400 {object} utils.Response
// @Failure 404 {object} utils.Response
// @Failure 500 {object} utils.Response
// @Security BearerAuth
// @Router /api/v1/users/{id} [get]
func (h *UserHandler) GetUser(c *gin.Context) {
	var req GetUserRequest
	if err := c.ShouldBindUri(&req); err != nil {
		utils.ResponseBadRequest(c, "无效的用户ID")
		return
	}

	ctx := context.Background()
	user, err := h.userService.GetUserByID(ctx, req.ID)
	if err != nil {
		if err == service.ErrUserNotFound {
			utils.ResponseNotFound(c, "用户不存在")
			return
		}
		utils.ResponseInternalError(c, "获取用户信息失败")
		return
	}

	resp := &GetUserResponse{
		ID:       user.ID,
		Username: user.Username,
		Email:    user.Email,
		Nickname: user.Nickname,
		Avatar:   user.Avatar,
		Status:   user.Status,
	}

	utils.ResponseSuccess(c, resp)
}

// ListUsers godoc
// @Summary 获取用户列表
// @Description 获取用户列表，支持分页
// @Tags users
// @Accept json
// @Produce json
// @Param page query int false "页码" default(1)
// @Param size query int false "每页数量" default(10)
// @Success 200 {object} utils.Response{data=ListUsersResponse}
// @Failure 400 {object} utils.Response
// @Failure 500 {object} utils.Response
// @Security BearerAuth
// @Router /api/v1/users [get]
func (h *UserHandler) ListUsers(c *gin.Context) {
	var req ListUsersRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		utils.ResponseBadRequest(c, "参数错误")
		return
	}

	ctx := context.Background()
	users, total, err := h.userService.ListUsers(ctx, req.Page, req.Size)
	if err != nil {
		utils.ResponseInternalError(c, "获取用户列表失败")
		return
	}

	userResponses := make([]*GetUserResponse, len(users))
	for i, user := range users {
		userResponses[i] = &GetUserResponse{
			ID:       user.ID,
			Username: user.Username,
			Email:    user.Email,
			Nickname: user.Nickname,
			Avatar:   user.Avatar,
			Status:   user.Status,
		}
	}

	resp := &ListUsersResponse{
		Users: userResponses,
		Total: total,
		Page:  req.Page,
		Size:  req.Size,
	}

	utils.ResponseSuccess(c, resp)
}

// UpdateUserStatusRequest 更新用户状态请求
type UpdateUserStatusRequest struct {
	Status     int    `json:"status" binding:"required,oneof=0 1"`
	InstanceID string `json:"instance_id"`
}

// UpdateUserStatus godoc
// @Summary 更新用户状态
// @Description 更新用户在线状态
// @Tags users
// @Accept json
// @Produce json
// @Param id path int true "用户ID"
// @Param request body UpdateUserStatusRequest true "更新用户状态请求"
// @Success 200 {object} utils.Response
// @Failure 400 {object} utils.Response
// @Failure 404 {object} utils.Response
// @Failure 500 {object} utils.Response
// @Security BearerAuth
// @Router /api/v1/users/{id}/status [put]
func (h *UserHandler) UpdateUserStatus(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		utils.ResponseBadRequest(c, "无效的用户ID")
		return
	}

	var req UpdateUserStatusRequest
	if err = c.ShouldBindJSON(&req); err != nil {
		utils.ResponseBadRequest(c, "参数错误")
		return
	}

	ctx := context.Background()
	err = h.userService.UpdateUserStatus(ctx, uint(id), req.Status, req.InstanceID)
	if err != nil {
		if err == service.ErrUserNotFound {
			utils.ResponseNotFound(c, "用户不存在")
			return
		}
		utils.ResponseInternalError(c, "更新用户状态失败")
		return
	}

	utils.ResponseSuccess(c, nil)
}

// UserHandlerSet 用户处理器依赖注入
var UserHandlerSet = wire.NewSet(NewUserHandler)
