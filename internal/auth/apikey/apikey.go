package apikey

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"time"

	"apihub/internal/auth/crypto"
	"apihub/internal/model"
	"apihub/internal/store"
)

// APIKeyService APIKey服务
type APIKeyService struct {
	store         store.Store
	cryptoService crypto.CryptoService
}

// NewAPIKeyService 创建APIKey服务实例
func NewAPIKeyService(store store.Store, cryptoService crypto.CryptoService) *APIKeyService {
	return &APIKeyService{
		store:         store,
		cryptoService: cryptoService,
	}
}

// GenerateAPIKey 生成新的APIKey
func (s *APIKeyService) GenerateAPIKey(length int) (string, error) {
	if length <= 0 {
		length = 32 // 默认32字符
	}

	// 生成随机字节
	bytes := make([]byte, length/2)
	if _, err := rand.Read(bytes); err != nil {
		return "", fmt.Errorf("failed to generate random bytes: %w", err)
	}

	// 转换为十六进制字符串
	return hex.EncodeToString(bytes), nil
}

// CreateAPIKey 创建APIKey记录
func (s *APIKeyService) CreateAPIKey(userID int, name, description string, expiresAt *time.Time, scopes []string) (*model.APIKey, error) {
	// 生成APIKey
	keyString, err := s.GenerateAPIKey(32)
	if err != nil {
		return nil, fmt.Errorf("生成API密钥失败: %w", err)
	}

	// 加密APIKey
	encryptedKey, err := s.cryptoService.Encrypt(keyString)
	if err != nil {
		return nil, fmt.Errorf("加密API密钥失败: %w", err)
	}

	// 创建APIKey模型
	apiKey := &model.APIKey{
		UserID:    userID,
		KeyName:   name,
		APIKey:    encryptedKey, // 存储加密后的APIKey
		Status:    model.APIKeyStatusActive,
		ExpiresAt: expiresAt,
		CreatedAt: time.Now(),
	}

	// 保存到数据库
	err = s.store.APIKeys().Create(context.Background(), apiKey)
	if err != nil {
		return nil, fmt.Errorf("创建API密钥失败: %w", err)
	}

	// 返回时包含明文APIKey（仅此一次）
	apiKey.APIKey = keyString
	return apiKey, nil
}

// ValidateAPIKey 验证APIKey
func (s *APIKeyService) ValidateAPIKey(keyString string) (*model.APIKey, error) {
	if keyString == "" {
		return nil, errors.New("API密钥不能为空")
	}

	fmt.Printf("验证API密钥: %s...\n", keyString[:4])

	// 加密输入的API密钥
	encryptedKey, err := s.cryptoService.Encrypt(keyString)
	if err != nil {
		return nil, fmt.Errorf("加密API密钥失败: %w", err)
	}

	// 直接使用加密后的密钥查询数据库
	apiKey, err := s.store.APIKeys().GetByKey(context.Background(), encryptedKey)
	if err != nil {
		return nil, fmt.Errorf("API密钥验证失败: %w", err)
	}

	// 检查APIKey状态
	if apiKey.Status != model.APIKeyStatusActive {
		return nil, errors.New("API密钥未激活")
	}

	// 检查过期时间
	if apiKey.ExpiresAt != nil && time.Now().After(*apiKey.ExpiresAt) {
		return nil, errors.New("API密钥已过期")
	}

	// 返回APIKey（包含明文密钥）
	apiKey.APIKey = keyString
	return apiKey, nil
}

// GetAPIKeysByUserID 获取用户的所有APIKey
func (s *APIKeyService) GetAPIKeysByUserID(userID int) ([]*model.APIKey, error) {
	apiKeys, err := s.store.APIKeys().GetByUserID(context.Background(), userID)
	if err != nil {
		return nil, fmt.Errorf("获取API密钥失败: %w", err)
	}

	// 解密所有APIKey
	for _, apiKey := range apiKeys {
		decryptedKey, err := s.cryptoService.Decrypt(apiKey.APIKey)
		if err != nil {
			// 如果解密失败，设置为空字符串而不是返回错误
			apiKey.APIKey = ""
			continue
		}
		apiKey.APIKey = decryptedKey
	}

	return apiKeys, nil
}

// UpdateAPIKey 更新APIKey
func (s *APIKeyService) UpdateAPIKey(apiKeyID int, name string, status int, expiresAt *time.Time) error {
	// 获取现有APIKey
	apiKey, err := s.store.APIKeys().GetByID(context.Background(), apiKeyID)
	if err != nil {
		return fmt.Errorf("failed to get API key: %w", err)
	}

	// 更新字段
	if name != "" {
		apiKey.KeyName = name
	}
	if status != 0 {
		apiKey.Status = status
	}
	if expiresAt != nil {
		apiKey.ExpiresAt = expiresAt
	}

	// 保存更新
	err = s.store.APIKeys().Update(context.Background(), apiKey)
	if err != nil {
		return fmt.Errorf("failed to update API key: %w", err)
	}

	return nil
}

// RevokeAPIKey 撤销APIKey
func (s *APIKeyService) RevokeAPIKey(apiKeyID int) error {
	return s.UpdateAPIKey(apiKeyID, "", model.APIKeyStatusDisabled, nil)
}

// DeleteAPIKey 删除APIKey
func (s *APIKeyService) DeleteAPIKey(apiKeyID int) error {
	err := s.store.APIKeys().Delete(context.Background(), apiKeyID)
	if err != nil {
		return fmt.Errorf("failed to delete API key: %w", err)
	}
	return nil
}

// RegenerateAPIKey 重新生成APIKey
func (s *APIKeyService) RegenerateAPIKey(apiKeyID int) (*model.APIKey, error) {
	// 获取现有APIKey
	apiKey, err := s.store.APIKeys().GetByID(context.Background(), apiKeyID)
	if err != nil {
		return nil, fmt.Errorf("获取API密钥失败: %w", err)
	}

	// 生成新的APIKey
	newKeyString, err := s.GenerateAPIKey(32)
	if err != nil {
		return nil, fmt.Errorf("生成新API密钥失败: %w", err)
	}

	// 加密新的APIKey
	encryptedKey, err := s.cryptoService.Encrypt(newKeyString)
	if err != nil {
		return nil, fmt.Errorf("加密新API密钥失败: %w", err)
	}

	// 更新APIKey
	apiKey.APIKey = encryptedKey

	err = s.store.APIKeys().Update(context.Background(), apiKey)
	if err != nil {
		return nil, fmt.Errorf("更新API密钥失败: %w", err)
	}

	// 返回时包含明文APIKey
	apiKey.APIKey = newKeyString
	return apiKey, nil
}

// CheckAPIKeyScope 检查APIKey是否具有指定的权限范围
// 注意：当前APIKey模型不包含Scopes字段，默认允许所有操作
func (s *APIKeyService) CheckAPIKeyScope(apiKey *model.APIKey, requiredScope string) bool {
	if apiKey == nil {
		return false
	}

	// 当前实现：如果APIKey有效，则允许所有操作
	// 未来可以扩展APIKey模型添加权限范围字段
	return apiKey.IsActive()
}
