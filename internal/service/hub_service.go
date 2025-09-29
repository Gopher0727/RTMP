package service

import (
	"context"
	"encoding/json"
	"log"
	"sync"
	"time"

	"github.com/google/wire"
	"github.com/gorilla/websocket"
	"gorm.io/gorm"

	"github.com/Gopher0727/RTMP/internal/model"
	"github.com/Gopher0727/RTMP/internal/repository"
)

// MessageNotifier 消息通知接口，用于解耦Kafka依赖
type MessageNotifier interface {
	SendUserMessage(userID uint, message *model.Message) error
	SendRoomMessage(roomID uint, message *model.Message) error
	SendStatusUpdate(userID uint, status int) error
	GetInstanceID() string
}

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
	SetMessageNotifier(notifier MessageNotifier)
}

// HubService Hub服务实现
type HubService struct {
	userRepo        repository.IUserRepository
	messageRepo     repository.IMessageRepository
	roomRepo        repository.IRoomRepository
	db              *gorm.DB
	instanceID      string
	messageNotifier MessageNotifier

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
		instanceID:  "unknown", // 初始为unknown，后续通过SetMessageNotifier更新
		clients:     make(map[uint]*Client),
	}
}

// SetMessageNotifier 设置消息通知器
func (h *HubService) SetMessageNotifier(notifier MessageNotifier) {
	h.mu.Lock()
	h.messageNotifier = notifier
	if notifier != nil {
		h.instanceID = notifier.GetInstanceID()
	}
	h.mu.Unlock()
}

// Register 注册客户端
func (h *HubService) Register(ctx context.Context, client *Client) error {
	// 更新用户状态为在线
	if err := h.userRepo.UpdateStatus(ctx, client.UserID, model.UserStatusOnline, h.instanceID); err != nil {
		return err
	}

	// 保存客户端连接到本地内存
	h.mu.Lock()
	h.clients[client.UserID] = client
	h.mu.Unlock()

	// 发送用户上线状态到消息通知器
	if h.messageNotifier != nil {
		go func() {
			if err := h.messageNotifier.SendStatusUpdate(client.UserID, model.UserStatusOnline); err != nil {
				log.Printf("Failed to send online status: %v", err)
			}
		}()
	}

	log.Printf("Client registered: UserID=%d, IsWS=%v, InstanceID=%s", client.UserID, client.IsWS, h.instanceID)

	return nil
}

// Unregister 注销客户端
func (h *HubService) Unregister(ctx context.Context, userID uint) error {
	h.mu.Lock()
	client, exists := h.clients[userID]
	if exists {
		client.Cancel() // 取消客户端上下文
		if client.IsWS && client.Conn != nil {
			client.Conn.Close()
		}
		delete(h.clients, userID)
	}
	h.mu.Unlock()

	// 更新用户状态为离线
	if err := h.userRepo.UpdateStatus(ctx, userID, model.UserStatusOffline, ""); err != nil {
		log.Printf("Failed to update user status to offline: %v", err)
	}

	// 发送用户下线状态到消息通知器
	if h.messageNotifier != nil && exists {
		go func() {
			if err := h.messageNotifier.SendStatusUpdate(userID, model.UserStatusOffline); err != nil {
				log.Printf("Failed to send offline status: %v", err)
			}
		}()
	}

	log.Printf("Client unregistered: UserID=%d, InstanceID=%s", userID, h.instanceID)
	return nil
}

// IsOnline 检查用户是否在线
func (h *HubService) IsOnline(ctx context.Context, userID uint) (bool, string, error) {
	// 首先检查本地实例是否有此用户连接
	h.mu.RLock()
	_, exists := h.clients[userID]
	h.mu.RUnlock()

	if exists {
		return true, h.instanceID, nil
	}

	// 从数据库查询用户状态（可能在其他实例上在线）
	return h.userRepo.IsOnline(ctx, userID)
}

// GetOnlineUsers 获取在线用户列表
func (h *HubService) GetOnlineUsers(ctx context.Context) ([]*model.User, error) {
	return h.userRepo.GetOnlineUsers(ctx)
}

// SendMessage 发送消息给指定用户
func (h *HubService) SendMessage(ctx context.Context, message *model.Message) error {
	// 保存消息到数据库
	if err := h.messageRepo.Create(ctx, message); err != nil {
		return err
	}

	// 检查消息接收者是否在当前实例
	h.mu.RLock()
	client, exists := h.clients[message.ReceiverID]
	h.mu.RUnlock()

	// 如果用户在当前实例，则直接发送消息
	if exists {
		// 序列化消息
		msgBytes, err := json.Marshal(message)
		if err != nil {
			return err
		}

		// 根据客户端类型发送消息
		if client.IsWS {
			// WebSocket客户端直接发送
			if err := client.Conn.WriteMessage(websocket.TextMessage, msgBytes); err != nil {
				// 如果发送失败，可能是连接已断开，需要注销客户端
				log.Printf("Failed to send message via WebSocket, unregistering client: %v", err)
				go h.Unregister(ctx, client.UserID)
			}
		} else {
			// HTTP长轮询客户端放入发送队列
			select {
			case client.SendQueue <- msgBytes:
			default:
				// 队列已满，可能需要处理
				// todo
				log.Printf("Send queue full for user %d", client.UserID)
			}
		}

		// 更新最后活跃时间
		client.LastActive = time.Now()
	}

	// 发送消息到消息通知器
	if h.messageNotifier != nil {
		go func() {
			if err := h.messageNotifier.SendUserMessage(message.ReceiverID, message); err != nil {
				log.Printf("Failed to send message to notifier: %v", err)
			}
		}()
	}

	return nil
}

// BroadcastToRoom 向房间内所有用户广播消息
func (h *HubService) BroadcastToRoom(ctx context.Context, roomID uint, message *model.Message) error {
	// 保存消息到数据库
	if err := h.messageRepo.Create(ctx, message); err != nil {
		return err
	}

	// 获取房间内的所有用户
	roomUsers, err := h.roomRepo.GetRoomUsers(ctx, roomID)
	if err != nil {
		return err
	}

	// 序列化消息
	msgBytes, err := json.Marshal(message)
	if err != nil {
		return err
	}

	// 向当前实例中在房间内的用户发送消息
	h.mu.RLock()
	for _, user := range roomUsers {
		if client, exists := h.clients[user.ID]; exists {
			if client.IsWS {
				// WebSocket客户端直接发送
				if err := client.Conn.WriteMessage(websocket.TextMessage, msgBytes); err != nil {
					// 如果发送失败，可能是连接已断开，需要注销客户端
					log.Printf("Failed to send room message via WebSocket, unregistering client: %v", err)
					// 复制user.ID以避免在RLock内修改map
					uID := user.ID
					h.mu.RUnlock() // 先释放读锁
					go h.Unregister(ctx, uID)
					h.mu.RLock() // 重新获取读锁继续循环
				}
			} else {
				// HTTP长轮询客户端放入发送队列
				select {
				case client.SendQueue <- msgBytes:
				default:
					// 队列已满，可能需要处理
					// todo
					log.Printf("Send queue full for user %d", client.UserID)
				}
			}

			// 更新最后活跃时间
			client.LastActive = time.Now()
		}
	}
	h.mu.RUnlock()

	// 发送消息到消息通知器
	if h.messageNotifier != nil {
		go func() {
			if err := h.messageNotifier.SendRoomMessage(roomID, message); err != nil {
				log.Printf("Failed to send room message to notifier: %v", err)
			}
		}()
	}

	return nil
}

// HubServiceSet Hub服务依赖注入
var HubServiceSet = wire.NewSet(
	NewHubService,
	wire.Bind(new(IHubService), new(*HubService)),
)
