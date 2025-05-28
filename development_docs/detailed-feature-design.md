# UniAPI 详细功能设计文档

## 1. 功能模块详细设计

### 1.1 系统初始化模块
#### 功能描述
- 系统启动时的一次性初始化检查
- 管理员账号创建
- 基础配置项设置

#### 实现方案
1. **启动时检查机制**：
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

func (sb *SystemBootstrap) isInitialized() (bool, error) {
    var count int64
    // 检查users表是否存在admin用户
    err := sb.db.Model(&User{}).Where("role = ?", "admin").Count(&count).Error
    return count > 0, err
}

func (sb *SystemBootstrap) performInitialization() error {
    return sb.db.Transaction(func(tx *gorm.DB) error {
        // 1. 创建系统配置
        configs := []SystemConfig{
            {ConfigKey: "register_enabled", ConfigValue: "false"},
            {ConfigKey: "rate_limit_global", ConfigValue: "100"},
            {ConfigKey: "rate_limit_per_user", ConfigValue: "5"},
        }
        if err := tx.Create(&configs).Error; err != nil {
            return err
        }

        // 2. 等待管理员账号创建（通过API）
        return nil
    })
}
```

2. **初始化API**：
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

3. **数据库迁移管理**：
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

#### 优化设计
1. **初始化状态缓存**：
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

2. **初始化流程优化**：
- 使用数据库事务确保原子性
- 采用版本化的数据库迁移
- 提供初始化状态的缓存机制
- 支持配置导入导出

3. **安全性考虑**：
- 初始化API仅在未初始化状态可用
- 管理员密码必须符合复杂度要求
- 初始化操作记录详细日志
- 支持初始化超时机制

#### 实现优势
1. **性能优化**：
   - 启动时一次性检查，避免运行时重复检查
   - 使用内存缓存记录初始化状态
   - 数据库操作采用事务处理

2. **可维护性**：
   - 清晰的模块化设计
   - 版本化的数据库迁移
   - 完整的错误处理
   - 详细的日志记录

3. **安全性**：
   - 严格的初始化状态控制
   - 安全的密码处理
   - 事务性的数据操作
   - 访问控制机制

4. **可扩展性**：
   - 支持自定义初始化配置
   - 可扩展的迁移机制
   - 模块化的设计结构

#### 注意事项
1. 确保数据库连接配置正确
2. 初始化API需要适当的访问控制
3. 密码必须在传输和存储时加密
4. 保留详细的初始化日志
5. 考虑多实例部署场景

### 1.2 认证与授权模块
#### 功能描述
- 双重认证系统：Dashboard JWT认证 + API Key认证
- 基于角色的权限控制
- API Key的生命周期管理

#### 实现细节
1. Dashboard认证流程：
```go
// JWT认证中间件示例
func JWTAuthMiddleware() gin.HandlerFunc {
    return func(c *gin.Context) {
        // 1. 获取Token
        // 2. 验证Token
        // 3. 解析用户信息
        // 4. 权限检查
    }
}
```

2. API Key认证：
```sql
-- API Key验证查询
SELECT ak.*, u.status as user_status, u.role 
FROM api_keys ak 
JOIN users u ON ak.user_id = u.id 
WHERE ak.api_key = ? AND ak.status = 1 
  AND (ak.expires_at IS NULL OR ak.expires_at > CURRENT_TIMESTAMP);
```

#### 优化建议
- 实现API Key轮换机制
- 添加Token黑名单机制
- 考虑实现2FA双因素认证
- 关键操作添加操作日志

### 1.3 配额控制模块
#### 功能描述
- 用户级别的服务配额管理
- 多维度的使用量统计
- 自动重置机制

#### 实现细节
1. 配额检查流程：
```sql
-- 检查用户服务配额
SELECT sq.*, sd.default_limit 
FROM service_quotas sq
JOIN service_definitions sd ON sq.service_name = sd.service_name
WHERE sq.user_id = ? 
  AND sq.service_name = ?
  AND sq.time_window = ?;
```

2. 使用量统计：
```sql
-- 统计特定时间窗口的使用量
SELECT COUNT(*) as usage_count
FROM access_logs
WHERE user_id = ?
  AND service_name = ?
  AND created_at BETWEEN ? AND ?;
```

#### 优化建议
- 实现配额使用预警机制
- 添加分布式锁确保配额计数准确性
- 考虑使用Redis缓存热点配额数据
- 实现配额使用报告自动推送

### 1.4 服务管理模块
#### 功能描述
- 内部服务的注册与管理
- 服务状态监控
- 访问控制策略

#### 实现细节
1. 服务注册：
```sql
-- 注册新服务
INSERT INTO service_definitions 
(service_name, description, default_limit, status)
VALUES (?, ?, ?, 1);
```

2. 服务状态检查：
```go
func ServiceHealthCheck() {
    // 1. 检查服务可用性
    // 2. 更新服务状态
    // 3. 触发告警（如果需要）
}
```

#### 优化建议
- 实现服务健康检查机制
- 添加服务降级策略
- 实现服务调用链路追踪
- 考虑添加服务文档自动生成

### 1.5 监控统计模块
#### 功能描述
- 实时调用统计
- 多维度数据分析
- 性能监控

#### 实现细节
1. 访问日志记录：
```sql
-- 记录API调用
INSERT INTO access_logs 
(api_key_id, user_id, service_name, endpoint, status, cost)
VALUES (?, ?, ?, ?, ?, ?);
```

2. 统计分析：
```sql
-- 服务调用趋势分析
SELECT 
    date(created_at) as date,
    service_name,
    COUNT(*) as call_count,
    SUM(CASE WHEN status = 200 THEN 1 ELSE 0 END) as success_count
FROM access_logs
WHERE created_at BETWEEN ? AND ?
GROUP BY date, service_name;
```

#### 优化建议
- 实现日志异步写入
- 考虑使用时序数据库存储监控数据
- 添加自动化报告生成
- 实现智能告警机制

## 2. 数据库优化建议

### 2.1 性能优化
- 针对`access_logs`表实现分区策略
- 定期归档历史数据
- 为热点查询添加合适的索引
- 考虑使用缓存层减少数据库压力

### 2.2 可靠性优化
- 实现定期备份机制
- 添加数据完整性检查
- 实现故障自动恢复
- 考虑主从复制方案

### 2.3 安全性优化
- 实现敏感数据加密存储
- 添加操作审计日志
- 实现数据访问权限控制
- 定期安全扫描和评估

## 3. 开发注意事项

### 3.1 代码规范
- 遵循Go语言最佳实践
- 统一错误处理机制
- 规范化日志记录
- 完善的注释文档

### 3.2 测试规范
- 单元测试覆盖核心逻辑
- 集成测试验证功能模块
- 性能测试确保系统稳定
- 安全测试防范潜在风险

### 3.3 部署规范
- 环境配置文件管理
- 数据库迁移脚本
- 监控告警配置
- 备份恢复方案

## 4. 后续优化方向

### 4.1 功能扩展
- 支持更多认证方式
- 添加更细粒度的权限控制
- 实现更复杂的配额策略
- 支持服务编排和组合

### 4.2 性能提升
- 引入缓存机制
- 优化数据库查询
- 实现负载均衡
- 添加服务熔断机制

### 4.3 运维优化
- 自动化部署流程
- 完善监控体系
- 优化日志管理
- 增强系统可观测性 