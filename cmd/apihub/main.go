package main

import (
	"apihub/internal/core"
	"context"
	"log"
	"time"

	"apihub/internal/auth"
	"apihub/internal/provider"
	"apihub/internal/provider/registry"
	"apihub/internal/router"
	"apihub/internal/store/sqlite"

	// 导入 Swagger 文档
	_ "apihub/docs"

	"github.com/gin-gonic/gin"
)

func main() {
	// 设置Gin模式
	gin.SetMode(gin.ReleaseMode)

	// 创建数据库连接
	store := sqlite.NewSQLiteStore("apihub.db")

	// 创建初始化服务
	initService := core.NewInitializationService(store)

	// 执行系统初始化
	ctx := context.Background()
	if err := initService.InitializeSystem(ctx); err != nil {
		log.Fatalf("系统初始化失败: %v", err)
	}

	// 创建认证服务配置
	authConfig := auth.AuthConfig{
		JWT: auth.JWTConfig{
			AccessExpiry: 24 * time.Hour, // 访问令牌24小时过期
			Issuer:       "apihub",
			// 生产环境应该从配置文件或环境变量读取密钥
			PrivateKeyPEM: "", // 留空将自动生成密钥对
			PublicKeyPEM:  "",
		},
		Crypto: auth.CryptoConfig{
			SecretKey: "apihub-secret-key-change-in-production-32chars", // 32字符密钥
		},
		Cache: auth.CacheConfig{
			DefaultExpiration: 30 * time.Minute, // 默认缓存30分钟
			CleanupInterval:   10 * time.Minute, // 每10分钟清理一次过期缓存
		},
	}

	// 确保数据库保持连接
	if err := store.Connect(); err != nil {
		log.Fatalf("数据库连接失败: %v", err)
	}

	// 创建认证服务
	authServices, err := auth.NewAuthServices(authConfig, store)
	if err != nil {
		log.Fatalf("创建认证服务失败: %v", err)
	}

	// 创建服务注册中心
	serviceRegistry := registry.NewServiceRegistry(store)

	// 注册功能API服务
	if err := provider.RegisterServices(serviceRegistry); err != nil {
		log.Fatalf("注册功能API服务失败: %v", err)
	}

	// 创建路由器
	mainRouter := router.NewRouter(store, authServices, serviceRegistry)

	// 设置路由
	engine := mainRouter.SetupRoutes()

	// 启动服务器
	log.Println("启动APIHub服务器，监听端口 :8080")
	log.Println("API文档: http://localhost:8080/swagger/index.html")
	log.Println("认证端点:")
	log.Println("  POST /api/v1/auth/login")
	log.Println("  POST /api/v1/auth/logout")
	log.Println("  GET  /api/v1/auth/profile")
	log.Println("API密钥端点:")
	log.Println("  GET  /api/v1/dashboard/apikeys/list")
	log.Println("  POST /api/v1/dashboard/apikeys/generate")
	log.Println("  POST /api/v1/dashboard/apikeys/delete")
	log.Println("功能API端点:")
	log.Println("  GET  /api/v1/provider/services")
	log.Println("  POST /api/v1/provider/:service/execute")
	log.Println("  POST /api/v1/provider/:service/public")

	if err := engine.Run(":8080"); err != nil {
		log.Fatalf("启动服务器失败: %v", err)
	}
}
