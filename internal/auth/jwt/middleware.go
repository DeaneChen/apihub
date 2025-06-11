package jwt

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

// ContextKey 上下文键类型
type ContextKey string

const (
	// UserClaimsKey 用户Claims在上下文中的键
	UserClaimsKey ContextKey = "user_claims"
	// UserIDKey 用户ID在上下文中的键
	UserIDKey ContextKey = "user_id"
	// UsernameKey 用户名在上下文中的键
	UsernameKey ContextKey = "username"
	// UserRoleKey 用户角色在上下文中的键
	UserRoleKey ContextKey = "user_role"
)

// JWTAuthMiddleware JWT认证中间件
func JWTAuthMiddleware(jwtService *JWTService) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 从请求头获取Token
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error":   "unauthorized",
				"message": "missing authorization header",
			})
			c.Abort()
			return
		}

		// 检查Bearer前缀
		const bearerPrefix = "Bearer "
		if !strings.HasPrefix(authHeader, bearerPrefix) {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error":   "unauthorized",
				"message": "invalid authorization header format",
			})
			c.Abort()
			return
		}

		// 提取Token
		tokenString := authHeader[len(bearerPrefix):]
		if tokenString == "" {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error":   "unauthorized",
				"message": "missing token",
			})
			c.Abort()
			return
		}

		// 验证Token
		claims, err := jwtService.ValidateToken(tokenString)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error":   "unauthorized",
				"message": "invalid token: " + err.Error(),
			})
			c.Abort()
			return
		}

		// 将用户信息存入上下文
		c.Set(string(UserClaimsKey), claims)
		c.Set(string(UserIDKey), claims.UserID)
		c.Set(string(UsernameKey), claims.Username)
		c.Set(string(UserRoleKey), claims.Role)

		c.Next()
	}
}

// OptionalJWTAuthMiddleware 可选JWT认证中间件
// 如果提供了Token则验证，否则继续执行
func OptionalJWTAuthMiddleware(jwtService *JWTService) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 从请求头获取Token
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			// 没有Token，继续执行
			c.Next()
			return
		}

		// 检查Bearer前缀
		const bearerPrefix = "Bearer "
		if !strings.HasPrefix(authHeader, bearerPrefix) {
			// 格式错误，继续执行
			c.Next()
			return
		}

		// 提取Token
		tokenString := authHeader[len(bearerPrefix):]
		if tokenString == "" {
			// Token为空，继续执行
			c.Next()
			return
		}

		// 验证Token
		claims, err := jwtService.ValidateToken(tokenString)
		if err != nil {
			// Token无效，继续执行
			c.Next()
			return
		}

		// 将用户信息存入上下文
		c.Set(string(UserClaimsKey), claims)
		c.Set(string(UserIDKey), claims.UserID)
		c.Set(string(UsernameKey), claims.Username)
		c.Set(string(UserRoleKey), claims.Role)

		c.Next()
	}
}

// GetUserClaims 从上下文获取用户Claims
func GetUserClaims(c *gin.Context) (*CustomClaims, bool) {
	claims, exists := c.Get(string(UserClaimsKey))
	if !exists {
		return nil, false
	}

	userClaims, ok := claims.(*CustomClaims)
	return userClaims, ok
}

// GetUserID 从上下文获取用户ID
func GetUserID(c *gin.Context) (int, bool) {
	userID, exists := c.Get(string(UserIDKey))
	if !exists {
		return 0, false
	}

	id, ok := userID.(int)
	return id, ok
}

// GetUsername 从上下文获取用户名
func GetUsername(c *gin.Context) (string, bool) {
	username, exists := c.Get(string(UsernameKey))
	if !exists {
		return "", false
	}

	name, ok := username.(string)
	return name, ok
}

// GetUserRole 从上下文获取用户角色
func GetUserRole(c *gin.Context) (string, bool) {
	role, exists := c.Get(string(UserRoleKey))
	if !exists {
		return "", false
	}

	userRole, ok := role.(string)
	return userRole, ok
}

// RequireRole 要求特定角色的中间件
func RequireRole(roles ...string) gin.HandlerFunc {
	return func(c *gin.Context) {
		userRole, exists := GetUserRole(c)
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error":   "unauthorized",
				"message": "user role not found",
			})
			c.Abort()
			return
		}

		// 检查用户角色是否在允许的角色列表中
		for _, role := range roles {
			if userRole == role {
				c.Next()
				return
			}
		}

		c.JSON(http.StatusForbidden, gin.H{
			"error":   "forbidden",
			"message": "insufficient permissions",
		})
		c.Abort()
	}
}
