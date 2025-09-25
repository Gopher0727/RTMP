package router

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"

	"github.com/Gopher0727/RTMP/internal/ws"
)

// RegisterRoutes 将应用路由注册到传入的 gin 引擎，并注入 Hub 实例。
// 如果传入 jwtMiddleware（非 nil），则在 /ws 路由上应用该中间件进行鉴权。
// - /ws: WebSocket 端点，接受 ?id=... & room=... 的查询参数
func RegisterRoutes(r *gin.Engine, hub *ws.Hub, jwtMiddleware gin.HandlerFunc) {
	// 健康检查
	r.GET("/ping", func(ctx *gin.Context) {
		ctx.JSON(http.StatusOK, gin.H{
			"message": "pong",
		})
	})

	// 简单服务列表接口，便于调试或测试脚本使用
	r.GET("/services", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"services": []string{"ws", "longpoll", "api"},
		})
	})

	upgrader := websocket.Upgrader{
		CheckOrigin: func(r *http.Request) bool { return true },
	}

	wsHandler := func(c *gin.Context) {
		// 完成 websocket 升级
		conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
		if err != nil {
			log.Printf("ws: upgrade failed: %v", err)
			c.AbortWithStatus(http.StatusBadRequest)
			return
		}

		// 获取客户端 ID 与房间参数；若未提供 ID 则生成一个短 ID
		id := c.Query("id")
		if id == "" {
			id = fmt.Sprintf("%s-%d", c.ClientIP(), time.Now().UnixNano())
		}
		room := c.Query("room")

		client := ws.NewClient(conn, id, room, hub)

		// 注册到 hub 并启动读写循环：
		hub.Register(client)
		go client.Write()
		// Read 在当前协程阻塞，当 Read 返回时表示连接已关闭，控制权回到这里
		client.Read()
	}

	// 根据是否提供 jwtMiddleware 决定如何注册 /ws
	if jwtMiddleware != nil {
		r.GET("/ws", jwtMiddleware, wsHandler)
	} else {
		r.GET("/ws", wsHandler)
	}

	// 长轮询接口：客户端发起 GET /longpoll?id=... 将在服务器端等待消息或超时返回
	r.GET("/longpoll", func(c *gin.Context) {
		id := c.Query("id")
		if id == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "missing id"})
			return
		}

		ch := ws.LP.Subscribe(id)
		defer ws.LP.Unsubscribe(id)

		select {
		case data, ok := <-ch:
			if !ok {
				c.Status(http.StatusNoContent)
				return
			}
			c.Data(http.StatusOK, "application/octet-stream", data)
		case <-time.After(60 * time.Second):
			// 超时返回 204 表示无内容，客户端可立即重试
			c.Status(http.StatusNoContent)
		}
	})

	// 发送到长轮询客户端：POST /send?to=clientid
	r.POST("/send", func(c *gin.Context) {
		to := c.Query("to")
		if to == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "missing to"})
			return
		}
		data, err := io.ReadAll(c.Request.Body)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		if ws.LP.Send(to, data) {
			c.JSON(http.StatusOK, gin.H{"status": "sent"})
			return
		}
		c.JSON(http.StatusNotFound, gin.H{"error": "no subscriber"})
	})
}
