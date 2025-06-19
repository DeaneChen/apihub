package middleware

import (
	"fmt"
	"net/http"
	"sync"
	"time"

	"apihub/internal/model"
	"apihub/internal/provider/registry"

	"github.com/gin-gonic/gin"
)

// 限流器结构体，用于存储不同类型的限流器
type RateLimiter struct {
	mu            sync.RWMutex
	ipLimiters    map[string]*rateLimiterEntry // IP地址 -> 限流器条目
	userLimiters  map[int]*rateLimiterEntry    // 用户ID -> 限流器条目
	serviceLimits map[string]int               // 服务名称 -> 限流值(每分钟)
	defaultLimit  int                          // 默认限流值(每分钟)
}

// 限流器条目，包含限流器和最后访问时间
type rateLimiterEntry struct {
	count       int       // 当前时间窗口内的请求计数
	windowStart time.Time // 当前时间窗口的开始时间
	limit       int       // 限流值(每分钟)
	lastAccess  time.Time // 最后访问时间
}

// 创建新的限流器
func NewRateLimiter(defaultLimit int) *RateLimiter {
	return &RateLimiter{
		ipLimiters:    make(map[string]*rateLimiterEntry),
		userLimiters:  make(map[int]*rateLimiterEntry),
		serviceLimits: make(map[string]int),
		defaultLimit:  defaultLimit, // 每分钟请求数
	}
}

// 检查并更新IP限流器
func (r *RateLimiter) checkIPLimit(ip string, serviceLimit int) bool {
	r.mu.Lock()
	defer r.mu.Unlock()

	now := time.Now()

	// 获取或创建限流器条目
	entry, exists := r.ipLimiters[ip]
	if !exists || now.Sub(entry.windowStart) > time.Minute {
		// 创建新条目或重置时间窗口
		limit := serviceLimit
		if limit <= 0 {
			limit = r.defaultLimit
		}

		r.ipLimiters[ip] = &rateLimiterEntry{
			count:       1,
			windowStart: now,
			limit:       limit,
			lastAccess:  now,
		}
		return true // 允许请求
	}

	// 更新最后访问时间
	entry.lastAccess = now

	// 检查是否超出限制
	if entry.count >= entry.limit {
		return false // 拒绝请求
	}

	// 增加计数
	entry.count++
	return true // 允许请求
}

// 检查并更新用户限流器
func (r *RateLimiter) checkUserLimit(userID int, serviceLimit int) bool {
	r.mu.Lock()
	defer r.mu.Unlock()

	now := time.Now()

	// 获取或创建限流器条目
	entry, exists := r.userLimiters[userID]
	if !exists || now.Sub(entry.windowStart) > time.Minute {
		// 创建新条目或重置时间窗口
		limit := serviceLimit
		if limit <= 0 {
			limit = r.defaultLimit
		}

		r.userLimiters[userID] = &rateLimiterEntry{
			count:       1,
			windowStart: now,
			limit:       limit,
			lastAccess:  now,
		}
		return true // 允许请求
	}

	// 更新最后访问时间
	entry.lastAccess = now

	// 检查是否超出限制
	if entry.count >= entry.limit {
		return false // 拒绝请求
	}

	// 增加计数
	entry.count++
	return true // 允许请求
}

// 设置服务限流值
func (r *RateLimiter) SetServiceLimit(serviceName string, limit int) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.serviceLimits[serviceName] = limit
}

// 获取服务限流值
func (r *RateLimiter) GetServiceLimit(serviceName string) int {
	r.mu.RLock()
	defer r.mu.RUnlock()
	if limit, exists := r.serviceLimits[serviceName]; exists {
		return limit
	}
	return r.defaultLimit // 返回默认限流值(每分钟)
}

// CleanupExpired 清理过期的限流器
// 删除超过指定时间未访问的限流器
func (r *RateLimiter) CleanupExpired(maxAge time.Duration) {
	r.mu.Lock()
	defer r.mu.Unlock()

	now := time.Now()

	// 清理IP限流器
	for ip, entry := range r.ipLimiters {
		if now.Sub(entry.lastAccess) > maxAge {
			delete(r.ipLimiters, ip)
		}
	}

	// 清理用户限流器
	for userID, entry := range r.userLimiters {
		if now.Sub(entry.lastAccess) > maxAge {
			delete(r.userLimiters, userID)
		}
	}
}

// StartCleanupTask 启动定期清理任务
func (r *RateLimiter) StartCleanupTask(interval, maxAge time.Duration) {
	go func() {
		ticker := time.NewTicker(interval)
		defer ticker.Stop()

		for range ticker.C {
			r.CleanupExpired(maxAge)
		}
	}()
}

// RateLimitMiddleware 创建限流中间件
// 根据不同的认证方式（匿名/认证用户）应用不同的限流策略
func RateLimitMiddleware(limiter *RateLimiter) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 获取服务名称
		serviceName := c.Param("service")
		if serviceName == "" {
			// 如果路径中没有服务名称参数，尝试从上下文中获取
			if serviceInfo, exists := c.Get("service_info"); exists {
				if si, ok := serviceInfo.(*registry.ServiceInfo); ok {
					serviceName = si.Definition.ServiceName
				}
			}
		}

		// 如果无法确定服务名称，使用默认限流
		serviceLimit := limiter.GetServiceLimit(serviceName)

		// 检查是否为认证用户
		userID, exists := GetCurrentUserID(c)

		var allowed bool

		if exists && userID > 0 {
			// 认证用户 - 使用用户级限流
			allowed = limiter.checkUserLimit(userID, serviceLimit)
		} else {
			// 匿名用户 - 使用IP级限流
			allowed = limiter.checkIPLimit(c.ClientIP(), serviceLimit)
		}

		if !allowed {
			c.JSON(http.StatusTooManyRequests, model.NewErrorResponse(
				model.CodeRateLimitExceeded,
				"请求过于频繁，请稍后再试",
			))
			c.Abort()
			return
		}

		c.Next()
	}
}

// ServiceRateLimitMiddleware 服务级限流中间件
// 需要服务信息已经被设置到上下文中
func ServiceRateLimitMiddleware(limiter *RateLimiter) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 从上下文中获取服务信息
		serviceInfo, exists := c.Get("service_info")
		if !exists {
			// 如果没有服务信息，跳过限流检查
			c.Next()
			return
		}

		// 获取服务定义
		var rateLimit int
		var serviceName string

		// 类型断言，获取服务限流值
		if si, ok := serviceInfo.(*registry.ServiceInfo); ok {
			rateLimit = si.Definition.RateLimit
			serviceName = si.Definition.ServiceName
		} else {
			// 如果无法获取服务限流值，使用默认值
			rateLimit = limiter.defaultLimit
			serviceName = c.Param("service")
		}

		// 更新服务限流值
		if rateLimit > 0 {
			limiter.SetServiceLimit(serviceName, rateLimit)
		}

		// 获取用户ID或IP地址
		userID, userExists := GetCurrentUserID(c)

		var allowed bool

		if userExists && userID > 0 {
			// 认证用户 - 使用用户级限流
			allowed = limiter.checkUserLimit(userID, rateLimit)

			if !allowed {
				fmt.Printf("用户 %d 访问服务 %s 被限流\n", userID, serviceName)
			}
		} else {
			// 匿名用户 - 使用IP级限流
			ip := c.ClientIP()
			allowed = limiter.checkIPLimit(ip, rateLimit)

			if !allowed {
				fmt.Printf("IP %s 访问服务 %s 被限流\n", ip, serviceName)
			}
		}

		if !allowed {
			c.JSON(http.StatusTooManyRequests, model.NewErrorResponse(
				model.CodeRateLimitExceeded,
				"请求过于频繁，请稍后再试",
			))
			c.Abort()
			return
		}

		c.Next()
	}
}
