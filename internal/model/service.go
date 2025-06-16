package model

import (
	"time"
)

// ServiceDefinition 服务定义模型
type ServiceDefinition struct {
	ID           int       `json:"id" db:"id"`
	ServiceName  string    `json:"service_name" db:"service_name"`
	Description  string    `json:"description" db:"description"`
	DefaultLimit int       `json:"default_limit" db:"default_limit"` // 默认限制值，-1表示无限制
	Status       int       `json:"status" db:"status"`               // 1-启用，0-禁用
	CreatedAt    time.Time `json:"created_at" db:"created_at"`
	UpdatedAt    time.Time `json:"updated_at" db:"updated_at"`
	// 新增字段
	AllowAnonymous bool `json:"allow_anonymous" db:"allow_anonymous"` // 是否允许匿名访问
	RateLimit      int  `json:"rate_limit" db:"rate_limit"`           // 限流值（每分钟请求数）
	QuotaCost      int  `json:"quota_cost" db:"quota_cost"`           // 每次调用消耗的配额
}

// ServiceStatus 服务状态常量
const (
	ServiceStatusDisabled = 0
	ServiceStatusEnabled  = 1
)

// CreateServiceRequest 创建服务请求
type CreateServiceRequest struct {
	ServiceName    string `json:"service_name" binding:"required,min=1,max=100"`
	Description    string `json:"description" binding:"required,min=1,max=500"`
	DefaultLimit   int    `json:"default_limit" binding:"min=-1"`
	AllowAnonymous bool   `json:"allow_anonymous"`
	RateLimit      int    `json:"rate_limit" binding:"min=0"`
	QuotaCost      int    `json:"quota_cost" binding:"min=0"`
}

// UpdateServiceRequest 更新服务请求
type UpdateServiceRequest struct {
	Description    string `json:"description" binding:"omitempty,min=1,max=500"`
	DefaultLimit   int    `json:"default_limit" binding:"omitempty,min=-1"`
	Status         int    `json:"status" binding:"omitempty,oneof=0 1"`
	AllowAnonymous bool   `json:"allow_anonymous"`
	RateLimit      int    `json:"rate_limit" binding:"omitempty,min=0"`
	QuotaCost      int    `json:"quota_cost" binding:"omitempty,min=0"`
}

// ServiceResponse 服务响应
type ServiceResponse struct {
	ID             int       `json:"id"`
	ServiceName    string    `json:"service_name"`
	Description    string    `json:"description"`
	DefaultLimit   int       `json:"default_limit"`
	Status         int       `json:"status"`
	AllowAnonymous bool      `json:"allow_anonymous"`
	RateLimit      int       `json:"rate_limit"`
	QuotaCost      int       `json:"quota_cost"`
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
}

// IsEnabled 检查服务是否启用
func (sd *ServiceDefinition) IsEnabled() bool {
	return sd.Status == ServiceStatusEnabled
}

// HasLimit 检查服务是否有限制
func (sd *ServiceDefinition) HasLimit() bool {
	return sd.DefaultLimit != -1
}

// ToResponse 转换为响应格式
func (sd *ServiceDefinition) ToResponse() *ServiceResponse {
	return &ServiceResponse{
		ID:             sd.ID,
		ServiceName:    sd.ServiceName,
		Description:    sd.Description,
		DefaultLimit:   sd.DefaultLimit,
		Status:         sd.Status,
		AllowAnonymous: sd.AllowAnonymous,
		RateLimit:      sd.RateLimit,
		QuotaCost:      sd.QuotaCost,
		CreatedAt:      sd.CreatedAt,
		UpdatedAt:      sd.UpdatedAt,
	}
}

// ServiceConfig 服务配置（内存中使用，不存储到数据库）
type ServiceConfig struct {
	// 是否允许匿名访问
	AllowAnonymous bool `json:"allow_anonymous"`
	// 默认限流配置（每分钟请求数）
	RateLimit int `json:"rate_limit"`
	// 默认消耗配额
	QuotaCost int `json:"quota_cost"`
	// 服务描述信息
	Description string `json:"description"`
	// 请求示例
	RequestExample interface{} `json:"request_example,omitempty"`
	// 响应示例
	ResponseExample interface{} `json:"response_example,omitempty"`
}
