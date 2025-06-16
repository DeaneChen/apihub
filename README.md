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
│   ├── config.yaml.example
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
3. 运行项目：`go run cmd/main.go` 