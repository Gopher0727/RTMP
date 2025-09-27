package api

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// HealthHandler godoc
// @Summary 健康检查接口
// @Description 返回服务健康状态，供监控/负载均衡器使用
// @Tags health
// @Accept json
// @Produce json
// @Success 200 {object} map[string]string
// @Router /health [get]
func HealthHandler(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status": "ok",
	})
}
