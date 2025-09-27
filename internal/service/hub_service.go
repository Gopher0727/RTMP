package service

import (
	"context"
	"sync"
	"time"

	"github.com/google/wire"
	"github.com/gorilla/websocket"
	"gorm.io/gorm"

	"github.com/Gopher0727/RTMP/internal/model"
	"github.com/Gopher0727/RTMP/internal/repository"
)

// Client 客户端连接
type Client struct {
	UserID     uint
	IsWS       bool
	Conn       *websocket.Conn // WebSocket 连接，可为空
	SendQueue  chan []byte     // HTTP 长轮询客户端用于暂存消息
	LastActive time.Time       // 上次活跃时间，用于心跳或超时清理
	Ctx        context.Context
	Cancel     context.CancelFunc
}

// NewWSClient 创建WebSocket客户端
func NewWSClient(userID uint, conn *websocket.Conn) *Client {
	ctx, cancel := context.WithCancel(context.Background())
	return &Client{
		UserID:     userID,
		IsWS:       true,
		Conn:       conn,
		LastActive: time.Now(),
		Ctx:        ctx,
		Cancel:     cancel,
	}
}

// NewHTTPClient 创建HTTP长轮询客户端
func NewHTTPClient(userID uint) *Client {
	ctx, cancel := context.WithCancel(context.Background())
	return &Client{
		UserID:     userID,
		IsWS:       false,
		SendQueue:  make(chan []byte, 100),
		LastActive: time.Now(),
		Ctx:        ctx,
		Cancel:     cancel,
	}
}

// IHubService Hub服务接口
type IHubService interface {
	Register(ctx context.Context, client *Client) error
	Unregister(ctx context.Context, userID uint) error
	IsOnline(ctx context.Context, userID uint) (bool, string, error)
	GetOnlineUsers(ctx context.Context) ([]*model.User, error)
	SendMessage(ctx context.Context, message *model.Message) error
	BroadcastToRoom(ctx context.Context, roomID uint, message *model.Message) error
}

// HubService Hub服务实现
type HubService struct {
	userRepo    repository.IUserRepository
	messageRepo repository.IMessageRepository
	roomRepo    repository.IRoomRepository
	db          *gorm.DB

	// 本地内存中的客户端连接
	mu      sync.RWMutex
	clients map[uint]*Client // userID -> Client
}

// NewHubService 创建Hub服务
func NewHubService(
	userRepo repository.IUserRepository,
	messageRepo repository.IMessageRepository,
	roomRepo repository.IRoomRepository,
	db *gorm.DB,
) IHubService {
	return &HubService{
		userRepo:    userRepo,
		messageRepo: messageRepo,
		roomRepo:    roomRepo,
		db:          db,
		clients:     make(map[uint]*Client),
	}
}

// Register 注册客户端
func (h *HubService) Register(ctx context.Context, client *Client) error {
	// 更新用户状态为在线
	instanceID := "instance-1" // 实际应用中应该从配置获取
	if err := h.userRepo.UpdateStatus(ctx, client.UserID, model.UserStatusOnline, instanceID); err != nil {
		return err
	}

	// 保存客户端连接到本地内存
	h.mu.Lock()
	defer h.mu.Unlock()
	h.clients[client.UserID] = client

	return nil
}

// Unregister 注销客户端
func (h *HubService) Unregister(ctx context.Context, userID uint) error {
	// 更新用户状态为离线
	if err := h.userRepo.UpdateStatus(ctx, userID, model.UserStatusOffline, ""); err != nil {
		return err
	}

	// 从本地内存中移除客户端连接
	h.mu.Lock()
	defer h.mu.Unlock()
	if client, ok := h.clients[userID]; ok {
		if client.Cancel != nil {
			client.Cancel() // 取消相关的goroutine
		}
		delete(h.clients, userID)
	}

	return nil
}

// IsOnline 查询用户是否在线
func (h *HubService) IsOnline(ctx context.Context, userID uint) (bool, string, error) {
	user, err := h.userRepo.GetByID(ctx, userID)
	if err != nil {
		return false, "", err
	}

	return user.Status == model.UserStatusOnline, user.InstanceID, nil
}

// GetOnlineUsers 获取在线用户列表
func (h *HubService) GetOnlineUsers(ctx context.Context) ([]*model.User, error) {
	// 实际应用中应该添加分页和过滤条件
	var users []*model.User
	if err := h.db.WithContext(ctx).Where("status = ?", model.UserStatusOnline).Find(&users).Error; err != nil {
		return nil, err
	}
	return users, nil
}

// SendMessage 发送消息
func (h *HubService) SendMessage(ctx context.Context, message *model.Message) error {
	// 1. 保存消息到数据库
	if err := h.messageRepo.Create(ctx, message); err != nil {
		return err
	}

	// 2. 如果是单用户消息，检查用户是否在线
	if message.TargetType == model.MessageTargetUser {
		online, instanceID, err := h.IsOnline(ctx, message.TargetID)
		if err != nil {
			return err
		}

		// 3. 如果用户在线且在当前实例，直接发送
		if online && instanceID == "instance-1" { // 实际应用中应该从配置获取
			h.mu.RLock()
			client, ok := h.clients[message.TargetID]
			h.mu.RUnlock()

			if ok {
				h.deliverToClient(client, message)
			}
		}
		// 如果用户在其他实例，通过Kafka发送（在实际应用中实现）
	}

	return nil
}

// BroadcastToRoom 广播消息到房间
func (h *HubService) BroadcastToRoom(ctx context.Context, roomID uint, message *model.Message) error {
	// 1. 保存消息到数据库
	if err := h.messageRepo.Create(ctx, message); err != nil {
		return err
	}

	// 2. 获取房间成员
	members, err := h.roomRepo.GetMembers(ctx, roomID)
	if err != nil {
		return err
	}

	// 3. 遍历房间成员，发送消息
	for _, member := range members {
		// 跳过发送者自己
		if member.UserID == message.SenderID {
			continue
		}

		// 检查用户是否在线
		online, instanceID, err := h.IsOnline(ctx, member.UserID)
		if err != nil {
			continue
		}

		// 如果用户在线且在当前实例，直接发送
		if online && instanceID == "instance-1" { // 实际应用中应该从配置获取
			h.mu.RLock()
			client, ok := h.clients[member.UserID]
			h.mu.RUnlock()

			if ok {
				h.deliverToClient(client, message)
			}
		}
		// 如果用户在其他实例，通过Kafka发送（在实际应用中实现）
	}

	return nil
}

// deliverToClient 将消息发送给客户端
func (h *HubService) deliverToClient(client *Client, message *model.Message) {
	// 实际应用中应该序列化消息
	data := []byte("消息内容") // 示例，实际应用中应该序列化message对象

	if client.IsWS && client.Conn != nil {
		// WebSocket 发送
		client.Conn.WriteMessage(websocket.TextMessage, data)
	} else if !client.IsWS && client.SendQueue != nil {
		// HTTP长轮询发送
		select {
		case client.SendQueue <- data:
		default:
			// 队列满，可丢弃或记录
		}
	}
}

// HubServiceSet Hub服务依赖注入
var HubServiceSet = wire.NewSet(NewHubService)
