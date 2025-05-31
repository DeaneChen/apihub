# APIHub 认证与授权模块设计

## 1. 开发背景

**开发阶段**: 第三阶段 - 认证授权系统实现
**主要目标**: 实现JWT和APIKey双重认证系统，完成权限控制框架

## 2. 当前项目进展分析

通过分析项目文档和代码，项目当前进展如下：

1. **已完成部分**:
   - 基本项目结构搭建
   - 存储层架构(Store接口和SQLite实现)
   - 数据模型定义(User, APIKey, Config等)
   - 数据库初始化和迁移
   - 系统初始化功能

2. **正在进行**:
   - 认证与授权系统的设计与实现
   - 用户管理API的开发

3. **未开始部分**:
   - 服务管理系统
   - 配额控制实现
   - 前端开发
   - 测试与优化

## 3. 认证与授权模块设计

### 3.1 模块架构

```
internal/
  ├── auth/
  │   ├── jwt/
  │   │   ├── jwt.go           // JWT工具函数
  │   │   ├── middleware.go    // JWT中间件
  │   │   └── claims.go        // JWT Claims定义
  │   ├── apikey/
  │   │   ├── apikey.go        // APIKey工具函数
  │   │   └── middleware.go    // APIKey中间件
  │   ├── crypto/
  │   │   └── crypto.go        // 加密工具函数
  │   └── permission/
  │       ├── permission.go    // 权限检查逻辑
  │       └── middleware.go    // 权限中间件
  ├── middleware/
  │   └── auth.go              // 认证中间件组合
```

### 3.2 JWT认证系统

#### 3.2.1 JWT Token结构

```go
// internal/auth/jwt/claims.go
type CustomClaims struct {
    UserID    uint   `json:"user_id"`
    Username  string `json:"username"`
    Role      string `json:"role"`
    TokenType string `json:"token_type"` // "access" 或 "refresh"
    jwt.RegisteredClaims
}
```

#### 3.2.2 JWT工具函数

```go
// internal/auth/jwt/jwt.go
type JWTService struct {
    secretKey     []byte
    accessExpiry  time.Duration
    refreshExpiry time.Duration
    store         store.Store
}

// 主要功能
// - 生成访问令牌和刷新令牌
// - 验证令牌有效性
// - 刷新访问令牌
// - 令牌黑名单管理
```

#### 3.2.3 JWT中间件

```go
// JWT认证中间件
// - 从请求头获取Token
// - 验证Token有效性
// - 检查Token类型和黑名单
// - 将用户信息存入上下文
```

### 3.3 APIKey认证系统

#### 3.3.1 APIKey管理

APIKey将采用可逆加密存储在数据库中，确保可以随时向用户展示其完整的APIKey。主要功能包括：

- **APIKey生成**：生成随机字符串作为APIKey
- **APIKey存储**：使用AES等对称加密算法加密后存储
- **APIKey验证**：解密后进行比对验证
- **APIKey查询**：用户可查看自己的完整APIKey

#### 3.3.2 加密层设计

```go
// 加密服务接口
type CryptoService interface {
    Encrypt(plaintext string) (string, error)
    Decrypt(ciphertext string) (string, error)
}

// AES实现
type AESCryptoService struct {
    key []byte
}
```

在存储层中添加加密处理：

```go
// 在Repository层处理加密逻辑，保持数据模型和业务逻辑不变
func (r *SQLiteAPIKeyRepository) Create(apiKey *model.APIKey) error {
    // 加密APIKey
    encryptedKey, err := r.cryptoService.Encrypt(apiKey.APIKey)
    if err != nil {
        return err
    }
    
    // 存储加密后的APIKey
    // ...
}

func (r *SQLiteAPIKeyRepository) GetByKey(key string) (*model.APIKey, error) {
    // 先加密输入的key
    encryptedKey, err := r.cryptoService.Encrypt(key)
    if err != nil {
        return nil, err
    }
    
    // 直接使用加密后的值作为查询条件
    
    // 执行查询并返回结果
    // ...
    
    // 用户查看时，解密返回
    apiKey.APIKey, _ = r.cryptoService.Decrypt(encryptedKey)
    return apiKey, nil
}
```

#### 3.3.3 APIKey中间件

```go
// APIKey认证中间件
// - 从请求获取APIKey
// - 验证APIKey有效性
// - 检查APIKey状态和过期时间
// - 将用户和APIKey信息存入上下文
```

### 3.4 权限控制系统

#### 3.4.1 权限定义

```go
// 权限常量定义
const (
    // 用户相关权限
    PermCreateUser   = "user:create"
    PermReadUser     = "user:read"
    // ... 其他权限 ...
)

// 角色权限映射
var RolePermissions = map[string][]string{
    "admin": { /* 管理员权限 */ },
    "user":  { /* 普通用户权限 */ },
}
```

#### 3.4.2 权限中间件

```go
// 权限检查中间件
// - 从上下文获取用户角色
// - 检查角色是否具有所需权限
// - 授权或拒绝访问
```

### 3.5 认证路由定义

```go
// internal/api/dashboard/auth/handler.go
func (h *AuthHandler) RegisterRoutes(router *gin.RouterGroup) {
    authGroup := router.Group("/auth")
    {
        authGroup.POST("/login", h.Login)
        authGroup.POST("/refresh", h.RefreshToken)
        authGroup.POST("/logout", jwt.JWTAuthMiddleware(h.jwtService), h.Logout)
    }
}

// 登录处理
func (h *AuthHandler) Login(c *gin.Context) {
    // 1. 验证登录信息
    // 2. 生成JWT令牌
    // 3. 返回访问令牌和刷新令牌
}

// 刷新令牌处理
func (h *AuthHandler) RefreshToken(c *gin.Context) {
    // 1. 验证刷新令牌
    // 2. 生成新的访问令牌
    // 3. 返回新的访问令牌
}

// 登出处理
func (h *AuthHandler) Logout(c *gin.Context) {
    // 1. 获取当前令牌
    // 2. 将令牌加入黑名单
    // 3. 返回成功消息
}
```

## 4. 认证存储层设计

### 4.1 Token黑名单

使用内存缓存和数据库结合的方式实现令牌黑名单，提高验证效率：

- 内存缓存：快速查询，减少数据库访问
- 数据库存储：持久化存储，系统重启后恢复

### 4.2 存储接口扩展

为Store接口添加TokenBlacklist仓库支持：

```go
type TokenBlacklistRepository interface {
    Create(token *model.TokenBlacklist) error
    GetByToken(token string) (*model.TokenBlacklist, error)
    DeleteExpired() error
}

// Store接口扩展
type Store interface {
    // ... 现有方法 ...
    TokenBlacklist() TokenBlacklistRepository
}
```

### 4.3 SQLite实现

```go
// internal/store/sqlite/token_blacklist_repository.go
type SQLiteTokenBlacklistRepository struct {
    db *sql.DB
}

func (r *SQLiteTokenBlacklistRepository) Create(token *model.TokenBlacklist) error {
    query := `INSERT INTO token_blacklist (token, expires_at) VALUES (?, ?)`
    _, err := r.db.Exec(query, token.Token, token.ExpiresAt)
    return err
}

func (r *SQLiteTokenBlacklistRepository) GetByToken(token string) (*model.TokenBlacklist, error) {
    query := `SELECT token, expires_at FROM token_blacklist WHERE token = ?`
    row := r.db.QueryRow(query, token)
    
    var t model.TokenBlacklist
    err := row.Scan(&t.Token, &t.ExpiresAt)
    if err != nil {
        return nil, err
    }
    
    return &t, nil
}

func (r *SQLiteTokenBlacklistRepository) DeleteExpired() error {
    query := `DELETE FROM token_blacklist WHERE expires_at < ?`
    _, err := r.db.Exec(query, time.Now())
    return err
}
```

## 5. 安全性考虑

### 5.1 密码安全
- 使用bcrypt进行密码加密存储
- 登录时密码验证使用时间恒定比较避免计时攻击
- 密码复杂度要求：最少8位，包含大小写字母、数字和特殊字符

### 5.2 JWT安全
- 使用非对称加密算法(RSA)生成JWT密钥对
- 设置合理的Token过期时间(访问令牌15分钟，刷新令牌7天)
- 关键操作验证JWT时需重新验证用户密码
- 令牌轮换和黑名单机制

### 5.3 APIKey安全
- 生成高熵随机APIKey(32字符)
- 支持设置APIKey过期时间
- 使用AES等对称加密算法加密存储APIKey
- 提供APIKey轮换机制

## 6. 实现步骤

1. **数据库表扩展**:
   - 添加`token_blacklist`表
   - 更新`api_keys`表，添加使用范围字段

2. **核心功能实现**:
   - 实现JWT服务和中间件
   - 实现APIKey服务和中间件
   - 实现加密服务
   - 实现权限控制系统
   - 实现认证API接口

3. **安全加固**:
   - 实现令牌黑名单
   - 完善密码安全机制
   - 添加关键操作日志

4. **优化与测试**:
   - 完成单元测试
   - 添加集成测试
   - 性能测试和优化

## 7. 下一步工作

完成认证授权系统后，项目将进入第四阶段：用户管理API开发，包括：

1. **用户认证接口**:
   - `POST /auth/login`: 用户登录
   - `POST /auth/logout`: 用户登出
   - `POST /auth/refresh`: Token刷新
   - `GET /auth/profile`: 获取用户信息

2. **用户管理接口**:
   - `GET /users`: 用户列表（管理员）
   - `POST /users`: 创建用户（管理员）
   - `PUT /users/:id`: 更新用户信息
   - `DELETE /users/:id`: 删除用户（管理员）
   - `PUT /users/:id/password`: 修改密码

3. **API密钥管理**:
   - `GET /api-keys`: 获取用户API密钥列表
   - `POST /api-keys`: 创建新API密钥
   - `PUT /api-keys/:id`: 更新API密钥
   - `DELETE /api-keys/:id`: 删除API密钥

## 8. 总结

本设计文档详细规划了APIHub项目第三阶段的认证授权系统实现。通过双重认证机制(JWT+APIKey)，结合细粒度的权限控制，为系统提供全面的安全保障。同时考虑了性能、可扩展性和安全性，使系统能够满足不同场景的认证需求。 