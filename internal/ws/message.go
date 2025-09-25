package ws

import (
	"encoding/json"
	"time"
)

// Message 是 websocket 连接之间传递的统一消息结构。
// 设计为通用 JSON 包装：类型、来源、目标、所属房间、负载与时间戳。
// 使用 json.RawMessage 保持负载的灵活性（可以是任意 JSON 对象）。
type Message struct {
	Type    string          `json:"type"`              // 消息类型，如 "message"、"system"、"join"
	From    string          `json:"from,omitempty"`    // 发送者 ID（可选）
	To      string          `json:"to,omitempty"`      // 目标用户 ID（可选，针对单播）
	Room    string          `json:"room,omitempty"`    // 频道/房间（可选，针对房间广播）
	Payload json.RawMessage `json:"payload,omitempty"` // 业务负载，保持原始 JSON
	Ts      int64           `json:"ts,omitempty"`      // Unix 时间戳（秒）
}

// NewMessage 创建一个带当前时间戳的 Message
func NewMessage(typ, from, to, room string, payload interface{}) (*Message, error) {
	var raw json.RawMessage
	if payload != nil {
		b, err := json.Marshal(payload)
		if err != nil {
			return nil, err
		}
		raw = json.RawMessage(b)
	}
	return &Message{
		Type:    typ,
		From:    from,
		To:      to,
		Room:    room,
		Payload: raw,
		Ts:      time.Now().Unix(),
	}, nil
}

// ToJSON 将 Message 编码成字节切片，便于通过 websocket 传输或写入 hub 的 broadcast 通道。
func (m *Message) ToJSON() ([]byte, error) {
	return json.Marshal(m)
}

// MessageFromJSON 从 JSON 字节解码出 Message
func MessageFromJSON(b []byte) (*Message, error) {
	var m Message
	if err := json.Unmarshal(b, &m); err != nil {
		return nil, err
	}
	return &m, nil
}
