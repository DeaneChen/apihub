package model

import (
	"time"
)

// SystemConfig 系统配置模型
type SystemConfig struct {
	ID          int       `json:"id" db:"id"`
	ConfigKey   string    `json:"config_key" db:"config_key"`
	ConfigValue string    `json:"config_value" db:"config_value"`
	UpdatedAt   time.Time `json:"updated_at" db:"updated_at"`
}

// 系统配置键常量
const (
	ConfigKeySystemInitialized = "system_initialized"  // 系统是否已初始化
	ConfigKeyDefaultQuotaLimit = "default_quota_limit" // 默认配额限制
	ConfigKeyJWTSecret         = "jwt_secret"          // JWT密钥
	ConfigKeySystemTitle       = "system_title"        // 系统标题
	ConfigKeySystemDescription = "system_description"  // 系统描述
	ConfigKeyRegistrationOpen  = "registration_open"   // 是否开放注册
)

// ConfigRequest 配置请求
type ConfigRequest struct {
	ConfigKey   string `json:"config_key" binding:"required"`
	ConfigValue string `json:"config_value" binding:"required"`
}

// ConfigResponse 配置响应
type ConfigResponse struct {
	ConfigKey   string    `json:"config_key"`
	ConfigValue string    `json:"config_value"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// BatchConfigRequest 批量配置请求
type BatchConfigRequest struct {
	Configs []ConfigRequest `json:"configs" binding:"required,dive"`
}

// ToResponse 转换为响应格式
func (sc *SystemConfig) ToResponse() *ConfigResponse {
	return &ConfigResponse{
		ConfigKey:   sc.ConfigKey,
		ConfigValue: sc.ConfigValue,
		UpdatedAt:   sc.UpdatedAt,
	}
}
