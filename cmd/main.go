package main

import "github.com/gin-gonic/gin"

func MyMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 在处理请求之前
		c.Next()
		// 在处理请求之后
	}
}

func main() {
	router := gin.Default()
	router.Use(MyMiddleware()) // 应用中间件
	router.GET("/hello", func(c *gin.Context) {
		c.String(200, "Hello, World!")
	})
	err := router.Run(":8080")
	if err != nil {
		return
	}
}
