package registry

import (
	"context"
	"fmt"
	"sync"

	"apihub/internal/model"
	"apihub/internal/store"

	"github.com/gin-gonic/gin"
)

// ServiceHandler 服务处理函数类型
type ServiceHandler func(c *gin.Context) (interface{}, error)

// ServiceInfo 服务信息
type ServiceInfo struct {
	// 服务定义（来自数据库）
	Definition *model.ServiceDefinition
	// 服务处理函数
	Handler ServiceHandler
}

// ServiceRegistry 服务注册中心
type ServiceRegistry struct {
	// 服务映射表 serviceName -> ServiceInfo
	services map[string]*ServiceInfo
	// 存储层接口
	store store.Store
	// 互斥锁，保护services映射表
	mu sync.RWMutex
}

// NewServiceRegistry 创建服务注册中心
func NewServiceRegistry(store store.Store) *ServiceRegistry {
	return &ServiceRegistry{
		services: make(map[string]*ServiceInfo),
		store:    store,
	}
}

// RegisterService 注册服务
func (r *ServiceRegistry) RegisterService(name string, handler ServiceHandler, config model.ServiceConfig) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	// 检查服务是否已存在于内存中
	if _, exists := r.services[name]; exists {
		return fmt.Errorf("服务 %s 已存在于内存中", name)
	}

	// 从数据库获取服务定义
	definition, err := r.store.Services().GetByName(context.Background(), name)
	if err != nil {
		// 服务定义不存在于数据库，创建新定义
		definition = &model.ServiceDefinition{
			ServiceName:    name,
			Description:    config.Description,
			DefaultLimit:   1000, // 默认每日配额
			Status:         model.ServiceStatusEnabled,
			AllowAnonymous: config.AllowAnonymous,
			RateLimit:      config.RateLimit,
			QuotaCost:      config.QuotaCost,
		}

		// 保存到数据库
		if err := r.store.Services().Create(context.Background(), definition); err != nil {
			return fmt.Errorf("创建服务定义失败: %w", err)
		}
	}
	// 如果服务已存在于数据库，则使用数据库中的配置，不更新数据库
	// 打印服务定义
	fmt.Println("服务定义:", definition)

	// 注册服务到内存
	r.services[name] = &ServiceInfo{
		Definition: definition,
		Handler:    handler,
	}

	return nil
}

// GetService 获取服务
func (r *ServiceRegistry) GetService(name string) (*ServiceInfo, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	service, exists := r.services[name]
	return service, exists
}

// ListServices 列出所有服务
func (r *ServiceRegistry) ListServices() []*ServiceInfo {
	r.mu.RLock()
	defer r.mu.RUnlock()

	services := make([]*ServiceInfo, 0, len(r.services))
	for _, service := range r.services {
		services = append(services, service)
	}

	return services
}

// GetServiceNames 获取所有服务名称
func (r *ServiceRegistry) GetServiceNames() []string {
	r.mu.RLock()
	defer r.mu.RUnlock()

	names := make([]string, 0, len(r.services))
	for name := range r.services {
		names = append(names, name)
	}

	return names
}

// ServiceCount 获取服务数量
func (r *ServiceRegistry) ServiceCount() int {
	r.mu.RLock()
	defer r.mu.RUnlock()

	return len(r.services)
}
