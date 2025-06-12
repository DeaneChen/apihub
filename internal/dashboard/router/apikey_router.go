package router

import (
	"apihub/internal/auth"
	"apihub/internal/auth/jwt"
	"apihub/internal/dashboard/handler"
	"apihub/internal/middleware"
	"apihub/internal/store"

	"github.com/gin-gonic/gin"
)

// APIKeyRouter API密钥路由
type APIKeyRouter struct {
	apiKeyHandler *handler.APIKeyHandler
	jwtService    *jwt.JWTService
}

// NewAPIKeyRouter 创建API密钥路由实例
func NewAPIKeyRouter(store store.Store, authServices *auth.AuthServices) *APIKeyRouter {
	// 创建API密钥处理器
	apiKeyHandler := handler.NewAPIKeyHandler(authServices.APIKeyService)

	return &APIKeyRouter{
		apiKeyHandler: apiKeyHandler,
		jwtService:    authServices.JWTService,
	}
}

// RegisterRoutes 注册API密钥相关路由
func (r *APIKeyRouter) RegisterRoutes(router *gin.RouterGroup) {
	// API密钥路由组，需要JWT认证
	apiKeyGroup := router.Group("/apikeys")
	apiKeyGroup.Use(middleware.JWTOnlyMiddleware(r.jwtService))

	{
		// @Summary      列出用户的API密钥
		// @Description  列出当前用户的所有API密钥
		// @Tags         API密钥
		// @Accept       json
		// @Produce      json
		// @Security     BearerAuth
		// @Success      200  {object}  model.APIResponse{data=[]model.APIKey}
		// @Failure      401  {object}  model.APIResponse
		// @Router       /api/v1/dashboard/apikeys/list [get]
		apiKeyGroup.GET("/list", r.apiKeyHandler.ListAPIKeys)

		// @Summary      生成API密钥
		// @Description  为当前用户生成新的API密钥
		// @Tags         API密钥
		// @Accept       json
		// @Produce      json
		// @Security     BearerAuth
		// @Param        request body handler.GenerateAPIKeyRequest true "API密钥生成请求"
		// @Success      200  {object}  model.APIResponse{data=model.APIKey}
		// @Failure      400  {object}  model.APIResponse
		// @Failure      401  {object}  model.APIResponse
		// @Router       /api/v1/dashboard/apikeys/generate [post]
		apiKeyGroup.POST("/generate", r.apiKeyHandler.GenerateAPIKey)

		// @Summary      删除API密钥
		// @Description  删除当前用户的指定API密钥
		// @Tags         API密钥
		// @Accept       json
		// @Produce      json
		// @Security     BearerAuth
		// @Param        request body handler.DeleteAPIKeyRequest true "API密钥删除请求"
		// @Success      200  {object}  model.APIResponse
		// @Failure      400  {object}  model.APIResponse
		// @Failure      401  {object}  model.APIResponse
		// @Failure      403  {object}  model.APIResponse
		// @Router       /api/v1/dashboard/apikeys/delete [post]
		apiKeyGroup.POST("/delete", r.apiKeyHandler.DeleteAPIKey)
	}
}
