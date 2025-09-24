//go:build wireinject

package wire

import (
	"github.com/google/wire"
	"github.com/noah-loop/backend/modules/mcp/internal/application/service"
	"github.com/noah-loop/backend/modules/mcp/internal/domain"
	httpHandler "github.com/noah-loop/backend/modules/mcp/internal/interface/http"
	"github.com/noah-loop/backend/modules/mcp/internal/infrastructure/repository"
	"github.com/noah-loop/backend/shared/pkg/infrastructure"
)

// MCPApp MCP应用结构
type MCPApp struct {
	MCPService *service.MCPService
	Handler    *httpHandler.MCPHandler
	Router     *httpHandler.Router
	Metrics    *infrastructure.MetricsRegistry
	Database   *infrastructure.Database
}

// InitializeMCPApp 初始化MCP应用
func InitializeMCPApp() (*MCPApp, func(), error) {
	wire.Build(
		// 基础设施
		infrastructure.InfrastructureProviderSet,
		
		// 仓储
		MCPRepositoryProviderSet,
		
		// 应用服务
		MCPServiceProviderSet,
		
		// HTTP处理器和路由
		MCPHandlerProviderSet,
		
		// 应用结构
		wire.Struct(new(MCPApp), "*"),
		
		// 提供服务名称
		wire.Value("mcp"),
	)
	
	return &MCPApp{}, nil, nil
}

// MCPRepositoryProviderSet 仓储提供者集合
var MCPRepositoryProviderSet = wire.NewSet(
	repository.NewGormSessionRepository,
	repository.NewGormContextRepository,
)

// MCPServiceProviderSet 应用服务提供者集合
var MCPServiceProviderSet = wire.NewSet(
	NewMCPServiceWithMetrics,
	// 事件总线暂时为nil
	wire.Value((interface{})(nil)),
	wire.Bind(new(interface{}), new(interface{})),
)

// MCPHandlerProviderSet HTTP处理器提供者集合
var MCPHandlerProviderSet = wire.NewSet(
	httpHandler.NewMCPHandler,
	httpHandler.NewRouter,
)

// NewMCPServiceWithMetrics 创建带有指标收集的MCP服务
func NewMCPServiceWithMetrics(
	sessionRepo domain.SessionRepository,
	contextRepo domain.ContextRepository,
	eventBus interface{},
	logger infrastructure.Logger,
	metrics *infrastructure.MetricsRegistry,
) *service.MCPService {
	mcpService := service.NewMCPService(sessionRepo, contextRepo, eventBus, logger, metrics)
	
	// 启动指标收集
	mcpService.StartMetricsCollection()
	
	return mcpService
}
