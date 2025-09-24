//go:build wireinject

package wire

import (
	"github.com/google/wire"
	"github.com/noah-loop/backend/modules/agent/internal/application/service"
	"github.com/noah-loop/backend/modules/agent/internal/domain"
	"github.com/noah-loop/backend/modules/agent/internal/infrastructure/executors"
	httpHandler "github.com/noah-loop/backend/modules/agent/internal/interface/http"
	"github.com/noah-loop/backend/modules/agent/internal/infrastructure/repository"
	"github.com/noah-loop/backend/shared/pkg/infrastructure"
)

// AgentApp Agent应用结构
type AgentApp struct {
	AgentService *service.AgentService
	Handler      *httpHandler.AgentHandler
	Router       *httpHandler.Router
	Metrics      *infrastructure.MetricsRegistry
	Database     *infrastructure.Database
}

// InitializeAgentApp 初始化Agent应用
func InitializeAgentApp() (*AgentApp, func(), error) {
	wire.Build(
		// 基础设施
		infrastructure.InfrastructureProviderSet,
		
		// 仓储
		AgentRepositoryProviderSet,
		
		// 应用服务
		AgentServiceProviderSet,
		
		// 工具执行器
		ToolExecutorProviderSet,
		
		// HTTP处理器和路由
		AgentHandlerProviderSet,
		
		// 应用结构
		wire.Struct(new(AgentApp), "*"),
		
		// 提供服务名称
		wire.Value("agent"),
	)
	
	return &AgentApp{}, nil, nil
}

// AgentRepositoryProviderSet 仓储提供者集合
var AgentRepositoryProviderSet = wire.NewSet(
	repository.NewGormAgentRepository,
	repository.NewGormToolRepository,
	repository.NewGormToolExecutionRepository,
)

// AgentServiceProviderSet 应用服务提供者集合
var AgentServiceProviderSet = wire.NewSet(
	NewAgentServiceWithExecutors,
	// 事件总线暂时为nil
	wire.Value((interface{})(nil)),
	wire.Bind(new(interface{}), new(interface{})),
)

// ToolExecutorProviderSet 工具执行器提供者集合
var ToolExecutorProviderSet = wire.NewSet(
	executors.NewCalculatorExecutor,
)

// AgentHandlerProviderSet HTTP处理器提供者集合
var AgentHandlerProviderSet = wire.NewSet(
	httpHandler.NewAgentHandler,
	httpHandler.NewRouter,
)

// NewAgentServiceWithExecutors 创建带有执行器的Agent服务
func NewAgentServiceWithExecutors(
	agentRepo domain.AgentRepository,
	toolRepo domain.ToolRepository,
	toolExecutionRepo domain.ToolExecutionRepository,
	eventBus interface{},
	logger infrastructure.Logger,
	metrics *infrastructure.MetricsRegistry,
	calculatorExecutor service.ToolExecutor,
) *service.AgentService {
	agentService := service.NewAgentService(agentRepo, toolRepo, toolExecutionRepo, eventBus, logger, metrics)
	
	// 注册工具执行器
	agentService.RegisterToolExecutor(domain.ToolTypeCalculator, calculatorExecutor)
	
	// 启动指标收集
	agentService.StartMetricsCollection()
	
	return agentService
}
