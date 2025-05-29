package model

import (
	"time"
)

// ServiceQuota 服务配额模型
type ServiceQuota struct {
	ID          int       `json:"id" db:"id"`
	UserID      int       `json:"user_id" db:"user_id"`
	ServiceName string    `json:"service_name" db:"service_name"`
	TimeWindow  string    `json:"time_window" db:"time_window"` // 时间窗口：2024-03 或 2024-03-15
	Usage       int       `json:"usage" db:"usage"`             // 当前使用量
	LimitValue  int       `json:"limit_value" db:"limit_value"` // -1表示无限制
	ResetTime   time.Time `json:"reset_time" db:"reset_time"`   // 下次重置时间
	CreatedAt   time.Time `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time `json:"updated_at" db:"updated_at"`
}

// AccessLog 访问日志模型
type AccessLog struct {
	ID          int       `json:"id" db:"id"`
	APIKeyID    int       `json:"api_key_id" db:"api_key_id"`
	UserID      int       `json:"user_id" db:"user_id"`
	ServiceName string    `json:"service_name" db:"service_name"`
	Endpoint    string    `json:"endpoint" db:"endpoint"`
	Status      int       `json:"status" db:"status"`
	Cost        int       `json:"cost" db:"cost"` // API调用计费单位
	CreatedAt   time.Time `json:"created_at" db:"created_at"`
}

// QuotaRequest 配额设置请求
type QuotaRequest struct {
	UserID      int    `json:"user_id" binding:"required"`
	ServiceName string `json:"service_name" binding:"required"`
	LimitValue  int    `json:"limit_value" binding:"min=-1"` // -1表示无限制
}

// QuotaResponse 配额响应
type QuotaResponse struct {
	UserID      int       `json:"user_id"`
	ServiceName string    `json:"service_name"`
	TimeWindow  string    `json:"time_window"`
	Usage       int       `json:"usage"`
	LimitValue  int       `json:"limit_value"`
	ResetTime   time.Time `json:"reset_time"`
	IsExceeded  bool      `json:"is_exceeded"`
}

// UsageStatsRequest 使用统计请求
type UsageStatsRequest struct {
	UserID      int    `json:"user_id"`
	ServiceName string `json:"service_name"`
	StartDate   string `json:"start_date"` // YYYY-MM-DD
	EndDate     string `json:"end_date"`   // YYYY-MM-DD
}

// UsageStatsResponse 使用统计响应
type UsageStatsResponse struct {
	UserID      int                `json:"user_id"`
	ServiceName string             `json:"service_name"`
	TotalUsage  int                `json:"total_usage"`
	DailyUsage  map[string]int     `json:"daily_usage"` // 日期 -> 使用量
	Details     []AccessLogSummary `json:"details"`
}

// AccessLogSummary 访问日志摘要
type AccessLogSummary struct {
	Date         string `json:"date"`
	TotalCalls   int    `json:"total_calls"`
	SuccessCalls int    `json:"success_calls"`
	ErrorCalls   int    `json:"error_calls"`
	TotalCost    int    `json:"total_cost"`
}

// IsExceeded 检查配额是否超限
func (sq *ServiceQuota) IsExceeded() bool {
	if sq.LimitValue == -1 {
		return false // 无限制
	}
	return sq.Usage >= sq.LimitValue
}

// CanUse 检查是否可以使用（考虑成本）
func (sq *ServiceQuota) CanUse(cost int) bool {
	if sq.LimitValue == -1 {
		return true // 无限制
	}
	return sq.Usage+cost <= sq.LimitValue
}

// ToResponse 转换为响应格式
func (sq *ServiceQuota) ToResponse() *QuotaResponse {
	return &QuotaResponse{
		UserID:      sq.UserID,
		ServiceName: sq.ServiceName,
		TimeWindow:  sq.TimeWindow,
		Usage:       sq.Usage,
		LimitValue:  sq.LimitValue,
		ResetTime:   sq.ResetTime,
		IsExceeded:  sq.IsExceeded(),
	}
}
