package utils

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// Response 标准API响应结构
type Response struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Data    any    `json:"data,omitempty"`
}

// ResponseSuccess 成功响应
func ResponseSuccess(c *gin.Context, data any) {
	c.JSON(http.StatusOK, Response{
		Code:    0,
		Message: "success",
		Data:    data,
	})
}

// ResponseError 错误响应
func ResponseError(c *gin.Context, httpCode int, errCode int, message string) {
	c.JSON(httpCode, Response{
		Code:    errCode,
		Message: message,
	})
}

// ResponseBadRequest 400错误响应
func ResponseBadRequest(c *gin.Context, message string) {
	ResponseError(c, http.StatusBadRequest, 400, message)
}

// ResponseUnauthorized 401错误响应
func ResponseUnauthorized(c *gin.Context, message string) {
	if message == "" {
		message = "未授权访问"
	}
	ResponseError(c, http.StatusUnauthorized, 401, message)
}

// ResponseForbidden 403错误响应
func ResponseForbidden(c *gin.Context, message string) {
	if message == "" {
		message = "禁止访问"
	}
	ResponseError(c, http.StatusForbidden, 403, message)
}

// ResponseNotFound 404错误响应
func ResponseNotFound(c *gin.Context, message string) {
	if message == "" {
		message = "资源不存在"
	}
	ResponseError(c, http.StatusNotFound, 404, message)
}

// ResponseInternalError 500错误响应
func ResponseInternalError(c *gin.Context, message string) {
	if message == "" {
		message = "服务器内部错误"
	}
	ResponseError(c, http.StatusInternalServerError, 500, message)
}
