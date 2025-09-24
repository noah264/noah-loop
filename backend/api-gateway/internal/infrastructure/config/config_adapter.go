package config

import (
	"github.com/noah-loop/backend/api-gateway/internal/application/service"
	"github.com/noah-loop/backend/shared/pkg/infrastructure"
)

// ConfigAdapter 配置适配器
type ConfigAdapter struct {
	config *infrastructure.Config
}

// NewConfigAdapter 创建配置适配器
func NewConfigAdapter(config *infrastructure.Config) *ConfigAdapter {
	return &ConfigAdapter{
		config: config,
	}
}

// GetGatewayName 获取网关名称
func (c *ConfigAdapter) GetGatewayName() string {
	return c.config.App.Name + "-gateway"
}

// GetGatewayVersion 获取网关版本
func (c *ConfigAdapter) GetGatewayVersion() string {
	return c.config.App.Version
}

// GetServices 获取服务配置
func (c *ConfigAdapter) GetServices() map[string]service.ServiceConfig {
	services := make(map[string]service.ServiceConfig)
	
	// Agent服务
	services["agent"] = service.ServiceConfig{
		Name: "agent",
		Host: "localhost",
		Port: c.config.Services.Agent.Port,
		Path: "/api/v1/agent",
	}
	
	// LLM服务
	services["llm"] = service.ServiceConfig{
		Name: "llm",
		Host: "localhost",
		Port: c.config.Services.LLM.Port,
		Path: "/api/v1/llm",
	}
	
	// MCP服务
	services["mcp"] = service.ServiceConfig{
		Name: "mcp",
		Host: "localhost",
		Port: c.config.Services.MCP.Port,
		Path: "/api/v1/mcp",
	}
	
	// Orchestrator服务
	services["orchestrator"] = service.ServiceConfig{
		Name: "orchestrator",
		Host: "localhost",
		Port: c.config.Services.Orchestrator.Port,
		Path: "/api/v1/orchestrator",
	}
	
	return services
}
