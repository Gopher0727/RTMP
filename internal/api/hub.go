package api

import (
	"encoding/json"
	"log"
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
		MessageType string `json:"message_type" binding:"required,oneof=user room"`
		TargetID    uint   `json:"target_id" binding:"required"`
		Content     string `json:"content" binding:"required"`
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
		Content:  req.Content,
		SenderID: uint(senderID),
		IsRead:   false,
	}

	// 根据消息类型设置目标
	if req.MessageType == "user" {
		message.Type = "user"
		message.ReceiverID = req.TargetID
		// 发送私聊消息
		if err := h.hubService.SendMessage(c, message); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "发送消息失败"})
			return
		}
	} else {
		// 房间消息
		message.Type = "room"
		message.RoomID = req.TargetID
		// 广播房间消息
		if err := h.hubService.BroadcastToRoom(c, req.TargetID, message); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "发送房间消息失败"})
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
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("WebSocket读取错误: %v", err)
			}
			break
		}

		// 解析消息
		var msg struct {
			Type        string `json:"type"`         // 消息类型: message, ping等
			TargetID    uint   `json:"target_id"`    // 目标ID: 房间ID或用户ID
			Content     string `json:"content"`      // 消息内容
			SenderID    uint   `json:"sender_id"`    // 发送者ID
			MessageType string `json:"message_type"` // 内部消息类型: user, room
		}

		if err := json.Unmarshal(message, &msg); err != nil {
			log.Printf("解析WebSocket消息失败: %v", err)
			continue
		}

		// 处理不同类型的消息
		switch msg.Type {
		case "ping":
			// 处理心跳消息
			client.LastActive = time.Now()
			if err := client.Conn.WriteMessage(websocket.PongMessage, nil); err != nil {
				log.Printf("发送Pong消息失败: %v", err)
			}
		case "message":
			// 创建消息对象
			messageObj := &model.Message{
				SenderID:  msg.SenderID,
				Content:   msg.Content,
				CreatedAt: time.Now(),
				UpdatedAt: time.Now(),
				IsRead:    false,
			}

			// 根据消息类型设置接收者
			if msg.MessageType == "user" {
				// 私聊消息
				messageObj.Type = "user"
				messageObj.ReceiverID = msg.TargetID
			} else {
				// 房间消息
				messageObj.Type = "room"
				messageObj.RoomID = msg.TargetID
			}

			// 通过hubService发送消息，这会通过Kafka广播到其他实例
			if err := h.hubService.SendMessage(client.Ctx, messageObj); err != nil {
				log.Printf("发送消息失败: %v", err)
			}
		default:
			log.Printf("未知的消息类型: %s", msg.Type)
		}
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
