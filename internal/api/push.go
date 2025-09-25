package api

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/Gopher0727/RTMP/internal/ws"
)

// PushHandler 提供一个兼容的推送接口，实现方式与 SendMessageHandler 相同。
func PushHandler(hub *ws.Hub) gin.HandlerFunc {
	return func(c *gin.Context) {
		var msg ws.Message
		if err := c.ShouldBindJSON(&msg); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		if msg.Ts == 0 {
			msg.Ts = time.Now().Unix()
		}

		hub.PushMessage(&msg)
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	}
}
