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

		// @Summary      用户登录
		// @Description  用户登录并获取JWT令牌
		// @Tags         认证
		// @Accept       json
		// @Produce      json
		// @Param        credentials  body      object  true  "登录凭证"
		// @Success      200          {object}  model.APIResponse
		// @Failure      400          {object}  model.APIResponse
		// @Failure      401          {object}  model.APIResponse
		// @Router       /api/v1/auth/login [post]
		authGroup.POST("/login", r.authHandler.Login)

		// 需要认证的路由
		protected := authGroup.Group("")
		protected.Use(middleware.JWTOnlyMiddleware(r.authService.JWTService))
		{
			// @Summary      用户登出
			// @Description  使当前JWT令牌失效
			// @Tags         认证
			// @Accept       json
			// @Produce      json
			// @Security     BearerAuth
			// @Success      200  {object}  model.APIResponse
			// @Failure      401  {object}  model.APIResponse
			// @Router       /api/v1/auth/logout [post]
			protected.POST("/logout", r.authHandler.Logout)

			// @Summary      获取用户资料
			// @Description  获取当前登录用户的资料信息
			// @Tags         认证
			// @Accept       json
			// @Produce      json
			// @Security     BearerAuth
			// @Success      200  {object}  model.APIResponse
			// @Failure      401  {object}  model.APIResponse
			// @Router       /api/v1/auth/profile [get]
			protected.GET("/profile", r.authHandler.GetProfile)

			// @Summary      更新个人资料
			// @Description  更新当前登录用户的个人资料
			// @Tags         认证
			// @Accept       json
			// @Produce      json
			// @Security     BearerAuth
			// @Param        request  body      model.UpdateProfileRequest  true  "更新个人资料请求"
			// @Success      200      {object}  model.APIResponse{data=model.UpdateProfileResponse}
			// @Failure      400      {object}  model.APIResponse
			// @Failure      401      {object}  model.APIResponse
			// @Router       /api/v1/auth/profile/update [post]
			protected.POST("/profile/update", r.authHandler.UpdateProfile)

			// @Summary      修改密码
			// @Description  修改当前登录用户的密码
			// @Tags         认证
			// @Accept       json
			// @Produce      json
			// @Security     BearerAuth
			// @Param        request  body      model.ChangePasswordRequest  true  "修改密码请求"
			// @Success      200      {object}  model.APIResponse{data=model.ChangePasswordResponse}
			// @Failure      400      {object}  model.APIResponse
			// @Failure      401      {object}  model.APIResponse
			// @Router       /api/v1/auth/password/change [post]
			protected.POST("/password/change", r.authHandler.ChangePassword)
		}
	}
}

// RegisterDashboardRoutes 注册Dashboard路由（需要JWT认证）
func (r *AuthRouter) RegisterDashboardRoutes(dashboardGroup *gin.RouterGroup) {
	// 添加JWT认证中间件
	dashboardGroup.Use(middleware.JWTOnlyMiddleware(r.authService.JWTService))

	// @Summary      获取用户资料
	// @Description  获取当前登录用户的资料信息（Dashboard版本）
	// @Tags         仪表盘
	// @Accept       json
	// @Produce      json
	// @Security     BearerAuth
	// @Success      200  {object}  model.APIResponse
	// @Failure      401  {object}  model.APIResponse
	// @Router       /api/v1/dashboard/profile [get]
	dashboardGroup.GET("/profile", r.authHandler.GetProfile)

	// @Summary      更新个人资料
	// @Description  更新当前登录用户的个人资料（Dashboard版本）
	// @Tags         仪表盘
	// @Accept       json
	// @Produce      json
	// @Security     BearerAuth
	// @Param        request  body      model.UpdateProfileRequest  true  "更新个人资料请求"
	// @Success      200      {object}  model.APIResponse{data=model.UpdateProfileResponse}
	// @Failure      400      {object}  model.APIResponse
	// @Failure      401      {object}  model.APIResponse
	// @Router       /api/v1/dashboard/profile/update [post]
	dashboardGroup.POST("/profile/update", r.authHandler.UpdateProfile)

	// @Summary      修改密码
	// @Description  修改当前登录用户的密码（Dashboard版本）
	// @Tags         仪表盘
	// @Accept       json
	// @Produce      json
	// @Security     BearerAuth
	// @Param        request  body      model.ChangePasswordRequest  true  "修改密码请求"
	// @Success      200      {object}  model.APIResponse{data=model.ChangePasswordResponse}
	// @Failure      400      {object}  model.APIResponse
	// @Failure      401      {object}  model.APIResponse
	// @Router       /api/v1/dashboard/password/change [post]
	dashboardGroup.POST("/password/change", r.authHandler.ChangePassword)

	// TODO: 添加其他dashboard功能路由
	// dashboardGroup.GET("/users", userHandler.ListUsers)
	// dashboardGroup.POST("/users", userHandler.CreateUser)
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
