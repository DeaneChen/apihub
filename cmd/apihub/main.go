package main

import (
	"apihub/configs"
	"apihub/internal/core"
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"

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
	// 解析命令行参数
	configPath := flag.String("config", "", "配置文件路径")
	generateConfig := flag.Bool("generate-config", false, "生成默认配置文件")
	flag.Parse()

	// 如果指定了生成配置文件
	if *generateConfig {
		// 如果没有指定配置路径，使用默认路径
		configFilePath := *configPath
		if configFilePath == "" {
			configFilePath = "config.json"
		}

		if err := configs.GenerateConfigFile(configFilePath); err != nil {
			log.Fatalf("生成配置文件失败: %v", err)
		}
		log.Printf("配置文件已生成: %s", configFilePath)
		return
	}

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

	// 确保数据库保持连接
	if err := store.Connect(); err != nil {
		log.Fatalf("数据库连接失败: %v", err)
	}

	// 查找配置文件
	configFilePath := findConfigFile(*configPath)
	log.Printf("使用配置文件: %s", configFilePath)

	// 加载配置
	config, err := configs.LoadConfig(configFilePath, store)
	if err != nil {
		log.Fatalf("加载配置失败: %v", err)
	}

	// 创建认证服务配置
	authConfig := auth.AuthConfig{
		JWT: auth.JWTConfig{
			AccessExpiry:  config.Auth.JWT.AccessExpiry,
			Issuer:        config.Auth.JWT.Issuer,
			PrivateKeyPEM: "", // 留空将自动生成密钥对
			PublicKeyPEM:  "",
		},
		Crypto: auth.CryptoConfig{
			SecretKey: config.Auth.APIKey.Secret, // 使用配置中的APIKey密钥
		},
		Cache: auth.CacheConfig{
			DefaultExpiration: config.Auth.Cache.DefaultExpiration,
			CleanupInterval:   config.Auth.Cache.CleanupInterval,
		},
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

	// 构建服务器地址
	address := config.Server.Host + ":" + fmt.Sprintf("%d", config.Server.Port)

	// 启动服务器
	log.Printf("启动APIHub服务器，监听地址 %s", address)
	log.Println("API文档: http://" + address + "/swagger/index.html")
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

	if err := engine.Run(address); err != nil {
		log.Fatalf("启动服务器失败: %v", err)
	}
}

// findConfigFile 查找配置文件
// 按以下顺序查找:
// 1. 命令行参数指定的路径
// 2. 当前工作目录
// 3. 可执行文件所在目录
// 4. 系统配置目录
func findConfigFile(configPath string) string {
	// 如果命令行参数指定了配置文件路径，直接使用
	if configPath != "" {
		return configPath
	}

	// 配置文件名
	const configFileName = "config.json"

	// 尝试在当前工作目录查找
	if _, err := os.Stat(configFileName); err == nil {
		absPath, _ := filepath.Abs(configFileName)
		return absPath
	}

	// 尝试在可执行文件目录查找
	execPath, err := os.Executable()
	if err == nil {
		execDir := filepath.Dir(execPath)
		execConfig := filepath.Join(execDir, configFileName)
		if _, err := os.Stat(execConfig); err == nil {
			return execConfig
		}
	}

	// 尝试在系统配置目录查找
	// Linux/Unix: /etc/apihub/config.json
	// Windows: %PROGRAMDATA%\apihub\config.json
	// macOS: /Library/Application Support/apihub/config.json
	var systemConfigPaths []string

	// Unix-like系统
	systemConfigPaths = append(systemConfigPaths, "/etc/apihub/"+configFileName)

	// macOS
	systemConfigPaths = append(systemConfigPaths, "/Library/Application Support/apihub/"+configFileName)

	// Windows
	if programData := os.Getenv("PROGRAMDATA"); programData != "" {
		systemConfigPaths = append(systemConfigPaths, filepath.Join(programData, "apihub", configFileName))
	}

	// 用户主目录
	if homeDir, err := os.UserHomeDir(); err == nil {
		// Unix-like系统
		systemConfigPaths = append(systemConfigPaths, filepath.Join(homeDir, ".config/apihub", configFileName))

		// macOS
		systemConfigPaths = append(systemConfigPaths, filepath.Join(homeDir, "Library/Application Support/apihub", configFileName))
	}

	// 检查系统配置路径
	for _, path := range systemConfigPaths {
		if _, err := os.Stat(path); err == nil {
			return path
		}
	}

	// 如果都没找到，返回默认路径（当前工作目录）
	return configFileName
}
