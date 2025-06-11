package auth

import (
	"time"

	"apihub/internal/auth/apikey"
	"apihub/internal/auth/cache"
	"apihub/internal/auth/crypto"
	"apihub/internal/auth/jwt"
	"apihub/internal/auth/permission"
	"apihub/internal/store"
)

// AuthConfig 认证配置
type AuthConfig struct {
	// JWT配置
	JWT JWTConfig `json:"jwt"`

	// 加密配置
	Crypto CryptoConfig `json:"crypto"`

	// 缓存配置
	Cache CacheConfig `json:"cache"`
}

// JWTConfig JWT配置
type JWTConfig struct {
	PrivateKeyPEM string        `json:"private_key_pem"` // RSA私钥PEM格式
	PublicKeyPEM  string        `json:"public_key_pem"`  // RSA公钥PEM格式
	AccessExpiry  time.Duration `json:"access_expiry"`   // 访问令牌过期时间
	Issuer        string        `json:"issuer"`          // 签发者
}

// CryptoConfig 加密配置
type CryptoConfig struct {
	SecretKey string `json:"secret_key"` // 加密密钥
}

// CacheConfig 缓存配置
type CacheConfig struct {
	DefaultExpiration time.Duration `json:"default_expiration"` // 默认过期时间
	CleanupInterval   time.Duration `json:"cleanup_interval"`   // 清理间隔
}

// AuthServices 认证服务集合
type AuthServices struct {
	JWTService        *jwt.JWTService
	APIKeyService     *apikey.APIKeyService
	CryptoService     crypto.CryptoService
	CacheService      cache.CacheService
	PermissionService *permission.PermissionService
}

// NewAuthServices 创建认证服务集合
func NewAuthServices(config AuthConfig, store store.Store) (*AuthServices, error) {
	// 创建缓存服务
	cacheService := cache.NewGoCacheService(
		config.Cache.DefaultExpiration,
		config.Cache.CleanupInterval,
	)

	// 创建JWT服务
	jwtConfig := jwt.JWTConfig{
		PrivateKeyPEM: config.JWT.PrivateKeyPEM,
		PublicKeyPEM:  config.JWT.PublicKeyPEM,
		AccessExpiry:  config.JWT.AccessExpiry,
		Issuer:        config.JWT.Issuer,
	}
	jwtService, err := jwt.NewJWTService(jwtConfig, cacheService)
	if err != nil {
		return nil, err
	}

	// 创建加密服务
	cryptoService := crypto.NewAESCryptoService(config.Crypto.SecretKey)

	// 创建APIKey服务
	apiKeyService := apikey.NewAPIKeyService(store, cryptoService)

	// 创建权限服务
	permissionService := permission.NewPermissionService()

	return &AuthServices{
		JWTService:        jwtService,
		APIKeyService:     apiKeyService,
		CryptoService:     cryptoService,
		CacheService:      cacheService,
		PermissionService: permissionService,
	}, nil
}

// DefaultAuthConfig 默认认证配置
func DefaultAuthConfig() AuthConfig {
	return AuthConfig{
		JWT: JWTConfig{
			AccessExpiry: 24 * time.Hour, // 访问令牌24小时过期
			Issuer:       "apihub",
		},
		Crypto: CryptoConfig{
			SecretKey: "default-secret-key-change-in-production", // 生产环境需要更改
		},
		Cache: CacheConfig{
			DefaultExpiration: 30 * time.Minute, // 默认缓存30分钟
			CleanupInterval:   10 * time.Minute, // 每10分钟清理一次过期缓存
		},
	}
}
