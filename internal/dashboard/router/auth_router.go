package router

import (
	"apihub/internal/auth"
	"apihub/internal/dashboard/handler"
	"apihub/internal/dashboard/service"
	"apihub/internal/middleware"
	"apihub/internal/store"

	"github.com/gin-gonic/gin"
)

// AuthRouter 认证路由
type AuthRouter struct {
	authHandler *handler.AuthHandler
	authService *auth.AuthServices
}

// NewAuthRouter 创建认证路由实例
func NewAuthRouter(store store.Store, authServices *auth.AuthServices) *AuthRouter {
	// 创建认证服务
	authService := service.NewAuthService(store, authServices.JWTService)

	// 创建认证处理器
	authHandler := handler.NewAuthHandler(authService)

	return &AuthRouter{
		authHandler: authHandler,
		authService: authServices,
	}
}

// RegisterRoutes 注册认证相关路由
func (r *AuthRouter) RegisterRoutes(router *gin.RouterGroup) {
	// 认证路由组
	authGroup := router.Group("/auth")
	{
		// 公开路由（无需认证）
		authGroup.POST("/login", r.authHandler.Login)

		// 需要认证的路由
		protected := authGroup.Group("")
		protected.Use(middleware.JWTOnlyMiddleware(r.authService.JWTService))
		{
			protected.POST("/logout", r.authHandler.Logout)
			protected.GET("/profile", r.authHandler.GetProfile)
		}
	}
}

// RegisterDashboardRoutes 注册Dashboard路由（需要JWT认证）
func (r *AuthRouter) RegisterDashboardRoutes(router *gin.RouterGroup) {
	// Dashboard路由组，需要JWT认证
	dashboardGroup := router.Group("/dashboard")
	dashboardGroup.Use(middleware.JWTOnlyMiddleware(r.authService.JWTService))
	{
		// 这里可以添加其他dashboard相关的路由
		// 例如：用户管理、API密钥管理等

		// 示例：获取用户信息
		dashboardGroup.GET("/profile", r.authHandler.GetProfile)

		// TODO: 添加其他dashboard功能路由
		// dashboardGroup.GET("/users", userHandler.ListUsers)
		// dashboardGroup.POST("/users", userHandler.CreateUser)
		// dashboardGroup.GET("/api-keys", apiKeyHandler.ListAPIKeys)
		// dashboardGroup.POST("/api-keys", apiKeyHandler.CreateAPIKey)
	}
}

// RegisterAPIRoutes 注册API路由（支持JWT和APIKey认证）
func (r *AuthRouter) RegisterAPIRoutes(router *gin.RouterGroup) {
	// API路由组，支持JWT和APIKey双重认证
	apiGroup := router.Group("/api")
	apiGroup.Use(middleware.AuthMiddleware(r.authService.JWTService, r.authService.APIKeyService))
	{
		// 这里可以添加需要认证的API路由
		// 例如：服务调用、数据查询等

		// TODO: 添加具体的API功能路由
		// apiGroup.GET("/services", serviceHandler.ListServices)
		// apiGroup.POST("/services/:name/execute", serviceHandler.ExecuteService)
	}
}
