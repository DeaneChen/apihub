package provider

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"apihub/internal/auth"
	"apihub/internal/auth/apikey"
	"apihub/internal/auth/jwt"
	"apihub/internal/middleware"
	"apihub/internal/model"
	"apihub/internal/provider/registry"
	"apihub/internal/store"

	"github.com/gin-gonic/gin"
)

// ProviderRouter 功能API路由器
type ProviderRouter struct {
	registry     *registry.ServiceRegistry
	authServices *auth.AuthServices
	store        store.Store
}

// NewProviderRouter 创建功能API路由器
func NewProviderRouter(registry *registry.ServiceRegistry, authServices *auth.AuthServices, store store.Store) *ProviderRouter {
	return &ProviderRouter{
		registry:     registry,
		authServices: authServices,
		store:        store,
	}
}

// RegisterRoutes 注册API路由
func (r *ProviderRouter) RegisterRoutes(router *gin.RouterGroup) {
	apiGroup := router.Group("/provider")

	// 服务状态检查端点
	apiGroup.GET("/status", r.statusHandler)

	// 服务列表端点
	apiGroup.GET("/services", r.listServicesHandler)

	// 服务信息端点
	apiGroup.GET("/:service/info", r.serviceInfoHandler)

	// 服务执行端点（带认证）
	authenticatedGroup := apiGroup.Group("/:service/execute")
	authenticatedGroup.Use(r.serviceAuthMiddleware()) // 先进行服务验证和用户认证
	authenticatedGroup.Use(r.logMiddleware())         // 然后记录日志
	authenticatedGroup.POST("", r.executeServiceHandler)

	// 公开API端点（可选认证）
	publicGroup := apiGroup.Group("/:service/public")
	publicGroup.Use(r.optionalAuthMiddleware()) // 先进行服务验证和可选用户认证
	publicGroup.Use(r.logMiddleware())          // 然后记录日志
	publicGroup.POST("", r.executePublicServiceHandler)
}

// statusHandler 服务状态检查处理函数
func (r *ProviderRouter) statusHandler(c *gin.Context) {
	c.JSON(http.StatusOK, model.NewSuccessResponse(gin.H{
		"status":        "ok",
		"service_count": r.registry.ServiceCount(),
		"service_names": r.registry.GetServiceNames(),
		"timestamp":     time.Now().Unix(),
	}))
}

// listServicesHandler 服务列表处理函数
func (r *ProviderRouter) listServicesHandler(c *gin.Context) {
	services := r.registry.ListServices()

	// 转换为响应格式
	response := make([]gin.H, 0, len(services))
	for _, service := range services {
		if service.Definition.IsEnabled() {
			response = append(response, gin.H{
				"service_name":    service.Definition.ServiceName,
				"description":     service.Definition.Description,
				"allow_anonymous": service.Definition.AllowAnonymous,
			})
		}
	}

	c.JSON(http.StatusOK, model.NewSuccessResponse(response))
}

// serviceInfoHandler 服务信息处理函数
func (r *ProviderRouter) serviceInfoHandler(c *gin.Context) {
	serviceName := c.Param("service")

	// 查找服务
	service, exists := r.registry.GetService(serviceName)
	if !exists {
		c.JSON(http.StatusNotFound, model.NewErrorResponse(
			model.CodeNotFound,
			"服务不存在",
		))
		return
	}

	// 检查服务状态
	if !service.Definition.IsEnabled() {
		c.JSON(http.StatusForbidden, model.NewErrorResponse(
			model.CodeForbidden,
			"服务已禁用",
		))
		return
	}

	// 返回服务信息
	c.JSON(http.StatusOK, model.NewSuccessResponse(service.Definition.ToResponse()))
}

// serviceAuthMiddleware 服务认证中间件
func (r *ProviderRouter) serviceAuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 获取服务名称
		serviceName := c.Param("service")

		// 查找服务
		service, exists := r.registry.GetService(serviceName)
		if !exists {
			c.JSON(http.StatusNotFound, model.NewErrorResponse(
				model.CodeNotFound,
				"服务不存在",
			))
			c.Abort()
			return
		}

		// 检查服务状态
		if !service.Definition.IsEnabled() {
			c.JSON(http.StatusForbidden, model.NewErrorResponse(
				model.CodeForbidden,
				"服务已禁用",
			))
			c.Abort()
			return
		}

		// 将服务信息存入上下文
		c.Set("service_info", service)

		// 检查是否允许匿名访问
		if !service.Definition.AllowAnonymous {
			// 使用现有的认证中间件
			middleware.AuthMiddleware(r.authServices.JWTService, r.authServices.APIKeyService)(c)
			if c.IsAborted() {
				return
			}
		} else {
			middleware.OptionalAuthMiddleware(r.authServices.JWTService, r.authServices.APIKeyService)(c)
		}
	}
}

// optionalAuthMiddleware 可选认证中间件
func (r *ProviderRouter) optionalAuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 获取服务名称
		serviceName := c.Param("service")

		// 查找服务
		service, exists := r.registry.GetService(serviceName)
		if !exists {
			c.JSON(http.StatusNotFound, model.NewErrorResponse(
				model.CodeNotFound,
				"服务不存在",
			))
			c.Abort()
			return
		}

		// 检查服务状态
		if !service.Definition.IsEnabled() {
			c.JSON(http.StatusForbidden, model.NewErrorResponse(
				model.CodeForbidden,
				"服务已禁用",
			))
			c.Abort()
			return
		}

		// 将服务信息存入上下文 - 无论认证是否成功，都需要设置服务信息
		c.Set("service_info", service)

		// 使用现有的可选认证中间件，它会自动调用c.Next()
		middleware.OptionalAuthMiddleware(r.authServices.JWTService, r.authServices.APIKeyService)(c)

		// 不需要再次调用c.Next()，因为OptionalAuthMiddleware已经调用过了
	}
}

// logMiddleware 日志中间件
func (r *ProviderRouter) logMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 在请求开始时获取服务信息
		serviceInfo, exists := c.Get("service_info")
		if !exists {
			fmt.Printf("日志中间件：未找到服务信息\n")
			c.Next()
			return
		}
		service := serviceInfo.(*registry.ServiceInfo)

		// 处理请求
		c.Next()

		// 获取用户ID和APIKey ID
		var userID int
		var apiKeyID int

		// 使用middleware包中的函数获取用户ID
		userIDFromAuth, exists := middleware.GetCurrentUserID(c)
		if exists {
			userID = userIDFromAuth
		} else {
			// 尝试从JWT获取用户ID（兼容旧代码）
			userIDFromJWT, exists := jwt.GetUserID(c)
			if exists {
				userID = userIDFromJWT
			}
		}

		// 尝试从APIKey获取用户ID和APIKey ID
		apiKey, exists := apikey.GetAPIKey(c)
		if exists {
			apiKeyID = apiKey.ID
			if userID == 0 {
				userID = apiKey.UserID
			}
		}

		// 创建访问日志
		accessLog := &model.AccessLog{
			APIKeyID:    apiKeyID, // 即使为0也允许，不强制外键约束
			UserID:      userID,   // 即使为0也允许，不强制外键约束
			ServiceName: service.Definition.ServiceName,
			Endpoint:    c.Request.URL.Path,
			Status:      c.Writer.Status(),
			Cost:        service.Definition.QuotaCost,
			CreatedAt:   time.Now(),
		}

		// 异步保存访问日志
		go func() {
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()

			if err := r.store.AccessLogs().Create(ctx, accessLog); err != nil {
				fmt.Printf("保存访问日志失败: %v\n", err)
			} else {
				fmt.Printf("成功记录访问日志: 用户ID=%d, 服务=%s, 状态=%d\n",
					userID, service.Definition.ServiceName, c.Writer.Status())
			}

			// 如果有用户ID和配额成本，增加使用量
			if userID > 0 && service.Definition.QuotaCost > 0 {
				// 检查配额
				quota, err := r.store.Quotas().GetByUserAndService(ctx, userID, service.Definition.ServiceName, "daily")
				if err != nil {
					// 配额不存在，创建默认配额
					quota = &model.ServiceQuota{
						UserID:      userID,
						ServiceName: service.Definition.ServiceName,
						TimeWindow:  "daily",
						Usage:       0,
						LimitValue:  service.Definition.DefaultLimit,
						ResetTime:   time.Now().Add(24 * time.Hour),
					}
					if err := r.store.Quotas().Create(ctx, quota); err != nil {
						fmt.Printf("创建配额失败: %v\n", err)
					}
				}

				// 增加使用量
				if err := r.store.Quotas().IncrementUsage(ctx, userID, service.Definition.ServiceName, "daily", service.Definition.QuotaCost); err != nil {
					fmt.Printf("增加使用量失败: %v\n", err)
				}
			}
		}()
	}
}

// executeServiceHandler 执行服务处理函数
func (r *ProviderRouter) executeServiceHandler(c *gin.Context) {
	// 获取服务信息
	serviceInfo, exists := c.Get("service_info")
	if !exists {
		c.JSON(http.StatusInternalServerError, model.NewErrorResponse(
			model.CodeInternalError,
			"服务信息不存在",
		))
		return
	}

	service := serviceInfo.(*registry.ServiceInfo)

	// 执行服务处理函数
	result, err := service.Handler(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, model.NewErrorResponse(
			model.CodeInvalidParams,
			err.Error(),
		))
		return
	}

	// 返回结果
	c.JSON(http.StatusOK, model.NewSuccessResponse(result))
}

// executePublicServiceHandler 执行公开服务处理函数
func (r *ProviderRouter) executePublicServiceHandler(c *gin.Context) {
	// 获取服务信息
	serviceInfo, exists := c.Get("service_info")
	if !exists {
		c.JSON(http.StatusInternalServerError, model.NewErrorResponse(
			model.CodeInternalError,
			"服务信息不存在",
		))
		return
	}

	service := serviceInfo.(*registry.ServiceInfo)

	// 执行服务处理函数
	result, err := service.Handler(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, model.NewErrorResponse(
			model.CodeInvalidParams,
			err.Error(),
		))
		return
	}

	// 返回结果
	c.JSON(http.StatusOK, model.NewSuccessResponse(result))
}
