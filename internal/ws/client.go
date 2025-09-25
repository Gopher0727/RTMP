package ws

import (
	"log"
	"time"

	"github.com/gorilla/websocket"
)

// 连接读写超时配置
const (
	writeWait      = 10 * time.Second
	pongWait       = 60 * time.Second
	pingPeriod     = (pongWait * 9) / 10
	maxMessageSize = 5120 // 最大消息尺寸（字节）
)

// Client 表示单个 websocket 连接。
// 由 Hub 创建并管理：
// - send 通道用于将要写回客户端的字节流
// - Read() 从 websocket 读取消息并转发给 Hub
// - Write() 将从 send 通道接收的消息写回 websocket

type Client struct {
	id   string
	hub  *Hub
	conn *websocket.Conn
	send chan []byte
	room string
}

// NewClient 创建并返回一个 Client 实例。
// - conn: 已完成升级的 *websocket.Conn
// - id: 客户端唯一 ID
// - room: 可选房间 ID
// - hub: 注入的 Hub 实例
func NewClient(conn *websocket.Conn, id string, room string, hub *Hub) *Client {
	return &Client{
		id:   id,
		hub:  hub,
		conn: conn,
		send: make(chan []byte, 256),
		room: room,
	}
}

// Read 从 websocket 读取消息并处理：
// - 解码 JSON 为 Message，并将业务消息推入 hub
// - 处理心跳（Pong）
// 当连接关闭或遇到错误时，会向 hub 发出 Unregister
func (c *Client) Read() {
	defer func() {
		c.hub.Unregister(c)
		_ = c.conn.Close()
	}()

	c.conn.SetReadLimit(512 * 1024)
	c.conn.SetReadDeadline(time.Now().Add(60 * time.Second))
	c.conn.SetPongHandler(func(string) error {
		c.conn.SetReadDeadline(time.Now().Add(60 * time.Second))
		return nil
	})

	for {
		_, message, err := c.conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("ws: unexpected close error: %v", err)
			}
			break
		}

		// 解析消息并分发到 Hub
		msg, err := MessageFromJSON(message)
		if err != nil {
			log.Printf("ws: invalid message from %s: %v", c.id, err)
			continue
		}

		// 如果消息没有 From 字段，补上当前客户端 ID
		if msg.From == "" {
			msg.From = c.id
		}

		// 如果消息没有指定目标或房间，则作为全局广播
		c.hub.PushMessage(msg)
	}
}

// Write 从 send 通道读取字节并写入 websocket 连接；也负责心跳（定期发送 Ping）。
// 当 send 通道关闭或出现写入错误时退出并注销客户端。
func (c *Client) Write() {
	ticker := time.NewTicker(54 * time.Second)
	defer func() {
		ticker.Stop()
		_ = c.conn.Close()
	}()

	for {
		select {
		case message, ok := <-c.send:
			c.conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
			if !ok {
				// 通道已被关闭，发送 close 控制消息并返回
				_ = c.conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			w, err := c.conn.NextWriter(websocket.TextMessage)
			if err != nil {
				return
			}
			if _, err := w.Write(message); err != nil {
				_ = w.Close()
				return
			}

			if err := w.Close(); err != nil {
				return
			}

		case <-ticker.C:
			// 定期发送 ping 保持连接
			c.conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
			if err := c.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}
