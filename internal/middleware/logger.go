package middleware

import (
	"fmt"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

// Logger 日志中间件
// @Summary 请求日志中间件
// @Description 记录HTTP请求的详细信息
// @Tags middleware
func Logger() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 开始时间
		startTime := time.Now()

		// 处理请求
		c.Next()

		// 结束时间
		endTime := time.Now()

		// 执行时间
		latencyTime := endTime.Sub(startTime)

		// 请求方式
		reqMethod := c.Request.Method

		// 请求路由
		reqUri := c.Request.RequestURI

		// 状态码
		statusCode := c.Writer.Status()

		// 请求IP
		clientIP := c.ClientIP()

		// 日志格式
		logrus.WithFields(logrus.Fields{
			"status_code":  statusCode,
			"latency_time": latencyTime,
			"client_ip":    clientIP,
			"req_method":   reqMethod,
			"req_uri":      reqUri,
		}).Info("HTTP请求")

		// 如果是错误请求，额外记录错误信息
		if len(c.Errors) > 0 {
			for _, e := range c.Errors.Errors() {
				logrus.Error(e)
			}
		}

		// 控制台输出
		fmt.Printf("[GIN] %v | %3d | %13v | %15s | %-7s %s\n",
			endTime.Format("2006/01/02 - 15:04:05"),
			statusCode,
			latencyTime,
			clientIP,
			reqMethod,
			reqUri,
		)
	}
}
