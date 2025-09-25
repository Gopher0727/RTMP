package api

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/Gopher0727/RTMP/internal/ws"
)

// SendMessageHandler 返回一个处理函数，该处理函数接收 JSON 格式的消息体并将消息推入 Hub。
// 支持全局广播、房间广播（room 字段）和单播（to 字段）。
func SendMessageHandler(hub *ws.Hub) gin.HandlerFunc {
	return func(c *gin.Context) {
		var msg ws.Message
		if err := c.ShouldBindJSON(&msg); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		// 补充时间戳与来源（如果需要）
		if msg.Ts == 0 {
			msg.Ts = time.Now().Unix()
		}

		// 将消息推入 hub 的广播队列，由 Hub 决定分发策略
		hub.PushMessage(&msg)
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	}
}
