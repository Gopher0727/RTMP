package api

import (
	"context"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/google/wire"

	"github.com/Gopher0727/RTMP/internal/model"
	"github.com/Gopher0727/RTMP/internal/service"
	"github.com/Gopher0727/RTMP/internal/utils"
)

// RoomHandler 房间处理器
type RoomHandler struct {
	roomService service.RoomService
}

// NewRoomHandler 创建房间处理器
func NewRoomHandler(roomService service.RoomService) *RoomHandler {
	return &RoomHandler{
		roomService: roomService,
	}
}

// CreateRoomRequest 创建房间请求
type CreateRoomRequest struct {
	Name        string `json:"name" binding:"required"`
	Description string `json:"description"`
	IsPrivate   bool   `json:"is_private"`
}

// RoomResponse 房间响应
type RoomResponse struct {
	ID          uint   `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	CreatorID   uint   `json:"creator_id"`
	IsPrivate   bool   `json:"is_private"`
	CreatedAt   string `json:"created_at"`
}

// ListRoomsRequest 获取房间列表请求
type ListRoomsRequest struct {
	Page int `form:"page,default=1" binding:"min=1"`
	Size int `form:"size,default=10" binding:"min=1,max=100"`
}

// ListRoomsResponse 获取房间列表响应
type ListRoomsResponse struct {
	Rooms []*RoomResponse `json:"rooms"`
	Total int64           `json:"total"`
	Page  int             `json:"page"`
	Size  int             `json:"size"`
}

// RoomMemberResponse 房间成员响应
type RoomMemberResponse struct {
	ID       uint   `json:"id"`
	RoomID   uint   `json:"room_id"`
	UserID   uint   `json:"user_id"`
	Role     int    `json:"role"`
	JoinedAt string `json:"joined_at"`
}

// CreateRoom godoc
// @Summary 创建房间
// @Description 创建新的聊天房间
// @Tags rooms
// @Accept json
// @Produce json
// @Param request body CreateRoomRequest true "创建房间请求"
// @Success 200 {object} utils.Response{data=RoomResponse}
// @Failure 400 {object} utils.Response
// @Failure 500 {object} utils.Response
// @Security BearerAuth
// @Router /api/v1/rooms [post]
func (h *RoomHandler) CreateRoom(c *gin.Context) {
	var req CreateRoomRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ResponseBadRequest(c, "参数错误")
		return
	}

	// 从JWT中获取用户信息
	_, exists := c.Get("username")
	if !exists {
		utils.ResponseUnauthorized(c, "未授权")
		return
	}

	// 这里简化处理，实际应该从用户服务获取用户ID
	creatorID := uint(1) // 临时处理

	room := &model.Room{
		Name:        req.Name,
		Description: req.Description,
		CreatorID:   creatorID,
		IsPrivate:   req.IsPrivate,
	}

	ctx := context.Background()
	err := h.roomService.CreateRoom(ctx, room)
	if err != nil {
		utils.ResponseInternalError(c, "创建房间失败")
		return
	}

	resp := &RoomResponse{
		ID:          room.ID,
		Name:        room.Name,
		Description: room.Description,
		CreatorID:   room.CreatorID,
		IsPrivate:   room.IsPrivate,
		CreatedAt:   room.CreatedAt.Format("2006-01-02 15:04:05"),
	}

	utils.ResponseSuccess(c, resp)
}

// GetRoom godoc
// @Summary 获取房间信息
// @Description 根据房间ID获取房间详细信息
// @Tags rooms
// @Accept json
// @Produce json
// @Param id path int true "房间ID"
// @Success 200 {object} utils.Response{data=RoomResponse}
// @Failure 400 {object} utils.Response
// @Failure 404 {object} utils.Response
// @Failure 500 {object} utils.Response
// @Security BearerAuth
// @Router /api/v1/rooms/{id} [get]
func (h *RoomHandler) GetRoom(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		utils.ResponseBadRequest(c, "无效的房间ID")
		return
	}

	ctx := context.Background()
	room, err := h.roomService.GetRoomByID(ctx, uint(id))
	if err != nil {
		if err == service.ErrRoomNotFound {
			utils.ResponseNotFound(c, "房间不存在")
			return
		}
		utils.ResponseInternalError(c, "获取房间信息失败")
		return
	}

	resp := &RoomResponse{
		ID:          room.ID,
		Name:        room.Name,
		Description: room.Description,
		CreatorID:   room.CreatorID,
		IsPrivate:   room.IsPrivate,
		CreatedAt:   room.CreatedAt.Format("2006-01-02 15:04:05"),
	}

	utils.ResponseSuccess(c, resp)
}

// ListRooms godoc
// @Summary 获取房间列表
// @Description 获取房间列表，支持分页
// @Tags rooms
// @Accept json
// @Produce json
// @Param page query int false "页码" default(1)
// @Param size query int false "每页数量" default(10)
// @Success 200 {object} utils.Response{data=ListRoomsResponse}
// @Failure 400 {object} utils.Response
// @Failure 500 {object} utils.Response
// @Security BearerAuth
// @Router /api/v1/rooms [get]
func (h *RoomHandler) ListRooms(c *gin.Context) {
	var req ListRoomsRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		utils.ResponseBadRequest(c, "参数错误")
		return
	}

	ctx := context.Background()
	rooms, total, err := h.roomService.ListRooms(ctx, req.Page, req.Size)
	if err != nil {
		utils.ResponseInternalError(c, "获取房间列表失败")
		return
	}

	roomResponses := make([]*RoomResponse, len(rooms))
	for i, room := range rooms {
		roomResponses[i] = &RoomResponse{
			ID:          room.ID,
			Name:        room.Name,
			Description: room.Description,
			CreatorID:   room.CreatorID,
			IsPrivate:   room.IsPrivate,
			CreatedAt:   room.CreatedAt.Format("2006-01-02 15:04:05"),
		}
	}

	resp := &ListRoomsResponse{
		Rooms: roomResponses,
		Total: total,
		Page:  req.Page,
		Size:  req.Size,
	}

	utils.ResponseSuccess(c, resp)
}

// AddMemberRequest 添加成员请求
type AddMemberRequest struct {
	UserID uint `json:"user_id" binding:"required"`
	Role   int  `json:"role" binding:"required,oneof=1 2"`
}

// AddMember godoc
// @Summary 添加房间成员
// @Description 向房间添加新成员
// @Tags rooms
// @Accept json
// @Produce json
// @Param id path int true "房间ID"
// @Param request body AddMemberRequest true "添加成员请求"
// @Success 200 {object} utils.Response
// @Failure 400 {object} utils.Response
// @Failure 404 {object} utils.Response
// @Failure 500 {object} utils.Response
// @Security BearerAuth
// @Router /api/v1/rooms/{id}/members [post]
func (h *RoomHandler) AddMember(c *gin.Context) {
	idStr := c.Param("id")
	roomID, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		utils.ResponseBadRequest(c, "无效的房间ID")
		return
	}

	var req AddMemberRequest
	if err = c.ShouldBindJSON(&req); err != nil {
		utils.ResponseBadRequest(c, "参数错误")
		return
	}

	ctx := context.Background()
	err = h.roomService.AddMember(ctx, uint(roomID), req.UserID, req.Role)
	if err != nil {
		if err == service.ErrRoomNotFound {
			utils.ResponseNotFound(c, "房间不存在")
			return
		}
		utils.ResponseInternalError(c, "添加房间成员失败")
		return
	}

	utils.ResponseSuccess(c, nil)
}

// RemoveMember godoc
// @Summary 移除房间成员
// @Description 从房间移除指定成员
// @Tags rooms
// @Accept json
// @Produce json
// @Param id path int true "房间ID"
// @Param user_id path int true "用户ID"
// @Success 200 {object} utils.Response
// @Failure 400 {object} utils.Response
// @Failure 500 {object} utils.Response
// @Security BearerAuth
// @Router /api/v1/rooms/{id}/members/{user_id} [delete]
func (h *RoomHandler) RemoveMember(c *gin.Context) {
	roomIDStr := c.Param("id")
	roomID, err := strconv.ParseUint(roomIDStr, 10, 32)
	if err != nil {
		utils.ResponseBadRequest(c, "无效的房间ID")
		return
	}

	userIDStr := c.Param("user_id")
	userID, err := strconv.ParseUint(userIDStr, 10, 32)
	if err != nil {
		utils.ResponseBadRequest(c, "无效的用户ID")
		return
	}

	ctx := context.Background()
	err = h.roomService.RemoveMember(ctx, uint(roomID), uint(userID))
	if err != nil {
		utils.ResponseInternalError(c, "移除房间成员失败")
		return
	}

	utils.ResponseSuccess(c, nil)
}

// GetMembers godoc
// @Summary 获取房间成员
// @Description 获取指定房间的所有成员
// @Tags rooms
// @Accept json
// @Produce json
// @Param id path int true "房间ID"
// @Success 200 {object} utils.Response{data=[]RoomMemberResponse}
// @Failure 400 {object} utils.Response
// @Failure 404 {object} utils.Response
// @Failure 500 {object} utils.Response
// @Security BearerAuth
// @Router /api/v1/rooms/{id}/members [get]
func (h *RoomHandler) GetMembers(c *gin.Context) {
	idStr := c.Param("id")
	roomID, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		utils.ResponseBadRequest(c, "无效的房间ID")
		return
	}

	ctx := context.Background()
	members, err := h.roomService.GetMembers(ctx, uint(roomID))
	if err != nil {
		if err == service.ErrRoomNotFound {
			utils.ResponseNotFound(c, "房间不存在")
			return
		}
		utils.ResponseInternalError(c, "获取房间成员失败")
		return
	}

	memberResponses := make([]*RoomMemberResponse, len(members))
	for i, member := range members {
		memberResponses[i] = &RoomMemberResponse{
			ID:       member.ID,
			RoomID:   member.RoomID,
			UserID:   member.UserID,
			Role:     member.Role,
			JoinedAt: member.CreatedAt.Format("2006-01-02 15:04:05"),
		}
	}

	utils.ResponseSuccess(c, memberResponses)
}

// RoomHandlerSet 房间处理器依赖注入
var RoomHandlerSet = wire.NewSet(NewRoomHandler)
