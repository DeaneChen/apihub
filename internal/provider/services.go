package provider

import (
	"apihub/internal/model"
	"apihub/internal/provider/registry"
	"apihub/internal/provider/services"
)

// RegisterServices 注册所有服务
func RegisterServices(registry *registry.ServiceRegistry) error {
	// 注册Echo服务
	if err := registry.RegisterService("echo", services.EchoServiceHandler, model.ServiceConfig{
		AllowAnonymous:  true,
		RateLimit:       60, // 每分钟60次
		QuotaCost:       1,  // 消耗1个配额
		Description:     "回显服务，返回请求的内容",
		RequestExample:  map[string]interface{}{"message": "Hello, APIHub!"},
		ResponseExample: map[string]interface{}{"message": "Hello, APIHub!", "timestamp": 1625097600},
	}); err != nil {
		return err
	}

	// 注册时间服务
	if err := registry.RegisterService("time", services.TimeServiceHandler, model.ServiceConfig{
		AllowAnonymous: true,
		RateLimit:      60, // 每分钟60次
		QuotaCost:      1,  // 消耗1个配额
		Description:    "时间服务，返回当前服务器时间",
		RequestExample: map[string]interface{}{},
		ResponseExample: map[string]interface{}{
			"timestamp": 1625097600,
			"iso8601":   "2021-07-01T00:00:00Z",
			"date":      "2021-07-01",
			"time":      "00:00:00",
			"timezone":  "UTC",
		},
	}); err != nil {
		return err
	}

	return nil
}
