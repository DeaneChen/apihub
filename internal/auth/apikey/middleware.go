package apikey

import (
	"net/http"
	"strings"

	"apihub/internal/model"

	"github.com/gin-gonic/gin"
)

// ContextKey 上下文键类型
type ContextKey string

const (
	// APIKeyKey APIKey在上下文中的键
	APIKeyKey ContextKey = "api_key"
	// APIKeyUserIDKey APIKey用户ID在上下文中的键
	APIKeyUserIDKey ContextKey = "api_key_user_id"
)

// APIKeyAuthMiddleware APIKey认证中间件
func APIKeyAuthMiddleware(apiKeyService *APIKeyService) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 从请求头获取APIKey
		apiKey := getAPIKeyFromRequest(c)
		if apiKey == "" {
			c.JSON(http.StatusUnauthorized, model.NewErrorResponse(model.CodeUnauthorized, "缺少API密钥"))
			c.Abort()
			return
		}

		// 验证APIKey
		apiKeyModel, err := apiKeyService.ValidateAPIKey(apiKey)
		if err != nil {
			c.JSON(http.StatusUnauthorized, model.NewErrorResponse(model.CodeUnauthorized, "API密钥无效: "+err.Error()))
			c.Abort()
			return
		}

		// 将APIKey信息存入上下文
		c.Set(string(APIKeyKey), apiKeyModel)
		c.Set(string(APIKeyUserIDKey), apiKeyModel.UserID)

		c.Next()
	}
}

// OptionalAPIKeyAuthMiddleware 可选APIKey认证中间件
func OptionalAPIKeyAuthMiddleware(apiKeyService *APIKeyService) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 从请求获取APIKey
		apiKey := getAPIKeyFromRequest(c)
		if apiKey == "" {
			// 没有APIKey，继续执行
			c.Next()
			return
		}

		// 验证APIKey
		apiKeyModel, err := apiKeyService.ValidateAPIKey(apiKey)
		if err != nil {
			// APIKey无效，继续执行
			c.Next()
			return
		}

		// 将APIKey信息存入上下文
		c.Set(string(APIKeyKey), apiKeyModel)
		c.Set(string(APIKeyUserIDKey), apiKeyModel.UserID)

		c.Next()
	}
}

// RequireScopeMiddleware 要求特定权限范围的中间件
func RequireScopeMiddleware(requiredScope string) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 获取APIKey
		apiKeyModel, exists := GetAPIKey(c)
		if !exists {
			c.JSON(http.StatusUnauthorized, model.NewErrorResponse(model.CodeUnauthorized, "上下文中未找到API密钥"))
			c.Abort()
			return
		}

		// 检查权限范围 - 当前APIKey模型不包含Scopes字段，默认允许所有操作
		if !apiKeyModel.IsActive() {
			c.JSON(http.StatusForbidden, model.NewErrorResponse(model.CodeForbidden, "API密钥未激活"))
			c.Abort()
			return
		}

		c.Next()
	}
}

// getAPIKeyFromRequest 从请求中获取APIKey
func getAPIKeyFromRequest(c *gin.Context) string {
	// 1. 从Authorization头获取 (Bearer token格式)
	authHeader := c.GetHeader("Authorization")
	if authHeader != "" {
		const bearerPrefix = "Bearer "
		if strings.HasPrefix(authHeader, bearerPrefix) {
			return authHeader[len(bearerPrefix):]
		}
	}

	// 2. 从X-API-Key头获取
	apiKey := c.GetHeader("X-API-Key")
	if apiKey != "" {
		return apiKey
	}

	// 3. 从查询参数获取
	apiKey = c.Query("api_key")
	if apiKey != "" {
		return apiKey
	}

	return ""
}

// GetAPIKey 从上下文获取APIKey
func GetAPIKey(c *gin.Context) (*model.APIKey, bool) {
	apiKey, exists := c.Get(string(APIKeyKey))
	if !exists {
		return nil, false
	}

	apiKeyModel, ok := apiKey.(*model.APIKey)
	return apiKeyModel, ok
}

// GetAPIKeyUserID 从上下文获取APIKey用户ID
func GetAPIKeyUserID(c *gin.Context) (int, bool) {
	userID, exists := c.Get(string(APIKeyUserIDKey))
	if !exists {
		return 0, false
	}

	id, ok := userID.(int)
	return id, ok
}

// 注意：GetAPIKeyScopes和hasScope函数已移除
// 因为当前APIKey模型不包含Scopes字段
// 如果需要权限控制，可以在未来扩展APIKey模型时重新添加
