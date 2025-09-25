package main

import (
	"fmt"
	"log"

	"github.com/gin-gonic/gin"

	"github.com/Gopher0727/RTMP/config"
	"github.com/Gopher0727/RTMP/internal/api"
	"github.com/Gopher0727/RTMP/internal/db"
	"github.com/Gopher0727/RTMP/internal/middleware"
	"github.com/Gopher0727/RTMP/internal/router"
	"github.com/Gopher0727/RTMP/internal/ws"
)

func main() {
	cfg, err := config.LoadConfig("config.toml")
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("AppName: %s, Env: %s, MySQLHost: %s\n", cfg.AppName, cfg.Env, cfg.MySQL.Host)

	// 初始化数据库（MySQL）
	if err := db.InitMySQL(cfg.MySQL); err != nil {
		log.Fatalf("failed to init mysql: %v", err)
	}
	// 在程序退出时关闭 DB 连接
	defer func() {
		if err := db.CloseDB(); err != nil {
			log.Printf("failed to close db: %v", err)
		}
	}()

	// 执行自动迁移
	if err := db.Migrate(); err != nil {
		log.Fatalf("failed to migrate db: %v", err)
	}

	// 创建 Gin 引擎
	r := gin.New()
	r.Use(gin.Logger())

	// 创建并启动 Hub（负责管理所有 websocket 连接）
	hub := ws.NewHub()
	go hub.Run()

	// 根据配置决定是否启用 JWT 中间件
	var jwtMW gin.HandlerFunc
	if cfg.JWT.Secret != "" {
		jwtMW = middleware.JWTMiddleware(cfg.JWT)
	} else {
		jwtMW = nil
	}

	// 注册路由：静态客户端、router(ws/longpoll) 与 api
	// 暴露静态示例页面，便于手动测试
	r.StaticFile("/client.html", "web/client.html")

	// 注册 WebSocket /longpoll 路由（router 内部会根据 jwtMW 决定是否保护 /ws）
	router.RegisterRoutes(r, hub, jwtMW)

	// 注册 REST API 路由，并注入 hub 以便推送到 ws
	api.RegisterRoutes(r, hub)

	// 启动 HTTP 服务，使用配置中的端口（若未设置则回退到 :8080）
	addr := ":8080"
	if cfg.Server.Port != 0 {
		addr = fmt.Sprintf(":%d", cfg.Server.Port)
	}
	if err := r.Run(addr); err != nil {
		log.Fatalf("failed to run server: %v", err)
	}
}
