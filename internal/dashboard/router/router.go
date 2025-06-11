package router

import (
	"apihub/internal/auth"
	"apihub/internal/model"
	"apihub/internal/store"

	"github.com/gin-gonic/gin"
)

// Router 主路由器
type Router struct {
	authRouter *AuthRouter
}

// NewRouter 创建主路由器实例
func NewRouter(store store.Store, authServices *auth.AuthServices) *Router {
	return &Router{
		authRouter: NewAuthRouter(store, authServices),
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

	// API版本1路由组
	v1 := engine.Group("/api/v1")
	{
		// 健康检查
		v1.GET("/health", healthCheck)

		// 认证相关路由
		r.authRouter.RegisterRoutes(v1)

		// Dashboard路由（需要JWT认证）
		r.authRouter.RegisterDashboardRoutes(v1)

		// API路由（支持JWT和APIKey认证）
		r.authRouter.RegisterAPIRoutes(v1)
	}

	return engine
}

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
