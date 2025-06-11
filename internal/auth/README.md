# APIHub 认证与授权系统

本文档介绍APIHub认证与授权系统的使用方法。

## 系统架构

认证与授权系统包含以下组件：

- **JWT认证**: 基于JSON Web Token的用户认证
- **APIKey认证**: 基于API密钥的服务认证
- **权限控制**: 基于角色的权限管理
- **缓存系统**: 使用go-cache的Token黑名单缓存
- **加密服务**: APIKey的AES加密存储

## 快速开始

### 1. 初始化认证服务

```go
package main

import (
    "apihub/internal/auth"
    "apihub/internal/store"
)

func main() {
    // 创建存储实例
    store := store.NewSQLiteStore("database.db")
    
    // 使用默认配置创建认证服务
    config := auth.DefaultAuthConfig()
    authServices, err := auth.NewAuthServices(config, store)
    if err != nil {
        panic(err)
    }
    
    // 现在可以使用认证服务了
    jwtService := authServices.JWTService
    apiKeyService := authServices.APIKeyService
    permissionService := authServices.PermissionService
}
```

### 2. JWT认证使用

#### 生成Token

```go
// 用户登录成功后生成Token
user := &model.User{
    ID:       1,
    Username: "admin",
    Role:     "admin",
}

tokenResponse, err := jwtService.GenerateToken(user)
if err != nil {
    // 处理错误
}

// 返回给客户端
response := map[string]interface{}{
    "access_token": tokenResponse.AccessToken,
    "expires_in":   tokenResponse.ExpiresIn,
}
```

#### 验证Token

```go
// 验证访问令牌
claims, err := jwtService.ValidateToken(accessToken)
if err != nil {
    // Token无效
}

// 获取用户信息
userID := claims.UserID
username := claims.Username
role := claims.Role
```

### 3. APIKey认证使用

#### 创建APIKey

```go
// 为用户创建APIKey
apiKey, err := apiKeyService.CreateAPIKey(
    userID,
    "My API Key",
    "用于访问API的密钥",
    nil, // 不设置过期时间
    []string{"service:use", "user:read"}, // 权限范围
)
if err != nil {
    // 处理错误
}

// 返回APIKey（仅此一次显示完整密钥）
fmt.Println("API Key:", apiKey.APIKey)
```

#### 验证APIKey

```go
// 验证APIKey
apiKey, err := apiKeyService.ValidateAPIKey(keyString)
if err != nil {
    // APIKey无效
}

// 检查权限范围
if apiKeyService.CheckAPIKeyScope(apiKey, "service:use") {
    // 有权限使用服务
}
```

### 4. 中间件使用

#### 在Gin路由中使用

```go
package main

import (
    "apihub/internal/auth"
    "apihub/internal/auth/permission"
    "apihub/internal/middleware"
    
    "github.com/gin-gonic/gin"
)

func setupRoutes(authServices *auth.AuthServices) *gin.Engine {
    r := gin.Default()
    
    // 公开路由
    public := r.Group("/api/v1")
    {
        public.POST("/auth/login", loginHandler)
        public.GET("/services", listServicesHandler)
    }
    
    // 需要认证的路由（支持JWT和APIKey）
    protected := r.Group("/api/v1")
    protected.Use(middleware.AuthMiddleware(
        authServices.JWTService,
        authServices.APIKeyService,
    ))
    {
        protected.GET("/profile", getProfileHandler)
        protected.GET("/api-keys", listAPIKeysHandler)
    }
    
    // 仅JWT认证的路由
    jwtOnly := r.Group("/api/v1")
    jwtOnly.Use(middleware.JWTOnlyMiddleware(authServices.JWTService))
    {
        jwtOnly.POST("/auth/logout", logoutHandler)
    }
    
    // 需要特定权限的路由
    admin := r.Group("/api/v1/admin")
    admin.Use(middleware.JWTOnlyMiddleware(authServices.JWTService))
    admin.Use(permission.AdminOnlyMiddleware())
    {
        admin.GET("/users", listUsersHandler)
        admin.POST("/users", createUserHandler)
    }
    
    return r
}
```

#### 权限检查中间件

```go
// 要求特定权限
r.GET("/users/:id", 
    middleware.JWTOnlyMiddleware(jwtService),
    permission.RequirePermissionMiddleware(permissionService, permission.PermReadUser),
    getUserHandler,
)

// 要求资源访问权限（用户只能访问自己的资源）
r.GET("/users/:id/api-keys",
    middleware.JWTOnlyMiddleware(jwtService),
    permission.RequireResourceAccessMiddleware(
        permissionService, 
        permission.PermListAPIKeys, 
        "id", // 路径参数名
    ),
    listUserAPIKeysHandler,
)
```

### 5. 在处理器中获取用户信息

```go
func getProfileHandler(c *gin.Context) {
    // 获取当前用户ID（支持JWT和APIKey）
    userID, exists := middleware.GetCurrentUserID(c)
    if !exists {
        c.JSON(400, gin.H{"error": "user not found"})
        return
    }
    
    // 检查认证方式
    if middleware.IsJWTAuth(c) {
        // JWT认证，可以获取更多信息
        username, _ := middleware.GetCurrentUsername(c)
        role, _ := middleware.GetCurrentUserRole(c)
        
        c.JSON(200, gin.H{
            "user_id":  userID,
            "username": username,
            "role":     role,
            "auth_type": "jwt",
        })
    } else if middleware.IsAPIKeyAuth(c) {
        // APIKey认证
        c.JSON(200, gin.H{
            "user_id":   userID,
            "auth_type": "apikey",
        })
    }
}
```

## 权限系统

### 预定义权限

系统预定义了以下权限：

- 用户相关: `user:create`, `user:read`, `user:update`, `user:delete`, `user:list`
- API密钥: `apikey:create`, `apikey:read`, `apikey:update`, `apikey:delete`, `apikey:list`
- 服务相关: `service:create`, `service:read`, `service:update`, `service:delete`, `service:list`, `service:use`
- 配额相关: `quota:create`, `quota:read`, `quota:update`, `quota:delete`, `quota:list`
- 系统配置: `config:create`, `config:read`, `config:update`, `config:delete`, `config:list`
- 访问日志: `accesslog:read`, `accesslog:list`
- 系统管理: `system:admin`, `system:read`

### 角色定义

- **admin**: 管理员，拥有所有权限
- **user**: 普通用户，可以管理自己的资源
- **guest**: 访客，只能查看公开信息

### 权限检查

```go
// 检查角色权限
hasPermission := permissionService.HasPermission("user", "service:use")

// 检查资源访问权限
canAccess := permissionService.CanAccessResource("user", 1, 1, "user:read")
```

## 安全考虑

1. **密钥管理**: 生产环境必须使用强密钥
2. **Token过期**: 合理设置Token过期时间
3. **APIKey加密**: APIKey使用AES加密存储
4. **权限最小化**: 遵循最小权限原则
5. **日志记录**: 记录认证和授权相关的操作日志

## 配置示例

```go
config := auth.AuthConfig{
    JWT: auth.JWTConfig{
        AccessExpiry: 24 * time.Hour, // 访问令牌24小时过期
        Issuer:       "apihub",
        // 生产环境应该提供RSA密钥对
        PrivateKeyPEM: "-----BEGIN RSA PRIVATE KEY-----\n...",
        PublicKeyPEM:  "-----BEGIN PUBLIC KEY-----\n...",
    },
    Crypto: auth.CryptoConfig{
        SecretKey: "your-secret-key-32-characters-long",
    },
    Cache: auth.CacheConfig{
        DefaultExpiration: 30 * time.Minute,
        CleanupInterval:   10 * time.Minute,
    },
}
``` 