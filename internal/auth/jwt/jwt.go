package jwt

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"errors"
	"fmt"
	"time"

	"apihub/internal/auth/cache"
	"apihub/internal/model"

	"github.com/golang-jwt/jwt/v5"
)

// JWTService JWT服务
type JWTService struct {
	privateKey   *rsa.PrivateKey
	publicKey    *rsa.PublicKey
	accessExpiry time.Duration
	issuer       string
	cacheService cache.CacheService
}

// JWTConfig JWT配置
type JWTConfig struct {
	PrivateKeyPEM string        // RSA私钥PEM格式
	PublicKeyPEM  string        // RSA公钥PEM格式
	AccessExpiry  time.Duration // 访问令牌过期时间
	Issuer        string        // 签发者
}

// TokenResponse Token响应
type TokenResponse struct {
	AccessToken string `json:"access_token"`
	ExpiresIn   int64  `json:"expires_in"` // 访问令牌过期时间(秒)
}

// NewJWTService 创建JWT服务实例
func NewJWTService(config JWTConfig, cacheService cache.CacheService) (*JWTService, error) {
	service := &JWTService{
		accessExpiry: config.AccessExpiry,
		issuer:       config.Issuer,
		cacheService: cacheService,
	}

	// 解析私钥
	if config.PrivateKeyPEM != "" {
		privateKey, err := parsePrivateKey(config.PrivateKeyPEM)
		if err != nil {
			return nil, fmt.Errorf("failed to parse private key: %w", err)
		}
		service.privateKey = privateKey
		service.publicKey = &privateKey.PublicKey
	} else {
		// 如果没有提供密钥，生成新的密钥对
		privateKey, err := generateRSAKeyPair()
		if err != nil {
			return nil, fmt.Errorf("failed to generate RSA key pair: %w", err)
		}
		service.privateKey = privateKey
		service.publicKey = &privateKey.PublicKey
	}

	return service, nil
}

// GenerateToken 生成访问令牌
func (s *JWTService) GenerateToken(user *model.User) (*TokenResponse, error) {
	now := time.Now()

	// 生成访问令牌
	claims := CustomClaims{
		UserID:   int(user.ID),
		Username: user.Username,
		Role:     user.Role,
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    s.issuer,
			Subject:   fmt.Sprintf("%d", user.ID),
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(s.accessExpiry)),
			NotBefore: jwt.NewNumericDate(now),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodRS256, claims)
	tokenString, err := token.SignedString(s.privateKey)
	if err != nil {
		return nil, fmt.Errorf("failed to sign access token: %w", err)
	}

	return &TokenResponse{
		AccessToken: tokenString,
		ExpiresIn:   int64(s.accessExpiry.Seconds()),
	}, nil
}

// ValidateToken 验证Token
func (s *JWTService) ValidateToken(tokenString string) (*CustomClaims, error) {
	// 检查Token是否在黑名单中
	if s.cacheService.IsBlacklisted(tokenString) {
		return nil, errors.New("token is blacklisted")
	}

	// 解析Token
	token, err := jwt.ParseWithClaims(tokenString, &CustomClaims{}, func(token *jwt.Token) (interface{}, error) {
		// 验证签名方法
		if _, ok := token.Method.(*jwt.SigningMethodRSA); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return s.publicKey, nil
	})

	if err != nil {
		return nil, fmt.Errorf("failed to parse token: %w", err)
	}

	// 验证Token有效性
	if !token.Valid {
		return nil, errors.New("token is invalid")
	}

	claims, ok := token.Claims.(*CustomClaims)
	if !ok {
		return nil, errors.New("invalid token claims")
	}

	return claims, nil
}

// RevokeToken 撤销Token
func (s *JWTService) RevokeToken(tokenString string) error {
	// 验证Token以获取过期时间
	claims, err := s.ValidateToken(tokenString)
	if err != nil {
		// 即使Token无效，也尝试加入黑名单
		return s.cacheService.AddToBlacklist(tokenString, s.accessExpiry)
	}

	// 计算剩余有效时间
	remainingTime := time.Until(claims.ExpiresAt.Time)
	if remainingTime <= 0 {
		// Token已过期，无需加入黑名单
		return nil
	}

	// 加入黑名单
	err = s.cacheService.AddToBlacklist(tokenString, remainingTime)
	if err != nil {
		return fmt.Errorf("failed to add token to blacklist: %w", err)
	}

	return nil
}

// GetPublicKeyPEM 获取公钥PEM格式
func (s *JWTService) GetPublicKeyPEM() (string, error) {
	publicKeyBytes, err := x509.MarshalPKIXPublicKey(s.publicKey)
	if err != nil {
		return "", fmt.Errorf("failed to marshal public key: %w", err)
	}

	publicKeyPEM := pem.EncodeToMemory(&pem.Block{
		Type:  "PUBLIC KEY",
		Bytes: publicKeyBytes,
	})

	return string(publicKeyPEM), nil
}

// parsePrivateKey 解析私钥PEM
func parsePrivateKey(privateKeyPEM string) (*rsa.PrivateKey, error) {
	block, _ := pem.Decode([]byte(privateKeyPEM))
	if block == nil {
		return nil, errors.New("failed to decode PEM block")
	}

	privateKey, err := x509.ParsePKCS1PrivateKey(block.Bytes)
	if err != nil {
		// 尝试PKCS8格式
		key, err := x509.ParsePKCS8PrivateKey(block.Bytes)
		if err != nil {
			return nil, fmt.Errorf("failed to parse private key: %w", err)
		}

		rsaKey, ok := key.(*rsa.PrivateKey)
		if !ok {
			return nil, errors.New("key is not RSA private key")
		}
		return rsaKey, nil
	}

	return privateKey, nil
}

// generateRSAKeyPair 生成RSA密钥对
func generateRSAKeyPair() (*rsa.PrivateKey, error) {
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return nil, fmt.Errorf("failed to generate RSA key: %w", err)
	}
	return privateKey, nil
}
