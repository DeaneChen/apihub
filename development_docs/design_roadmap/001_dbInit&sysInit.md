### 数据库初始化模块设计

#### 1. 目录结构
```go
internal/
  ├── store/
  │   ├── sqlite/
  │   │   ├── sqlite.go       // SQLite连接管理
  │   │   └── migrations/     // 数据库迁移文件
  │   └── store.go           // 存储接口定义
  ├── model/
  │   ├── user.go            // 用户模型
  │   ├── apikey.go          // API密钥模型
  │   ├── config.go          // 系统配置模型
  │   ├── quota.go           // 服务配额模型
  │   └── service.go         // 服务定义模型
```

#### 2. 核心接口设计
```go
// store.go
type Store interface {
    // 数据库连接管理
    Connect() error
    Close() error
    
    // 数据库迁移
    Migrate() error
    
    // 事务管理
    BeginTx() (Transaction, error)
    
    // 各个模型的CRUD接口
    Users() UserRepository
    APIKeys() APIKeyRepository
    Configs() ConfigRepository
    Quotas() QuotaRepository
    Services() ServiceRepository
}

// 事务接口
type Transaction interface {
    Commit() error
    Rollback() error
}
```

#### 3. 初始化流程设计
1. 数据库连接建立
2. 数据库迁移执行
3. 系统初始配置检查
4. 默认管理员账号创建（如果不存在）

#### 4. 错误处理设计
```go
type DBError struct {
    Code    int
    Message string
    Err     error
}

const (
    ErrConnectionFailed = iota + 1000
    ErrMigrationFailed
    ErrDataConstraint
    ErrNotFound
)
```

#### 5. 配置管理设计
```yaml
database:
  driver: sqlite3
  dsn: apihub.db
  max_open_conns: 10
  max_idle_conns: 5
  conn_max_lifetime: 1h
```

#### 6. 单元测试规划
- 连接测试
- 迁移测试
- CRUD操作测试
- 事务测试
- 错误处理测试

#### 7. 实现步骤
1. 实现基础的数据库连接管理
2. 创建数据库迁移文件
3. 实现各个模型的基础结构
4. 实现存储层接口
5. 添加错误处理
6. 编写单元测试
7. 实现系统初始化逻辑

这个设计考虑了以下几个方面：
- 模块化和接口分离
- 完整的错误处理
- 事务支持
- 可测试性
- 配置灵活性

你觉得这个设计怎么样？我们可以开始实现数据库初始化模块了。需要我详细说明某个具体部分吗？