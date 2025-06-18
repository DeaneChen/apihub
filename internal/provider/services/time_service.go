package services

import (
	"time"

	"apihub/internal/model"

	"github.com/gin-gonic/gin"
)

// TimeResponse 时间服务响应
type TimeResponse struct {
	Timestamp int64  `json:"timestamp"`
	ISO8601   string `json:"iso8601"`
	Date      string `json:"date"`
	Time      string `json:"time"`
	Timezone  string `json:"timezone"`
}

// TimeServiceConfig 获取时间服务配置
func TimeServiceConfig() model.ServiceConfig {
	return model.ServiceConfig{
		AllowAnonymous: true,
		RateLimit:      60, // 每分钟60次
		QuotaCost:      1,  // 消耗1个配额
		Description:    "时间服务，返回当前服务器时间",
		RequestExample: map[string]interface{}{},
		ResponseExample: map[string]interface{}{
			"timestamp": 1625097600,
			"iso8601":   "2021-07-01T00:00:00Z",
			"date":      "2021-07-01",
			"time":      "00:00:00",
			"timezone":  "UTC",
		},
	}
}

// TimeServiceHandler 时间服务处理函数
func TimeServiceHandler(c *gin.Context) (interface{}, error) {
	now := time.Now()

	return &TimeResponse{
		Timestamp: now.Unix(),
		ISO8601:   now.Format(time.RFC3339),
		Date:      now.Format("2006-01-02"),
		Time:      now.Format("15:04:05"),
		Timezone:  now.Location().String(),
	}, nil
}
