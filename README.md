# APIHub

统一 API 业务服务框架，实现多种 功能性服务API 并集中管理。

## 项目结构

```
apihub/
├── api/                    # API 定义目录
│   ├── dashboard/         # Dashboard API 定义
│   └── provider/         # 功能性 API 定义
├── cmd/                    
│   └── apihub/           # 主程序入口
│       └── main.go
├── configs/               # 配置文件目录
│   ├── config.json.example
│   └── config.go
├── internal/              # 私有应用代码
│   ├── auth/             # 认证相关
│   │   ├── jwt/
│   │   └── apikey/
│   ├── dashboard/        # Dashboard 相关
│   │   ├── handler/     # HTTP handlers
│   │   ├── service/     # 业务逻辑
│   │   └── repository/  # 数据访问
│   ├── provider/        # 功能性 API 提供者
│   │   ├── registry/    # 服务注册中心
│   │   └── services/    # 具体服务实现
│   ├── router/          # 路由入口
│   ├── middleware/      # 中间件
│   ├── model/          # 数据模型
│   └── store/          # 数据存储层
├── pkg/                  # 可复用的公共包
│   └── utils/          # 通用工具
├── web/                 # Web 前端
│   ├── dashboard/      # 管理面板前端
│   └── public/         # 静态资源
├── scripts/             # 构建、部署脚本
├── test/               # 测试文件
│   ├── integration/    # 集成测试
│   └── mock/          # 测试模拟数据
├── docs/               # 项目文档
│   ├── api/           # API 文档
│   └── guides/        # 使用指南
├── go.mod
├── go.sum
└── README.md
```

## 开发环境要求

- Go 1.21 或更高版本

## 快速开始

1. 克隆项目
2. 安装依赖：`go mod download`
3. 生成配置文件：`go run cmd/apihub/main.go --generate-config`
4. 运行项目：`go run cmd/apihub/main.go`

## 配置与运行

APIHub 提供了灵活的配置管理系统，支持多种配置方式。

### 命令行参数

```
Usage of apihub:
  -config string
        配置文件路径
  -generate-config
        生成默认配置文件
```

### 配置文件生成

生成默认配置文件：

```bash
# 在当前目录生成 config.json
./apihub --generate-config

# 指定配置文件路径
./apihub --generate-config --config=/path/to/config.json
```

### 配置文件位置

APIHub 会按以下顺序查找配置文件：

1. 命令行参数 `--config` 指定的路径
2. 当前工作目录中的 `config.json`
3. 可执行文件所在目录中的 `config.json`
4. 系统配置目录：
   - Linux/Unix: `/etc/apihub/config.json`
   - macOS: `/Library/Application Support/apihub/config.json`
   - Windows: `%PROGRAMDATA%\apihub\config.json`
5. 用户配置目录：
   - Linux/Unix: `~/.config/apihub/config.json`
   - macOS: `~/Library/Application Support/apihub/config.json`

### 环境变量配置

所有配置项都可以通过环境变量覆盖，环境变量命名规则为 `APIHUB_` 前缀加上配置项路径，例如：

```bash
# 服务器配置
APIHUB_SERVER_PORT=8081
APIHUB_SERVER_HOST=127.0.0.1

# 数据库配置
APIHUB_DB_TYPE=sqlite
APIHUB_DB_DSN=apihub.db

# 认证配置
APIHUB_JWT_SECRET=your-jwt-secret
APIHUB_APIKEY_SECRET=your-apikey-secret

# 日志配置
APIHUB_LOG_LEVEL=debug
```

### 运行方式

#### 开发环境运行

```bash
# 使用默认配置
go run cmd/apihub/main.go

# 指定配置文件
go run cmd/apihub/main.go --config=configs/dev.json

# 使用环境变量
APIHUB_SERVER_PORT=8081 go run cmd/apihub/main.go
```

#### 生产环境运行

```bash
# 构建可执行文件
go build -o apihub cmd/apihub/main.go

# 生成配置文件
./apihub --generate-config --config=/etc/apihub/config.json

# 运行服务
./apihub --config=/etc/apihub/config.json
```

#### Docker 环境运行

```bash
# 构建镜像
docker build -t apihub .

# 使用环境变量运行
docker run -p 8080:8080 -e APIHUB_SERVER_PORT=8080 -e APIHUB_APIKEY_SECRET=your-secret apihub

# 挂载配置文件运行
docker run -p 8080:8080 -v /path/to/config.json:/app/config.json apihub
```

### API 认证

APIHub 支持两种认证方式：

1. **JWT 认证**：用于 Dashboard API
2. **API Key 认证**：用于功能性 API

#### API Key 认证方式

API Key 可以通过以下三种方式在请求中提供：

1. 通过 `X-API-Key` 头部：
   ```
   X-API-Key: your-api-key-here
   ```

2. 通过 `Authorization` 头部（Bearer 格式）：
   ```
   Authorization: Bearer your-api-key-here
   ```

3. 通过 URL 查询参数：
   ```
   https://your-api-domain/api/v1/provider/service_name/execute?api_key=your-api-key-here
   ```

## 系统初始化

首次运行时，系统会自动初始化：

1. 创建并初始化数据库
2. 生成随机的 JWT 密钥和 API Key 密钥（存储在数据库中）
3. 创建默认管理员账号

默认管理员账号信息会在首次启动时输出到日志中，请注意保存并及时修改密码。 