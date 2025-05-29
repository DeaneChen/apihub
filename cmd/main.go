package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"

	"apihub/internal/core"
)

func main() {
	// 创建上下文
	ctx := context.Background()

	// 创建存储层实例
	store := core.CreateSQLiteStore("apihub.db")

	// 创建初始化服务
	initService := core.NewInitializationService(store)

	// 执行系统初始化
	if err := initService.InitializeSystem(ctx); err != nil {
		log.Fatalf("系统初始化失败: %v", err)
	}

	// 创建Gin路由器
	router := gin.Default()

	// 添加基础中间件
	router.Use(LoggerMiddleware())
	router.Use(CORSMiddleware())
	router.Use(ErrorHandlerMiddleware())

	// 健康检查端点
	router.GET("/health", func(c *gin.Context) {
		status, err := initService.GetSystemStatus(ctx)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error":   "获取系统状态失败",
				"details": err.Error(),
			})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"status": "ok",
			"data":   status,
		})
	})

	// 系统信息端点
	router.GET("/info", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"name":        "APIHub",
			"version":     "1.0.0",
			"description": "统一API业务服务框架",
		})
	})

	// 测试端点
	router.GET("/hello", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"message":   "Hello, APIHub!",
			"timestamp": time.Now().Unix(),
		})
	})

	// 启动服务器
	log.Println("APIHub 服务启动中...")
	log.Println("服务地址: http://localhost:8080")
	log.Println("健康检查: http://localhost:8080/health")
	log.Println("系统信息: http://localhost:8080/info")

	if err := router.Run(":8080"); err != nil {
		log.Fatalf("服务启动失败: %v", err)
	}
}

// LoggerMiddleware 日志中间件
func LoggerMiddleware() gin.HandlerFunc {
	return gin.LoggerWithFormatter(func(param gin.LogFormatterParams) string {
		return fmt.Sprintf("[%s] %s %s %d %s %s\n",
			param.TimeStamp.Format("2006-01-02 15:04:05"),
			param.Method,
			param.Path,
			param.StatusCode,
			param.Latency,
			param.ClientIP,
		)
	})
}

// CORSMiddleware CORS中间件
func CORSMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Origin, Content-Type, Authorization")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}

		c.Next()
	}
}

// ErrorHandlerMiddleware 错误处理中间件
func ErrorHandlerMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if err := recover(); err != nil {
				log.Printf("Panic recovered: %v", err)
				c.JSON(http.StatusInternalServerError, gin.H{
					"error": "内部服务器错误",
					"code":  500,
				})
				c.Abort()
			}
		}()

		c.Next()
	}
}
