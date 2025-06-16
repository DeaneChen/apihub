# APIHub 功能API提供者框架设计

## 1. 设计概述

APIHub功能API提供者框架是整个系统的核心组件之一，负责注册、管理和路由各种功能性API服务。本文档详细描述了该框架的设计方案，包括架构组件、数据模型、实现方式和工作流程。

### 1.1 设计目标

- **统一管理**：集中管理所有功能性API服务
- **简单易用**：采用代码注册方式，避免复杂的插件机制
- **安全可控**：集成认证、授权、限流和配额控制
- **可扩展性**：便于添加新的服务和功能
- **可监控性**：记录访问日志和统计信息

### 1.2 核心功能

- 服务注册与管理
- 请求路由与分发
- 认证与授权控制
- 限流与配额管理
- 访问日志记录

## 2. 架构设计

### 2.1 整体架构

功能API提供者框架主要由以下组件组成：

1. **服务注册中心(Registry)**：管理所有功能性API服务
2. **服务路由(Router)**：将请求路由到对应的服务处理函数
3. **中间件(Middleware)**：提供认证、限流、配额控制等功能
4. **服务实现(Services)**：具体功能API的实现

### 2.2 组件关系图

```
                  ┌─────────────────┐
                  │    HTTP请求     │
                  └────────┬────────┘
                           │
                           ▼
                  ┌─────────────────┐
                  │    API路由器    │
                  └────────┬────────┘
                           │
                           ▼
          ┌────────────────┴────────────────┐
          │                                 │
          ▼                                 ▼
┌─────────────────┐               ┌─────────────────┐
│  认证中间件     │               │  公开API路由    │
└────────┬────────┘               └────────┬────────┘
          │                                 │
          ▼                                 ▼
┌─────────────────┐               ┌─────────────────┐
│  限流中间件     │               │ 可选认证中间件  │
└────────┬────────┘               └────────┬────────┘
          │                                 │
          ▼                                 ▼
┌─────────────────┐               ┌─────────────────┐
│  配额中间件     │               │  限流中间件     │
└────────┬────────┘               └────────┬────────┘
          │                                 │
          ▼                                 ▼
┌─────────────────┐               ┌─────────────────┐
│  日志中间件     │               │  日志中间件     │
└────────┬────────┘               └────────┬────────┘
          │                                 │
          ▼                                 ▼
┌─────────────────┐               ┌─────────────────┐
│ 服务注册中心    │◄──────────────┤ 服务注册中心    │
└────────┬────────┘               └────────┬────────┘
          │                                 │
          ▼                                 ▼
┌─────────────────┐               ┌─────────────────┐
│  服务处理函数   │               │  服务处理函数   │
└─────────────────┘               └─────────────────┘
```

## 3. 数据模型

### 3.1 服务定义模型

服务定义模型已存在于`model`包中，可以继续使用：

```go
// ServiceDefinition 服务定义
type ServiceDefinition struct {
    ID           int       `json:"id" db:"id"`
    ServiceName  string    `json:"service_name" db:"service_name"`
    Description  string    `json:"description" db:"description"`
    DefaultLimit int       `json:"default_limit" db:"default_limit"`
    Status       int       `json:"status" db:"status"` // 0: disabled, 1: enabled
    CreatedAt    time.Time `json:"created_at" db:"created_at"`
    UpdatedAt    time.Time `json:"updated_at" db:"updated_at"`
}

// 服务状态常量
const (
    ServiceStatusDisabled = 0  // 禁用
    ServiceStatusEnabled  = 1  // 启用
)
```

### 3.2 服务注册中心模型

```go
// ServiceHandler 服务处理函数类型
type ServiceHandler func(c *gin.Context) (interface{}, error)

// ServiceConfig 服务配置
type ServiceConfig struct {
    // 是否允许匿名访问
    AllowAnonymous bool
    // 默认限流配置（每分钟请求数）
    RateLimit int
    // 默认消耗配额
    QuotaCost int
    // 服务描述信息（用于生成API文档）
    Description string
    // 请求示例
    RequestExample interface{}
    // 响应示例
    ResponseExample interface{}
}

// ServiceInfo 服务信息
type ServiceInfo struct {
    // 服务定义（来自数据库）
    Definition *model.ServiceDefinition
    // 服务配置
    Config ServiceConfig
    // 服务处理函数
    Handler ServiceHandler
}

// ServiceRegistry 服务注册中心
type ServiceRegistry struct {
    // 服务映射表 serviceName -> ServiceInfo
    services map[string]*ServiceInfo
    // 存储层接口
    store store.Store
    // 互斥锁，保护services映射表
    mu sync.RWMutex
}
```

## 4. 功能实现

### 4.1 服务注册中心

服务注册中心是整个框架的核心，负责管理所有功能性API服务。

```go
// NewServiceRegistry 创建服务注册中心
func NewServiceRegistry(store store.Store) *ServiceRegistry {
    return &ServiceRegistry{
        services: make(map[string]*ServiceInfo),
        store:    store,
    }
}

// RegisterService 注册服务
func (r *ServiceRegistry) RegisterService(name string, handler ServiceHandler, config ServiceConfig) error {
    r.mu.Lock()
    defer r.mu.Unlock()
    
    // 检查服务是否已存在
    if _, exists := r.services[name]; exists {
        return fmt.Errorf("服务 %s 已存在", name)
    }
    
    // 从数据库获取服务定义
    definition, err := r.store.Services().GetByName(context.Background(), name)
    if err != nil {
        // 服务定义不存在，创建默认定义
        definition = &model.ServiceDefinition{
            ServiceName:  name,
            Description:  config.Description,
            DefaultLimit: 1000, // 默认每日配额
            Status:       model.ServiceStatusEnabled,
        }
        
        // 保存到数据库
        if err := r.store.Services().Create(context.Background(), definition); err != nil {
            return fmt.Errorf("创建服务定义失败: %w", err)
        }
    }
    
    // 注册服务
    r.services[name] = &ServiceInfo{
        Definition: definition,
        Config:     config,
        Handler:    handler,
    }
    
    return nil
}

// GetService 获取服务
func (r *ServiceRegistry) GetService(name string) (*ServiceInfo, bool) {
    r.mu.RLock()
    defer r.mu.RUnlock()
    
    service, exists := r.services[name]
    return service, exists
}

// ListServices 列出所有服务
func (r *ServiceRegistry) ListServices() []*ServiceInfo {
    r.mu.RLock()
    defer r.mu.RUnlock()
    
    services := make([]*ServiceInfo, 0, len(r.services))
    for _, service := range r.services {
        services = append(services, service)
    }
    
    return services
}
```

### 4.2 服务路由

服务路由负责将请求路由到对应的服务处理函数。

```go
// ProviderRouter 功能API路由器
type ProviderRouter struct {
    registry    *ServiceRegistry
    authServices *auth.AuthServices
    store       store.Store
}

// NewProviderRouter 创建功能API路由器
func NewProviderRouter(registry *ServiceRegistry, authServices *auth.AuthServices, store store.Store) *ProviderRouter {
    return &ProviderRouter{
        registry:    registry,
        authServices: authServices,
        store:       store,
    }
}

// RegisterRoutes 注册API路由
func (r *ProviderRouter) RegisterRoutes(router *gin.RouterGroup) {
    apiGroup := router.Group("/api/v1")
    
    // 服务状态检查端点
    apiGroup.GET("/status", r.statusHandler)
    
    // 服务列表端点
    apiGroup.GET("/services", r.listServicesHandler)
    
    // 服务信息端点
    apiGroup.GET("/:service/info", r.serviceInfoHandler)
    
    // 服务执行端点（带认证）
    authenticatedGroup := apiGroup.Group("/:service/execute")
    authenticatedGroup.Use(r.serviceAuthMiddleware())
    authenticatedGroup.Use(r.rateLimitMiddleware())
    authenticatedGroup.Use(r.quotaMiddleware())
    authenticatedGroup.Use(r.logMiddleware())
    authenticatedGroup.POST("", r.executeServiceHandler)
    
    // 公开API端点（可选认证）
    publicGroup := apiGroup.Group("/:service/public")
    publicGroup.Use(r.optionalAuthMiddleware())
    publicGroup.Use(r.publicRateLimitMiddleware())
    publicGroup.Use(r.logMiddleware())
    publicGroup.POST("", r.executePublicServiceHandler)
}
```

### 4.3 中间件

中间件提供认证、限流、配额控制等功能。

#### 4.3.1 认证中间件

```go
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
        if service.Definition.Status != model.ServiceStatusEnabled {
            c.JSON(http.StatusForbidden, model.NewErrorResponse(
                model.CodeServiceDisabled,
                "服务已禁用",
            ))
            c.Abort()
            return
        }
        
        // 检查是否允许匿名访问
        if !service.Config.AllowAnonymous {
            // 使用现有的认证中间件
            middleware.AuthMiddleware(r.authServices.JWTService, r.authServices.APIKeyService)(c)
            if c.IsAborted() {
                return
            }
        }
        
        // 将服务信息存入上下文
        c.Set("service_info", service)
        
        c.Next()
    }
}
```

#### 4.3.2 限流中间件

```go
// rateLimitMiddleware 限流中间件
func (r *ProviderRouter) rateLimitMiddleware() gin.HandlerFunc {
    // 使用内存缓存实现简单的限流
    limiter := rate.NewLimiter(10, 30) // 默认限流：10 QPS，突发30个请求
    
    return func(c *gin.Context) {
        // 获取用户标识（APIKey ID或IP地址）
        var key string
        apiKey, exists := apikey.GetAPIKey(c)
        if exists {
            key = fmt.Sprintf("key:%d", apiKey.ID)
        } else {
            key = fmt.Sprintf("ip:%s", c.ClientIP())
        }
        
        // 获取服务信息
        service, exists := c.Get("service_info")
        if !exists {
            c.Next()
            return
        }
        
        serviceInfo := service.(*ServiceInfo)
        
        // 根据服务配置调整限流器
        serviceLimiter := rate.NewLimiter(
            rate.Limit(serviceInfo.Config.RateLimit/60), // 每秒请求数
            serviceInfo.Config.RateLimit/10,            // 突发请求数
        )
        
        // 检查限流
        if !serviceLimiter.Allow() {
            c.JSON(http.StatusTooManyRequests, model.NewErrorResponse(
                model.CodeRateLimitExceeded,
                "请求过于频繁，请稍后再试",
            ))
            c.Abort()
            return
        }
        
        c.Next()
    }
}
```

#### 4.3.3 配额中间件

```go
// quotaMiddleware 配额中间件
func (r *ProviderRouter) quotaMiddleware() gin.HandlerFunc {
    return func(c *gin.Context) {
        // 获取服务名称
        serviceName := c.Param("service")
        
        // 获取用户ID
        userID, exists := jwt.GetUserID(c)
        if !exists {
            apiKeyUserID, exists := apikey.GetAPIKeyUserID(c)
            if exists {
                userID = apiKeyUserID
            } else {
                // 匿名访问，不检查配额
                c.Next()
                return
            }
        }
        
        // 获取服务信息
        service, exists := c.Get("service_info")
        if !exists {
            c.Next()
            return
        }
        
        serviceInfo := service.(*ServiceInfo)
        
        // 检查配额
        quota, err := r.store.Quotas().GetByUserAndService(c.Request.Context(), userID, serviceName, "daily")
        if err != nil {
            // 配额不存在，创建默认配额
            quota = &model.ServiceQuota{
                UserID:     userID,
                ServiceName: serviceName,
                TimeWindow: "daily",
                Usage:      0,
                LimitValue: serviceInfo.Definition.DefaultLimit,
                ResetTime:  time.Now().Add(24 * time.Hour),
            }
            if err := r.store.Quotas().Create(c.Request.Context(), quota); err != nil {
                c.JSON(http.StatusInternalServerError, model.NewErrorResponse(
                    model.CodeInternalError,
                    "创建配额失败",
                ))
                c.Abort()
                return
            }
        }
        
        // 检查是否超出配额
        if quota.Usage >= quota.LimitValue {
            c.JSON(http.StatusForbidden, model.NewErrorResponse(
                model.CodeQuotaExceeded,
                "已超出服务配额限制",
            ))
            c.Abort()
            return
        }
        
        // 存储配额信息，以便后续更新
        c.Set("service_quota", quota)
        c.Set("service_cost", serviceInfo.Config.QuotaCost)
        
        c.Next()
    }
}
```

### 4.4 服务处理

服务处理函数负责执行具体的服务逻辑。

```go
// executeServiceHandler 执行服务处理函数
func (r *ProviderRouter) executeServiceHandler(c *gin.Context) {
    // 获取服务信息
    service, exists := c.Get("service_info")
    if !exists {
        c.JSON(http.StatusInternalServerError, model.NewErrorResponse(
            model.CodeInternalError,
            "服务信息不存在",
        ))
        return
    }
    
    serviceInfo := service.(*ServiceInfo)
    
    // 执行服务处理函数
    result, err := serviceInfo.Handler(c)
    if err != nil {
        c.JSON(http.StatusBadRequest, model.NewErrorResponse(
            model.CodeInvalidParams,
            err.Error(),
        ))
        return
    }
    
    // 更新配额使用情况
    quota, exists := c.Get("service_quota")
    if exists {
        cost, _ := c.Get("service_cost")
        serviceCost, _ := cost.(int)
        
        serviceQuota := quota.(*model.ServiceQuota)
        go func() {
            ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
            defer cancel()
            
            r.store.Quotas().IncrementUsage(ctx, serviceQuota.UserID, serviceQuota.ServiceName, serviceQuota.TimeWindow, serviceCost)
        }()
    }
    
    // 返回结果
    c.JSON(http.StatusOK, model.NewSuccessResponse(result))
}
```

## 5. 示例服务实现

### 5.1 Echo服务

```go
// echoServiceHandler Echo服务处理函数
func echoServiceHandler(c *gin.Context) (interface{}, error) {
    var request struct {
        Message string `json:"message" binding:"required"`
    }
    
    if err := c.ShouldBindJSON(&request); err != nil {
        return nil, fmt.Errorf("无效的请求参数: %w", err)
    }
    
    return gin.H{
        "message": request.Message,
        "timestamp": time.Now().Unix(),
    }, nil
}
```

### 5.2 时间服务

```go
// timeServiceHandler 时间服务处理函数
func timeServiceHandler(c *gin.Context) (interface{}, error) {
    now := time.Now()
    
    return gin.H{
        "timestamp": now.Unix(),
        "iso8601": now.Format(time.RFC3339),
        "date": now.Format("2006-01-02"),
        "time": now.Format("15:04:05"),
        "timezone": now.Location().String(),
    }, nil
}
```

## 6. 服务注册示例

```go
// RegisterServices 注册所有服务
func (r *ServiceRegistry) RegisterServices() error {
    // 注册Echo服务
    if err := r.RegisterService("echo", echoServiceHandler, ServiceConfig{
        AllowAnonymous: true,
        RateLimit:      60,  // 每分钟60次
        QuotaCost:      1,   // 消耗1个配额
        Description:    "回显服务，返回请求的内容",
        RequestExample: map[string]interface{}{
            "message": "Hello, APIHub!",
        },
        ResponseExample: map[string]interface{}{
            "message": "Hello, APIHub!",
            "timestamp": 1625097600,
        },
    }); err != nil {
        return err
    }
    
    // 注册时间服务
    if err := r.RegisterService("time", timeServiceHandler, ServiceConfig{
        AllowAnonymous: true,
        RateLimit:      60,  // 每分钟60次
        QuotaCost:      1,   // 消耗1个配额
        Description:    "时间服务，返回当前服务器时间",
        RequestExample: map[string]interface{}{},
        ResponseExample: map[string]interface{}{
            "timestamp": 1625097600,
            "iso8601": "2021-07-01T00:00:00Z",
            "date": "2021-07-01",
            "time": "00:00:00",
            "timezone": "UTC",
        },
    }); err != nil {
        return err
    }
    
    return nil
}
```

## 7. 目录结构设计

```
internal/provider/
├── registry/
│   ├── registry.go        # 服务注册中心
│   └── middleware.go      # 服务相关中间件
├── router.go              # 功能API路由器
├── handler.go             # 通用处理函数
└── services/
    ├── echo_service.go    # Echo服务实现
    └── time_service.go    # 时间服务实现
```

## 8. 实现步骤

1. 创建服务注册中心(Registry)
2. 实现服务路由机制
3. 集成现有的认证、限流和配额系统
4. 实现示例服务(Echo和Time)
5. 注册服务路由

## 9. 总结与展望

本设计方案保持了简单性，同时满足了APIHub项目的核心需求。它利用了现有的认证、配额和日志系统，避免了重复实现。服务注册采用代码注册的方式，而不是插件式的动态加载，这样可以减少复杂度并提高可维护性。

未来可以考虑的扩展：

1. **服务版本控制**：支持同一服务的多个版本并存
2. **服务文档生成**：自动生成API文档
3. **服务健康检查**：监控服务状态和性能
4. **服务依赖管理**：处理服务之间的依赖关系
5. **服务缓存机制**：缓存常用服务的响应结果

---

*文档更新日期：2024年7月* 