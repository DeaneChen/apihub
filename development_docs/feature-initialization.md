# 系统初始化模块设计

## 1. 功能概述

### 1.1 模块职责
- 系统首次启动时的初始化检查
- 数据库结构初始化
- 管理员账号创建
- 基础配置项设置

### 1.2 设计目标
- 确保系统只初始化一次
- 提供安全的初始化流程
- 支持多实例部署场景
- 保证数据一致性

## 2. 详细设计

### 2.1 启动时检查机制
```go
// internal/core/bootstrap/init.go
type SystemBootstrap struct {
    db *gorm.DB
    logger *zap.Logger
}

func (sb *SystemBootstrap) Initialize() error {
    // 1. 检查数据库连接
    if err := sb.checkDatabase(); err != nil {
        return fmt.Errorf("database check failed: %w", err)
    }

    // 2. 检查数据库版本和结构
    if err := sb.checkSchema(); err != nil {
        return fmt.Errorf("schema check failed: %w", err)
    }

    // 3. 检查是否需要初始化
    initialized, err := sb.isInitialized()
    if err != nil {
        return fmt.Errorf("initialization check failed: %w", err)
    }

    if !initialized {
        if err := sb.performInitialization(); err != nil {
            return fmt.Errorf("initialization failed: %w", err)
        }
    }

    return nil
}
```

### 2.2 初始化状态管理
```go
type InitializationCache struct {
    initialized bool
    mu         sync.RWMutex
}

func (ic *InitializationCache) IsInitialized() bool {
    ic.mu.RLock()
    defer ic.mu.RUnlock()
    return ic.initialized
}

func (ic *InitializationCache) SetInitialized() {
    ic.mu.Lock()
    defer ic.mu.Unlock()
    ic.initialized = true
}
```

### 2.3 初始化API接口
```go
// internal/api/admin/init.go
func (h *AdminHandler) InitializeSystem(c *gin.Context) {
    // 仅在系统未初始化时可用
    if initialized, _ := h.svc.IsInitialized(); initialized {
        c.JSON(http.StatusForbidden, gin.H{
            "error": "system already initialized",
        })
        return
    }

    var req InitRequest
    if err := c.ShouldBindJSON(&req); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
        return
    }

    // 创建管理员账号
    admin := &User{
        Username: req.Username,
        Password: req.Password, // 注意：实际使用时需要加密
        Role:     "admin",
        Status:   1,
    }

    if err := h.svc.CreateAdmin(admin); err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
        return
    }

    c.JSON(http.StatusOK, gin.H{"message": "system initialized successfully"})
}
```

### 2.4 数据库迁移管理
```go
// internal/core/bootstrap/migration.go
func (sb *SystemBootstrap) checkSchema() error {
    // 使用版本化的迁移文件
    return sb.db.AutoMigrate(
        &User{},
        &APIKey{},
        &SystemConfig{},
        &AccessLog{},
        &ServiceQuota{},
        &ServiceDefinition{},
    )
}
```

## 3. 优化设计

### 3.1 性能优化
- 启动时一次性检查，避免运行时重复检查
- 使用内存缓存记录初始化状态
- 数据库操作采用事务处理

### 3.2 可维护性优化
- 清晰的模块化设计
- 版本化的数据库迁移
- 完整的错误处理
- 详细的日志记录

### 3.3 安全性优化
- 初始化API仅在未初始化状态可用
- 管理员密码必须符合复杂度要求
- 初始化操作记录详细日志
- 支持初始化超时机制

### 3.4 可扩展性优化
- 支持自定义初始化配置
- 可扩展的迁移机制
- 模块化的设计结构

## 4. 注意事项

### 4.1 开发注意事项
1. 确保数据库连接配置正确
2. 初始化API需要适当的访问控制
3. 密码必须在传输和存储时加密
4. 保留详细的初始化日志

### 4.2 部署注意事项
1. 多实例部署时的初始化协调
2. 数据库迁移的版本控制
3. 配置文件的环境隔离
4. 日志的统一管理

### 4.3 安全注意事项
1. 初始化接口的访问控制
2. 管理员密码的安全要求
3. 敏感信息的加密存储
4. 操作日志的安全存储 