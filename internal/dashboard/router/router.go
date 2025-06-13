package router

import (
	"apihub/internal/auth"
	"apihub/internal/model"
	"apihub/internal/store"

	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

// @title           APIHub API
// @version         1.0
// @description     统一API业务服务框架，实现多种功能性服务API并集中管理
// @termsOfService  http://swagger.io/terms/

// @contact.name   API Support
// @contact.url    http://www.swagger.io/support
// @contact.email  support@swagger.io

// @license.name  Apache 2.0
// @license.url   http://www.apache.org/licenses/LICENSE-2.0.html

// @host      localhost:8080
// @BasePath  /api/v1

// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
// @description 请输入 "Bearer {token}" 格式的JWT令牌

// @securityDefinitions.apikey ApiKeyAuth
// @in header
// @name X-API-Key
// @description API Key 认证

// Router 主路由器
type Router struct {
	authRouter   *AuthRouter
	apiKeyRouter *APIKeyRouter
	userRouter   *UserRouter
	authServices *auth.AuthServices
}

// NewRouter 创建主路由器实例
func NewRouter(store store.Store, authServices *auth.AuthServices) *Router {
	return &Router{
		authRouter:   NewAuthRouter(store, authServices),
		apiKeyRouter: NewAPIKeyRouter(store, authServices),
		userRouter:   NewUserRouter(store, authServices.JWTService),
		authServices: authServices,
	}
}

// SetupRoutes 设置所有路由
func (r *Router) SetupRoutes() *gin.Engine {
	// 创建Gin引擎
	engine := gin.Default()

	// 添加全局中间件
	engine.Use(gin.Logger())
	engine.Use(gin.Recovery())
	engine.Use(corsMiddleware())

	// Swagger文档路由
	engine.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	// API版本1路由组
	v1 := engine.Group("/api/v1")
	{
		// 健康检查
		v1.GET("/health", healthCheck)

		// 认证相关路由
		r.authRouter.RegisterRoutes(v1)

		// Dashboard路由（需要JWT认证）
		dashboardGroup := v1.Group("/dashboard")
		r.authRouter.RegisterDashboardRoutes(dashboardGroup)

		// API密钥路由（需要JWT认证）
		r.apiKeyRouter.RegisterRoutes(dashboardGroup)

		// 用户管理路由（需要JWT认证）
		r.userRouter.RegisterRoutes(dashboardGroup)

		// API路由（支持JWT和APIKey认证）
		r.authRouter.RegisterAPIRoutes(v1)
	}

	return engine
}

// @Summary      健康检查接口
// @Description  返回服务健康状态
// @Tags         系统
// @Produce      json
// @Success      200  {object}  model.APIResponse
// @Router       /health [get]
// healthCheck 健康检查接口
func healthCheck(c *gin.Context) {
	c.JSON(200, model.NewSuccessResponse(gin.H{
		"status":  "ok",
		"service": "apihub",
		"version": "1.0.0",
	}))
}

// corsMiddleware CORS中间件
func corsMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Origin, Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, X-API-Key")
		c.Header("Access-Control-Expose-Headers", "Content-Length")
		c.Header("Access-Control-Allow-Credentials", "true")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	}
}
