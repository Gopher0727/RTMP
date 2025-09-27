package main

import (
	"fmt"
	"log"

	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"

	"github.com/Gopher0727/RTMP/config"
	_ "github.com/Gopher0727/RTMP/docs" // 导入swagger文档
	"github.com/Gopher0727/RTMP/internal/kafka"
	"github.com/Gopher0727/RTMP/internal/middleware"
	"github.com/Gopher0727/RTMP/internal/router"
)

// @title RTMP API
// @version 1.0
// @description 实时消息推送平台API文档
// @termsOfService http://swagger.io/terms/

// @contact.name API Support
// @contact.url http://www.swagger.io/support
// @contact.email support@swagger.io

// @license.name Apache 2.0
// @license.url http://www.apache.org/licenses/LICENSE-2.0.html

// @host localhost:8080
// @BasePath /api/v1
// @schemes http https

// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
// @description Type "Bearer" followed by a space and JWT token.

func main() {
	cfg := config.LoadConfig("config.toml")
	fmt.Printf("AppName: %s, Env: %s\n", cfg.AppName, cfg.Env)

	// 初始化Kafka
	if err := kafka.InitKafka(cfg); err != nil {
		log.Fatalf("Failed to initialize Kafka: %v", err)
	}

	// 创建 Gin 引擎
	r := gin.New()
	r.Use(gin.Logger(), gin.Recovery(), middleware.RateLimit())

	// 设置Swagger
	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	// 设置路由
	router.SetupRouter(r)

	// 启动 HTTP 服务，使用配置中的端口（若未设置则回退到 :8080）
	addr := ":8080"

	log.Printf("Server starting on %s", addr)
	if err := r.Run(addr); err != nil {
		log.Fatalf("failed to run server: %v", err)
	}
}
