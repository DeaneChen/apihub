package services

import (
	"time"

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
