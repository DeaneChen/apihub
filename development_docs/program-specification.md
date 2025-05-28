# UniAPI 项目规范说明

## 1. 设计目标

- **轻量级API管理平台**：低资源占用，易部署维护
- **性能指标**
  - 单用户并发：≤ 5 QPS
  - 总体并发：≤ 100 QPS
  - 每小时活跃用户：≤ 200
- **可扩展性**：支持模块化API服务架构，便捷接入新功能

## 2. 技术栈

- **Web框架**: Gin
- **数据库**: SQLite（轻量级选择）
- **缓存**: go-cache（内存缓存）
- **配置管理**: godotenv
- **日志系统**: 使用 Gin 内置的日志系统
- **认证系统**: golang-jwt

## 3. 核心功能特性

### 3.1 系统初始化
- **首次访问初始化**
  - 连接数据库是否初始化
  - 表单填写创建管理员账号（仅首次）
  - 初始化系统基础配置

### 3.2 管理员系统
- **系统配置管理**
  - 用户注册开关
  - 全局限流设置
  - API公开访问控制
  - 公开API限流策略
- **用户管理**
  - 用户账号管理
  - 用户权限设置
  - 用户配额调整

### 3.3 双重认证系统
- **Dashboard认证**
  - JWT基础的用户认证
  - 用于管理面板的访问控制
  - 用户权限管理

- **API认证**
  - 基于APIKey的认证机制
  - 支持多个APIKey管理
  - Key级别的访问控制
  - 公开API的匿名访问控制

### 3.4 用户系统
- 用户注册与管理（可控开关）
- 用户配额控制
- 使用量统计

### 3.5 APIKey管理
- Key的生成与管理
- 使用限制设置
- 状态管理（启用/禁用）

### 3.6 限流与配额
- 系统级别限流
- 用户级别限流
- APIKey级别限流
- 公开API限流
- 使用配额管理

### 3.7 API服务管理
- 模块化服务接入
- 服务状态监控
- 调用统计
- 公开/私有访问控制

### 3.8 监控统计
- 实时调用统计
- 用量报告
- 基础性能监控

## 4. 项目结构

```
uniapi/
├── cmd/                # 应用入口
├── config/            # 配置文件
├── internal/
│   ├── dashboard/    # 管理面板相关
│   │   ├── handler/  # 面板API处理器
│   │   ├── model/   # 面板数据模型
│   │   └── service/ # 面板业务逻辑
│   ├── core/        # 核心功能模块
│   │   ├── admin/   # 管理员功能
│   │   ├── apikey/  # APIKey管理
│   │   └── quota/   # 配额管理
│   ├── middleware/  # 中间件
│   └── provider/    # 功能性API提供者
├── pkg/             # 公共工具包
├── web/             # web应用程序
└── docs/           # 文档
```

## 5. API路由结构

### 5.1 Dashboard API
```
/dashboard
├── /admin
│   ├── POST /settings          # 更新系统设置
│   ├── GET  /settings          # 获取系统设置
│   └── POST /toggle-register   # 切换注册开关
├── /auth
│   ├── POST /login            # 用户登录
│   └── POST /logout           # 用户登出
├── /users
│   ├── GET  /list            # 获取用户列表
│   ├── GET  /profile         # 获取用户信息
│   ├── POST /create          # 创建用户
│   ├── POST /update          # 更新用户信息
│   ├── POST /quota           # 调整用户配额
│   └── POST /delete          # 删除用户
├── /apikeys
│   ├── GET  /list            # 获取APIKey列表
│   ├── POST /create          # 创建APIKey
│   ├── POST /update          # 更新APIKey
│   └── POST /delete          # 删除APIKey
└── /stats
    ├── GET  /overview        # 获取总体统计
    ├── GET  /users           # 获取用户统计
    └── GET  /apis            # 获取API使用统计

### 5.2 功能性 API
```
/api/v1
├── GET  /{service-name}/info      # 获取服务信息
├── POST /{service-name}/execute   # 执行服务
└── GET  /status                   # 服务状态检查
```

## 6. API 请求响应规范

### 6.1 统一响应格式
```json
{
    "code": 0,           // 状态码，0表示成功
    "message": "ok",     // 状态信息
    "data": {}          // 响应数据
}
```

### 6.2 分页请求参数
```
GET /users/list?page=1&size=10
```

### 6.3 批量操作
```
POST /users/delete
{
    "ids": [1, 2, 3]    // 批量操作的ID列表
}
```

### 6.4 查询过滤
```
GET /apikeys/list?status=active&created_after=2024-01-01
```

