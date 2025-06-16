package services

import (
	"fmt"
	"time"

	"github.com/gin-gonic/gin"
)

// EchoRequest Echo服务请求
type EchoRequest struct {
	Message string `json:"message" binding:"required"`
}

// EchoResponse Echo服务响应
type EchoResponse struct {
	Message   string `json:"message"`
	Timestamp int64  `json:"timestamp"`
}

// EchoServiceHandler Echo服务处理函数
func EchoServiceHandler(c *gin.Context) (interface{}, error) {
	var request EchoRequest

	if err := c.ShouldBindJSON(&request); err != nil {
		return nil, fmt.Errorf("无效的请求参数: %w", err)
	}

	return &EchoResponse{
		Message:   request.Message,
		Timestamp: time.Now().Unix(),
	}, nil
}
