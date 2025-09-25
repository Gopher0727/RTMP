package ws

import (
	"log"
)

// Hub 管理所有活跃的 websocket 连接（Client），负责以下职责：
// - 接收客户端注册与注销事件
// - 接收要广播或定向推送的消息并分发到对应客户端
// - 作为单一协程对 clients map 进行读写，避免并发访问竞争

// 设计要点：
// - 通过 channel 进行协程间通信，Run() 在单独 goroutine 中循环处理事件
// - PushMessage 接收外部消息并发送到 broadcast 通道，由 Run() 负责分发
// - 支持基于 Message.To 字段的单播，以及基于 Room 的简单分组广播

type Hub struct {
	register   chan *Client // 注册新客户端
	unregister chan *Client // 注销客户端

	broadcast chan *Message // 从客户端或外部系统收到的消息，等待分发

	clients map[string]*Client         // 活跃客户端集合，key 为 client.id（由客户端在创建时设置）
	rooms   map[string]map[string]bool // 可选：对房间进行简单管理（room -> map[clientID]bool）
}

// NewHub 创建并返回一个未启动的 Hub
func NewHub() *Hub {
	return &Hub{
		register:   make(chan *Client),
		unregister: make(chan *Client),
		broadcast:  make(chan *Message, 256),
		clients:    make(map[string]*Client),
		rooms:      make(map[string]map[string]bool),
	}
}

// Run 启动 Hub 的事件循环：处理注册、注销、广播消息等。
// 必须在一个独立 goroutine 中运行，例如：go hub.Run()
func (h *Hub) Run() {
	for {
		select {
		case c := <-h.register:
			// 新客户端注册
			h.clients[c.id] = c
			// 如果客户端有指定房间，则加入房间映射
			if c.room != "" {
				if _, ok := h.rooms[c.room]; !ok {
					h.rooms[c.room] = make(map[string]bool)
				}
				h.rooms[c.room][c.id] = true
			}
			log.Printf("ws: client registered: %s", c.id)

		case c := <-h.unregister:
			// 客户端注销：从 clients map 和 rooms 中移除
			if _, ok := h.clients[c.id]; ok {
				delete(h.clients, c.id)
				if c.room != "" {
					if rm, ok := h.rooms[c.room]; ok {
						delete(rm, c.id)
						if len(rm) == 0 {
							delete(h.rooms, c.room)
						}
					}
				}
				close(c.send)
				log.Printf("ws: client unregistered: %s", c.id)
			}

		case msg := <-h.broadcast:
			// 根据 Message 内容决定广播策略：
			// - If To != "" -> 单播（定向）
			// - Else if Room != "" -> 房间广播
			// - Else -> 全局广播
			if msg == nil {
				continue
			}

			// 单播
			if msg.To != "" {
				if c, ok := h.clients[msg.To]; ok {
					b, err := msg.ToJSON()
					if err != nil {
						log.Printf("ws: failed marshal message: %v", err)
						continue
					}
					select {
					case c.send <- b:
					default:
						// 若发送阻塞，说明客户端可能卡住，强制断开
						go func(cl *Client) { h.unregister <- cl }(c)
					}
				}
				continue
			}

			// 房间广播
			if msg.Room != "" {
				if rm, ok := h.rooms[msg.Room]; ok {
					b, err := msg.ToJSON()
					if err != nil {
						log.Printf("ws: failed marshal message: %v", err)
						continue
					}
					for cid := range rm {
						if c, ok := h.clients[cid]; ok {
							select {
							case c.send <- b:
							default:
								go func(cl *Client) { h.unregister <- cl }(c)
							}
						}
					}
				}
				continue
			}

			// 全局广播
			b, err := msg.ToJSON()
			if err != nil {
				log.Printf("ws: failed marshal message: %v", err)
				continue
			}
			for _, c := range h.clients {
				select {
				case c.send <- b:
				default:
					go func(cl *Client) { h.unregister <- cl }(c)
				}
			}
		}
	}
}

// Register 将客户端加入注册通道
func (h *Hub) Register(c *Client) {
	h.register <- c
}

// Unregister 将客户端加入注销通道
func (h *Hub) Unregister(c *Client) {
	h.unregister <- c
}

// PushMessage 将消息推入广播通道，由 Run 负责分发
func (h *Hub) PushMessage(m *Message) {
	if m == nil {
		return
	}

	select {
	case h.broadcast <- m:
	default:
		// 若 broadcast 队列已满，丢弃最旧或最新消息；此处选择丢弃新消息并记录日志
		log.Printf("ws: broadcast queue full, dropping message from %s", m.From)
	}
}
