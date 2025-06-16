package middleware

import (
	"net/http"

	"apihub/internal/auth/apikey"
	"apihub/internal/auth/jwt"
	"apihub/internal/model"

	"github.com/gin-gonic/gin"
)

// AuthMiddleware 统一认证中间件
// 支持JWT和APIKey两种认证方式
func AuthMiddleware(jwtService *jwt.JWTService, apiKeyService *apikey.APIKeyService) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 首先尝试JWT认证
		authHeader := c.GetHeader("Authorization")
		if authHeader != "" && len(authHeader) > 7 && authHeader[:7] == "Bearer " {
			tokenString := authHeader[7:]

			// 验证JWT Token
			claims, err := jwtService.ValidateToken(tokenString)
			if err == nil {
				// JWT认证成功，设置用户信息到上下文
				c.Set(string(jwt.UserClaimsKey), claims)
				c.Set(string(jwt.UserIDKey), claims.UserID)
				c.Set(string(jwt.UsernameKey), claims.Username)
				c.Set(string(jwt.UserRoleKey), claims.Role)
				c.Next()
				return
			}
		}

		// JWT认证失败，尝试APIKey认证
		apiKeyString := getAPIKeyFromRequest(c)
		if apiKeyString != "" {
			// 验证APIKey
			apiKeyModel, err := apiKeyService.ValidateAPIKey(apiKeyString)
			if err == nil {
				// APIKey认证成功，设置APIKey信息到上下文
				c.Set(string(apikey.APIKeyKey), apiKeyModel)
				c.Set(string(apikey.APIKeyUserIDKey), apiKeyModel.UserID)
				c.Next()
				return
			}
		}

		// 两种认证方式都失败
		c.JSON(http.StatusUnauthorized, model.NewErrorResponse(model.CodeUnauthorized, "需要身份认证"))
		c.Abort()
	}
}

// OptionalAuthMiddleware 可选认证中间件
// 如果提供了认证信息则验证，否则继续执行
func OptionalAuthMiddleware(jwtService *jwt.JWTService, apiKeyService *apikey.APIKeyService) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 首先尝试JWT认证
		authHeader := c.GetHeader("Authorization")
		if authHeader != "" && len(authHeader) > 7 && authHeader[:7] == "Bearer " {
			tokenString := authHeader[7:]

			// 验证JWT Token
			claims, err := jwtService.ValidateToken(tokenString)
			if err == nil {
				// JWT认证成功，设置用户信息到上下文
				c.Set(string(jwt.UserClaimsKey), claims)
				c.Set(string(jwt.UserIDKey), claims.UserID)
				c.Set(string(jwt.UsernameKey), claims.Username)
				c.Set(string(jwt.UserRoleKey), claims.Role)
				// 不要立即返回，继续执行后续中间件
			}
		}

		// 如果JWT认证失败或没有JWT，尝试APIKey认证
		if _, exists := c.Get(string(jwt.UserIDKey)); !exists {
			apiKeyString := getAPIKeyFromRequest(c)
			if apiKeyString != "" {
				// 验证APIKey
				apiKeyModel, err := apiKeyService.ValidateAPIKey(apiKeyString)
				if err == nil {
					// APIKey认证成功，设置APIKey信息到上下文
					c.Set(string(apikey.APIKeyKey), apiKeyModel)
					c.Set(string(apikey.APIKeyUserIDKey), apiKeyModel.UserID)
					// 不要立即返回，继续执行后续中间件
				}
			}
		}

		// 无论认证是否成功，都继续执行后续中间件
		c.Next()
	}
}

// JWTOnlyMiddleware 仅JWT认证中间件
func JWTOnlyMiddleware(jwtService *jwt.JWTService) gin.HandlerFunc {
	return jwt.JWTAuthMiddleware(jwtService)
}

// APIKeyOnlyMiddleware 仅APIKey认证中间件
func APIKeyOnlyMiddleware(apiKeyService *apikey.APIKeyService) gin.HandlerFunc {
	return apikey.APIKeyAuthMiddleware(apiKeyService)
}

// GetCurrentUserID 获取当前用户ID（支持JWT和APIKey）
func GetCurrentUserID(c *gin.Context) (int, bool) {
	// 首先尝试从JWT获取
	if userID, exists := c.Get(string(jwt.UserIDKey)); exists {
		if id, ok := userID.(int); ok {
			return id, true
		}
	}

	// 然后尝试从APIKey获取
	if userID, exists := c.Get(string(apikey.APIKeyUserIDKey)); exists {
		if id, ok := userID.(int); ok {
			return id, true
		}
	}

	return 0, false
}

// GetCurrentUsername 获取当前用户名（仅JWT支持）
func GetCurrentUsername(c *gin.Context) (string, bool) {
	username, exists := c.Get(string(jwt.UsernameKey))
	if !exists {
		return "", false
	}

	name, ok := username.(string)
	return name, ok
}

// GetCurrentUserRole 获取当前用户角色（仅JWT支持）
func GetCurrentUserRole(c *gin.Context) (string, bool) {
	role, exists := c.Get(string(jwt.UserRoleKey))
	if !exists {
		return "", false
	}

	userRole, ok := role.(string)
	return userRole, ok
}

// IsJWTAuth 检查是否为JWT认证
func IsJWTAuth(c *gin.Context) bool {
	_, exists := c.Get(string(jwt.UserClaimsKey))
	return exists
}

// IsAPIKeyAuth 检查是否为APIKey认证
func IsAPIKeyAuth(c *gin.Context) bool {
	_, exists := c.Get(string(apikey.APIKeyKey))
	return exists
}

// getAPIKeyFromRequest 从请求中获取APIKey
func getAPIKeyFromRequest(c *gin.Context) string {
	// 1. 从X-API-Key头获取
	apiKey := c.GetHeader("X-API-Key")
	if apiKey != "" {
		return apiKey
	}

	// 2. 从查询参数获取
	apiKey = c.Query("api_key")
	if apiKey != "" {
		return apiKey
	}

	return ""
}
