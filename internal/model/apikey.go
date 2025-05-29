package model

import (
	"time"
)

// APIKey API密钥模型
type APIKey struct {
	ID        int        `json:"id" db:"id"`
	UserID    int        `json:"user_id" db:"user_id"`
	KeyName   string     `json:"key_name" db:"key_name"`
	APIKey    string     `json:"api_key" db:"api_key"`
	Status    int        `json:"status" db:"status"` // 0: disabled, 1: active
	CreatedAt time.Time  `json:"created_at" db:"created_at"`
	ExpiresAt *time.Time `json:"expires_at" db:"expires_at"`
}

// APIKeyStatus API密钥状态常量
const (
	APIKeyStatusDisabled = 0
	APIKeyStatusActive   = 1
)

// CreateAPIKeyRequest 创建API密钥请求
type CreateAPIKeyRequest struct {
	KeyName   string     `json:"key_name" binding:"required,min=1,max=100"`
	ExpiresAt *time.Time `json:"expires_at"`
}

// UpdateAPIKeyRequest 更新API密钥请求
type UpdateAPIKeyRequest struct {
	KeyName   string     `json:"key_name" binding:"omitempty,min=1,max=100"`
	Status    int        `json:"status" binding:"omitempty,oneof=0 1"`
	ExpiresAt *time.Time `json:"expires_at"`
}

// APIKeyResponse API密钥响应（隐藏完整密钥）
type APIKeyResponse struct {
	ID        int        `json:"id"`
	KeyName   string     `json:"key_name"`
	KeyPrefix string     `json:"key_prefix"` // 只显示前几位
	Status    int        `json:"status"`
	CreatedAt time.Time  `json:"created_at"`
	ExpiresAt *time.Time `json:"expires_at"`
}

// IsActive 检查API密钥是否激活
func (ak *APIKey) IsActive() bool {
	if ak.Status != APIKeyStatusActive {
		return false
	}

	// 检查是否过期
	if ak.ExpiresAt != nil && ak.ExpiresAt.Before(time.Now()) {
		return false
	}

	return true
}

// IsExpired 检查API密钥是否过期
func (ak *APIKey) IsExpired() bool {
	if ak.ExpiresAt == nil {
		return false
	}
	return ak.ExpiresAt.Before(time.Now())
}

// ToResponse 转换为响应格式
func (ak *APIKey) ToResponse() *APIKeyResponse {
	keyPrefix := ""
	if len(ak.APIKey) > 8 {
		keyPrefix = ak.APIKey[:8] + "..."
	} else {
		keyPrefix = ak.APIKey
	}

	return &APIKeyResponse{
		ID:        ak.ID,
		KeyName:   ak.KeyName,
		KeyPrefix: keyPrefix,
		Status:    ak.Status,
		CreatedAt: ak.CreatedAt,
		ExpiresAt: ak.ExpiresAt,
	}
}
