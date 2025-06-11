package jwt

import (
	"github.com/golang-jwt/jwt/v5"
)

// CustomClaims 自定义JWT Claims
type CustomClaims struct {
	UserID   int    `json:"user_id"`
	Username string `json:"username"`
	Role     string `json:"role"`
	jwt.RegisteredClaims
}

// Valid 验证Claims有效性
func (c CustomClaims) Valid() error {
	// 验证自定义字段
	if c.UserID <= 0 {
		return jwt.ErrInvalidKey
	}

	if c.Username == "" {
		return jwt.ErrInvalidKey
	}

	return nil
}
