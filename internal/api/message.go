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

// MessageHandler 消息处理器
type MessageHandler struct {
	messageService service.IMessageService
}

// NewMessageHandler 创建消息处理器
func NewMessageHandler(messageService service.IMessageService) *MessageHandler {
	return &MessageHandler{
		messageService: messageService,
	}
}

// SendMessageRequest 发送消息请求
type SendMessageRequest struct {
	Content    string              `json:"content" binding:"required"`
	Type       model.MessageType   `json:"type" binding:"required,oneof=1 2 3"`
	TargetType model.MessageTarget `json:"target_type" binding:"required,oneof=1 2"`
	TargetID   uint                `json:"target_id" binding:"required"`
}

// MessageResponse 消息响应
type MessageResponse struct {
	ID         uint                `json:"id"`
	Content    string              `json:"content"`
	Type       model.MessageType   `json:"type"`
	TargetType model.MessageTarget `json:"target_type"`
	TargetID   uint                `json:"target_id"`
	SenderID   uint                `json:"sender_id"`
	SenderName string              `json:"sender_name"`
	IsRead     bool                `json:"is_read"`
	CreatedAt  string              `json:"created_at"`
}

// ListMessagesRequest 获取消息列表请求
type ListMessagesRequest struct {
	Page int `form:"page,default=1" binding:"min=1"`
	Size int `form:"size,default=10" binding:"min=1,max=100"`
}

// ListMessagesResponse 获取消息列表响应
type ListMessagesResponse struct {
	Messages []*MessageResponse `json:"messages"`
	Total    int64              `json:"total"`
	Page     int                `json:"page"`
	Size     int                `json:"size"`
}

// SendMessage godoc
// @Summary 发送消息
// @Description 发送消息到用户或房间
// @Tags messages
// @Accept json
// @Produce json
// @Param request body SendMessageRequest true "发送消息请求"
// @Success 200 {object} utils.Response
// @Failure 400 {object} utils.Response
// @Failure 403 {object} utils.Response
// @Failure 500 {object} utils.Response
// @Security BearerAuth
// @Router /api/v1/messages [post]
func (h *MessageHandler) SendMessage(c *gin.Context) {
	var req SendMessageRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ResponseBadRequest(c, "参数错误")
		return
	}

	// 从JWT中获取用户信息
	username, exists := c.Get("username")
	if !exists {
		utils.ResponseUnauthorized(c, "未授权")
		return
	}

	userID, exists := c.Get("user_id")
	if !exists {
		utils.ResponseUnauthorized(c, "未授权")
		return
	}

	message := &model.Message{
		Content:    req.Content,
		Type:       req.Type,
		TargetType: req.TargetType,
		TargetID:   req.TargetID,
		SenderID:   userID.(uint),
		SenderName: username.(string),
	}

	ctx := context.Background()
	err := h.messageService.SendMessage(ctx, message)
	if err != nil {
		if err == service.ErrNotRoomMember {
			utils.ResponseForbidden(c, "不是房间成员")
			return
		}
		utils.ResponseInternalError(c, "发送消息失败")
		return
	}

	utils.ResponseSuccess(c, nil)
}

// GetUserMessages godoc
// @Summary 获取用户消息
// @Description 获取指定用户的消息列表
// @Tags messages
// @Accept json
// @Produce json
// @Param user_id path int true "用户ID"
// @Param page query int false "页码" default(1)
// @Param size query int false "每页数量" default(10)
// @Success 200 {object} utils.Response{data=ListMessagesResponse}
// @Failure 400 {object} utils.Response
// @Failure 500 {object} utils.Response
// @Security BearerAuth
// @Router /api/v1/messages/user/{user_id} [get]
func (h *MessageHandler) GetUserMessages(c *gin.Context) {
	userIDStr := c.Param("user_id")
	userID, err := strconv.ParseUint(userIDStr, 10, 32)
	if err != nil {
		utils.ResponseBadRequest(c, "无效的用户ID")
		return
	}

	var req ListMessagesRequest
	if err = c.ShouldBindQuery(&req); err != nil {
		utils.ResponseBadRequest(c, "参数错误")
		return
	}

	ctx := context.Background()
	messages, total, err := h.messageService.GetUserMessages(ctx, uint(userID), req.Page, req.Size)
	if err != nil {
		utils.ResponseInternalError(c, "获取用户消息失败")
		return
	}

	messageResponses := make([]*MessageResponse, len(messages))
	for i, msg := range messages {
		messageResponses[i] = &MessageResponse{
			ID:         msg.ID,
			Content:    msg.Content,
			Type:       msg.Type,
			TargetType: msg.TargetType,
			TargetID:   msg.TargetID,
			SenderID:   msg.SenderID,
			SenderName: msg.SenderName,
			IsRead:     msg.IsRead,
			CreatedAt:  msg.CreatedAt.Format("2006-01-02 15:04:05"),
		}
	}

	resp := &ListMessagesResponse{
		Messages: messageResponses,
		Total:    total,
		Page:     req.Page,
		Size:     req.Size,
	}

	utils.ResponseSuccess(c, resp)
}

// GetRoomMessages godoc
// @Summary 获取房间消息
// @Description 获取指定房间的消息列表
// @Tags messages
// @Accept json
// @Produce json
// @Param room_id path int true "房间ID"
// @Param page query int false "页码" default(1)
// @Param size query int false "每页数量" default(10)
// @Success 200 {object} utils.Response{data=ListMessagesResponse}
// @Failure 400 {object} utils.Response
// @Failure 500 {object} utils.Response
// @Security BearerAuth
// @Router /api/v1/messages/room/{room_id} [get]
func (h *MessageHandler) GetRoomMessages(c *gin.Context) {
	roomIDStr := c.Param("room_id")
	roomID, err := strconv.ParseUint(roomIDStr, 10, 32)
	if err != nil {
		utils.ResponseBadRequest(c, "无效的房间ID")
		return
	}

	var req ListMessagesRequest
	if err = c.ShouldBindQuery(&req); err != nil {
		utils.ResponseBadRequest(c, "参数错误")
		return
	}

	ctx := context.Background()
	messages, total, err := h.messageService.GetRoomMessages(ctx, uint(roomID), req.Page, req.Size)
	if err != nil {
		utils.ResponseInternalError(c, "获取房间消息失败")
		return
	}

	messageResponses := make([]*MessageResponse, len(messages))
	for i, msg := range messages {
		messageResponses[i] = &MessageResponse{
			ID:         msg.ID,
			Content:    msg.Content,
			Type:       msg.Type,
			TargetType: msg.TargetType,
			TargetID:   msg.TargetID,
			SenderID:   msg.SenderID,
			SenderName: msg.SenderName,
			IsRead:     msg.IsRead,
			CreatedAt:  msg.CreatedAt.Format("2006-01-02 15:04:05"),
		}
	}

	resp := &ListMessagesResponse{
		Messages: messageResponses,
		Total:    total,
		Page:     req.Page,
		Size:     req.Size,
	}

	utils.ResponseSuccess(c, resp)
}

// MarkAsReadRequest 标记已读请求
type MarkAsReadRequest struct {
	MessageIDs []uint `json:"message_ids" binding:"required"`
}

// MarkAsRead godoc
// @Summary 标记消息已读
// @Description 标记指定消息为已读状态
// @Tags messages
// @Accept json
// @Produce json
// @Param request body MarkAsReadRequest true "标记已读请求"
// @Success 200 {object} utils.Response
// @Failure 400 {object} utils.Response
// @Failure 500 {object} utils.Response
// @Security BearerAuth
// @Router /api/v1/messages/read [put]
func (h *MessageHandler) MarkAsRead(c *gin.Context) {
	var req MarkAsReadRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ResponseBadRequest(c, "参数错误")
		return
	}

	ctx := context.Background()
	err := h.messageService.MarkAsRead(ctx, req.MessageIDs)
	if err != nil {
		utils.ResponseInternalError(c, "标记消息已读失败")
		return
	}

	utils.ResponseSuccess(c, nil)
}

// MessageHandlerSet 消息处理器依赖注入
var MessageHandlerSet = wire.NewSet(NewMessageHandler)
