package provider

import (
	"apihub/internal/provider/registry"
	"apihub/internal/provider/services"
)

// RegisterServices 注册所有服务
func RegisterServices(registry *registry.ServiceRegistry) error {
	// 注册Echo服务
	if err := registry.RegisterService("echo", services.EchoServiceHandler, services.EchoServiceConfig()); err != nil {
		return err
	}

	// 注册时间服务
	if err := registry.RegisterService("time", services.TimeServiceHandler, services.TimeServiceConfig()); err != nil {
		return err
	}

	return nil
}
