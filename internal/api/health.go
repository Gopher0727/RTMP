package api

import (
	"github.com/gin-gonic/gin"

	"github.com/Gopher0727/RTMP/internal/utils"
)

// HealthHandler godoc
// @Summary 健康检查接口
// @Description 返回服务健康状态，供监控/负载均衡器使用
// @Tags health
// @Accept json
// @Produce json
// @Success 200 {object} map[string]string
// @Router /health [get]
// HealthHandler 健康检查处理器
func HealthHandler(c *gin.Context) {
	utils.ResponseSuccess(c, gin.H{
		"status":  "ok",
		"service": "RTMP Service",
	})
}
