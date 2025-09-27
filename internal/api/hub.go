package api

import (
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/wire"
	"github.com/gorilla/websocket"

	"github.com/Gopher0727/RTMP/internal/model"
	"github.com/Gopher0727/RTMP/internal/service"
)

// HubHandler Hub API处理器
type HubHandler struct {
	hubService     service.IHubService
	userService    service.IUserService
	messageService service.IMessageService
	roomService    service.IRoomService
}

// NewHubHandler 创建Hub API处理器
func NewHubHandler(
	hubService service.IHubService,
	userService service.IUserService,
	messageService service.IMessageService,
	roomService service.IRoomService,
) *HubHandler {
	return &HubHandler{
		hubService:     hubService,
		userService:    userService,
		messageService: messageService,
		roomService:    roomService,
	}
}

// WebSocketHandler WebSocket连接处理
func (h *HubHandler) WebSocketHandler(c *gin.Context) {
	// 从上下文中获取用户ID
	userIDStr := c.GetString("userID")
	userID, err := strconv.ParseUint(userIDStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的用户ID"})
		return
	}

	// 升级HTTP连接为WebSocket
	upgrader := websocket.Upgrader{
		CheckOrigin: func(r *http.Request) bool {
			return true // 允许所有跨域请求，实际应用中应该限制
		},
	}
	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "WebSocket连接失败"})
		return
	}

	// 创建WebSocket客户端
	client := service.NewWSClient(uint(userID), conn)

	// 注册客户端
	if err := h.hubService.Register(c, client); err != nil {
		conn.Close()
		return
	}

	// 启动读写循环
	go h.readPump(client)
	go h.writePump(client)
}

// LongPollingHandler HTTP长轮询处理
func (h *HubHandler) LongPollingHandler(c *gin.Context) {
	// 从上下文中获取用户ID
	userIDStr := c.GetString("userID")
	userID, err := strconv.ParseUint(userIDStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的用户ID"})
		return
	}

	// 创建HTTP长轮询客户端
	client := service.NewHTTPClient(uint(userID))

	// 注册客户端
	if err := h.hubService.Register(c, client); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "注册客户端失败"})
		return
	}

	// 设置超时时间
	timeout := time.After(30 * time.Second)

	// 等待消息或超时
	select {
	case msg := <-client.SendQueue:
		c.Data(http.StatusOK, "application/json", msg)
	case <-timeout:
		c.JSON(http.StatusOK, gin.H{"status": "timeout"})
	case <-client.Ctx.Done():
		c.JSON(http.StatusOK, gin.H{"status": "canceled"})
	}

	// 注销客户端
	h.hubService.Unregister(c, uint(userID))
}

// GetOnlineUsers 获取在线用户列表
func (h *HubHandler) GetOnlineUsers(c *gin.Context) {
	users, err := h.hubService.GetOnlineUsers(c)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "获取在线用户失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"users": users})
}

// SendMessage 发送消息
func (h *HubHandler) SendMessage(c *gin.Context) {
	var req struct {
		TargetType  model.MessageTarget `json:"target_type" binding:"required"`
		TargetID    uint                `json:"target_id" binding:"required"`
		Content     string              `json:"content" binding:"required"`
		MessageType model.MessageType   `json:"message_type" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的请求参数"})
		return
	}

	// 从上下文中获取发送者ID
	senderIDStr := c.GetString("userID")
	senderID, err := strconv.ParseUint(senderIDStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的用户ID"})
		return
	}

	// 创建消息
	message := &model.Message{
		Content:    req.Content,
		Type:       req.MessageType,
		SenderID:   uint(senderID),
		TargetType: req.TargetType,
		TargetID:   req.TargetID,
		InstanceID: "instance-1", // 实际应用中应该从配置获取
		IsRead:     false,
	}

	// 发送消息
	if req.TargetType == model.MessageTargetRoom {
		if err := h.hubService.BroadcastToRoom(c, req.TargetID, message); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "发送房间消息失败"})
			return
		}
	} else {
		if err := h.hubService.SendMessage(c, message); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "发送消息失败"})
			return
		}
	}

	c.JSON(http.StatusOK, gin.H{"message": "消息发送成功"})
}

// readPump 处理WebSocket读取
func (h *HubHandler) readPump(client *service.Client) {
	defer func() {
		h.hubService.Unregister(client.Ctx, client.UserID)
		client.Conn.Close()
	}()

	client.Conn.SetReadLimit(1024 * 1024) // 1MB
	client.Conn.SetReadDeadline(time.Now().Add(60 * time.Second))
	client.Conn.SetPongHandler(func(string) error {
		client.Conn.SetReadDeadline(time.Now().Add(60 * time.Second))
		return nil
	})

	for {
		_, message, err := client.Conn.ReadMessage()
		if err != nil {
			break
		}

		// 处理接收到的消息
		// 实际应用中应该解析消息并调用相应的服务
		_ = message
	}
}

// writePump 处理WebSocket写入
func (h *HubHandler) writePump(client *service.Client) {
	ticker := time.NewTicker(54 * time.Second)
	defer func() {
		ticker.Stop()
		client.Conn.Close()
	}()

	for {
		select {
		case <-ticker.C:
			if err := client.Conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		case <-client.Ctx.Done():
			return
		}
	}
}

// HubHandlerSet Hub处理器依赖注入
var HubHandlerSet = wire.NewSet(NewHubHandler)
