package cache

import (
	"time"

	"github.com/patrickmn/go-cache"
)

// CacheService 缓存服务接口
type CacheService interface {
	// JWT Token相关缓存操作
	SetToken(token string, value interface{}, expiration time.Duration) error
	GetToken(token string) (interface{}, bool)
	DeleteToken(token string) error

	// 黑名单Token操作
	AddToBlacklist(token string, expiration time.Duration) error
	IsBlacklisted(token string) bool

	// 通用缓存操作
	Set(key string, value interface{}, expiration time.Duration) error
	Get(key string) (interface{}, bool)
	Delete(key string) error
	Clear() error
}

// GoCacheService go-cache实现的缓存服务
type GoCacheService struct {
	cache *cache.Cache
}

// NewGoCacheService 创建新的go-cache服务实例
func NewGoCacheService(defaultExpiration, cleanupInterval time.Duration) *GoCacheService {
	return &GoCacheService{
		cache: cache.New(defaultExpiration, cleanupInterval),
	}
}

// SetToken 设置Token缓存
func (s *GoCacheService) SetToken(token string, value interface{}, expiration time.Duration) error {
	s.cache.Set("token:"+token, value, expiration)
	return nil
}

// GetToken 获取Token缓存
func (s *GoCacheService) GetToken(token string) (interface{}, bool) {
	return s.cache.Get("token:" + token)
}

// DeleteToken 删除Token缓存
func (s *GoCacheService) DeleteToken(token string) error {
	s.cache.Delete("token:" + token)
	return nil
}

// AddToBlacklist 添加Token到黑名单
func (s *GoCacheService) AddToBlacklist(token string, expiration time.Duration) error {
	s.cache.Set("blacklist:"+token, true, expiration)
	return nil
}

// IsBlacklisted 检查Token是否在黑名单中
func (s *GoCacheService) IsBlacklisted(token string) bool {
	_, found := s.cache.Get("blacklist:" + token)
	return found
}

// Set 通用设置缓存
func (s *GoCacheService) Set(key string, value interface{}, expiration time.Duration) error {
	s.cache.Set(key, value, expiration)
	return nil
}

// Get 通用获取缓存
func (s *GoCacheService) Get(key string) (interface{}, bool) {
	return s.cache.Get(key)
}

// Delete 通用删除缓存
func (s *GoCacheService) Delete(key string) error {
	s.cache.Delete(key)
	return nil
}

// Clear 清空所有缓存
func (s *GoCacheService) Clear() error {
	s.cache.Flush()
	return nil
}
