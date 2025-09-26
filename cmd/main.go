package main

import (
	"fmt"
	"log"

	"github.com/gin-gonic/gin"

	"github.com/Gopher0727/RTMP/config"
)

func main() {
	cfg, err := config.LoadConfig("config.toml")
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("AppName: %s, Env: %s\n", cfg.AppName, cfg.Env)

	// 创建 Gin 引擎
	r := gin.New()
	r.Use(gin.Logger())

	// 启动 HTTP 服务，使用配置中的端口（若未设置则回退到 :8080）
	addr := ":8080"
	if cfg.Server.Port != 0 {
		addr = fmt.Sprintf(":%d", cfg.Server.Port)
	}
	if err := r.Run(addr); err != nil {
		log.Fatalf("failed to run server: %v", err)
	}
}
