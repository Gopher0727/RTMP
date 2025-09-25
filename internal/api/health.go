package api

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// HealthHandler 返回服务健康状态，供监控/负载均衡器使用。
func HealthHandler(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status": "ok",
	})
}
