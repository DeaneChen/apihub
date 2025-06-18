package services

import (
	"fmt"
	"time"

	"apihub/internal/model"

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

// EchoServiceConfig 获取Echo服务配置
func EchoServiceConfig() model.ServiceConfig {
	return model.ServiceConfig{
		AllowAnonymous:  true,
		RateLimit:       60, // 每分钟60次
		QuotaCost:       1,  // 消耗1个配额
		Description:     "回显服务，返回请求的内容",
		RequestExample:  map[string]interface{}{"message": "Hello, APIHub!"},
		ResponseExample: map[string]interface{}{"message": "Hello, APIHub!", "timestamp": 1625097600},
	}
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
