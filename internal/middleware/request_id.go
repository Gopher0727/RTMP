package middleware

import (
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

const (
	// RequestIDKey 请求ID在上下文中的键名
	RequestIDKey = "RequestID"
	// RequestIDHeader 请求ID在HTTP头中的键名
	RequestIDHeader = "X-Request-ID"
)

// RequestID 请求ID中间件
// @Summary 请求ID中间件
// @Description 为每个请求生成唯一标识符
// @Tags middleware
func RequestID() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 检查请求头中是否已有请求ID
		requestID := c.Request.Header.Get(RequestIDHeader)
		if requestID == "" {
			// 生成新的请求ID
			requestID = uuid.New().String()
		}

		// 将请求ID设置到上下文中
		c.Set(RequestIDKey, requestID)

		// 将请求ID添加到响应头
		c.Writer.Header().Set(RequestIDHeader, requestID)

		c.Next()
	}
}

// GetRequestID 从上下文中获取请求ID
func GetRequestID(c *gin.Context) string {
	if id, exists := c.Get(RequestIDKey); exists {
		return id.(string)
	}
	return ""
}
