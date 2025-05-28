# UniAPI

统一 API 业务服务框架，实现多种 功能性服务API 并集中管理。

## 项目结构

```
.
├── api/            # API 协议定义（OpenAPI/Swagger 规范、protobuf 文件等）
├── cmd/            # 主要的应用程序入口
│   └── main.go     # 主程序入口文件
├── config/         # 配置文件目录
├── docs/           # 项目文档
├── internal/       # 私有应用程序代码
│   ├── api/        # HTTP API 处理层
│   ├── model/      # 数据模型定义
│   ├── service/    # 业务逻辑层
│   ├── store/      # 数据存储层
│   └── middleware/ # 中间件
├── pkg/            # 可以被外部应用程序使用的库代码
├── web/             # web应用程序
├── scripts/        # 各类脚本
└── test/           # 测试相关文件
```

## 开发环境要求

- Go 1.21 或更高版本

## 快速开始

1. 克隆项目
2. 安装依赖：`go mod download`
3. 运行项目：`go run cmd/main.go` 