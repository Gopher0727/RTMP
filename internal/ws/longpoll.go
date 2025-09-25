package ws

import (
	"sync"
)

// LongPollHub 提供简单的长轮询回退机制：
// - Subscribe(id) 在 map 中为该 id 创建一个通道并返回，等待消息
// - Send(id, data) 若存在等待通道则发送并清理
// 这是一个轻量实现，适合单实例演示。若需跨实例请使用 Redis/Kafka 等中间件。

type LongPollHub struct {
	mu      sync.Mutex
	clients map[string]chan []byte
}

// NewLongPollHub 创建 LongPollHub
func NewLongPollHub() *LongPollHub {
	return &LongPollHub{
		clients: make(map[string]chan []byte),
	}
}

// 包级默认实例，router 可以直接使用 ws.LP
var LP = NewLongPollHub()

// Subscribe 为指定 id 创建一个通道并返回该通道。
// 调用方应当在退出时调用 Unsubscribe 清理资源。
func (h *LongPollHub) Subscribe(id string) <-chan []byte {
	h.mu.Lock()
	defer h.mu.Unlock()
	ch := make(chan []byte, 1)
	// 若已有订阅，先关闭并替换（保证只有一个挂起请求）
	if old, ok := h.clients[id]; ok {
		close(old)
	}
	h.clients[id] = ch
	return ch
}

// Unsubscribe 移除并关闭指定 id 的通道（若存在）。
func (h *LongPollHub) Unsubscribe(id string) {
	h.mu.Lock()
	defer h.mu.Unlock()
	if ch, ok := h.clients[id]; ok {
		delete(h.clients, id)
		close(ch)
	}
}

// Send 向指定 id 的等待通道发送数据；若存在则发送并返回 true，否则返回 false。
func (h *LongPollHub) Send(id string, data []byte) bool {
	h.mu.Lock()
	ch, ok := h.clients[id]
	if ok {
		delete(h.clients, id)
	}
	h.mu.Unlock()
	if !ok {
		return false
	}

	select {
	case ch <- data:
		// 发送成功后关闭通道以通知订阅方并释放资源
		close(ch)
		return true
	default:
		// 若发送失败也关闭通道
		close(ch)
		return false
	}
}
