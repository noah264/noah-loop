//go:build wireinject

package wire

import (
	"github.com/google/wire"
	"github.com/noah-loop/backend/modules/orchestrator/internal/application/service"
	httpHandler "github.com/noah-loop/backend/modules/orchestrator/internal/interface/http"
	"github.com/noah-loop/backend/shared/pkg/infrastructure"
)

// OrchestratorApp 编排器应用结构
type OrchestratorApp struct {
	OrchestratorService *service.OrchestratorService
	Handler             *httpHandler.OrchestratorHandler
	Router              *httpHandler.Router
	Metrics             *infrastructure.MetricsRegistry
	Database            *infrastructure.Database
}

// InitializeOrchestratorApp 初始化编排器应用
func InitializeOrchestratorApp() (*OrchestratorApp, func(), error) {
	wire.Build(
		// 基础设施
		infrastructure.InfrastructureProviderSet,
		
		// TODO: 仓储（当仓储实现完成后取消注释）
		// OrchestratorRepositoryProviderSet,
		
		// 应用服务
		OrchestratorServiceProviderSet,
		
		// HTTP处理器和路由
		OrchestratorHandlerProviderSet,
		
		// 应用结构
		wire.Struct(new(OrchestratorApp), "*"),
		
		// 提供服务名称
		wire.Value("orchestrator"),
	)
	
	return &OrchestratorApp{}, nil, nil
}

// TODO: 仓储提供者集合（当仓储实现完成后取消注释）
// var OrchestratorRepositoryProviderSet = wire.NewSet(
// 	repository.NewGormWorkflowRepository,
// 	repository.NewGormStepRepository,
// 	repository.NewGormTriggerRepository,
// 	repository.NewGormExecutionRepository,
// 	repository.NewGormStepExecutionRepository,
// )

// OrchestratorServiceProviderSet 应用服务提供者集合
var OrchestratorServiceProviderSet = wire.NewSet(
	NewOrchestratorServiceStub,
)

// OrchestratorHandlerProviderSet HTTP处理器提供者集合
var OrchestratorHandlerProviderSet = wire.NewSet(
	httpHandler.NewOrchestratorHandler,
	httpHandler.NewRouter,
)

// NewOrchestratorServiceStub 创建编排器服务存根（临时实现，直到完整实现完成）
func NewOrchestratorServiceStub(
	logger infrastructure.Logger,
	metrics *infrastructure.MetricsRegistry,
) *service.OrchestratorService {
	// TODO: 当仓储实现完成后，使用真实的仓储创建服务
	// return service.NewOrchestratorService(workflowRepo, stepRepo, triggerRepo, executionRepo, stepExecutionRepo, eventBus, logger, metrics)
	
	// 目前创建一个带有nil仓储的服务实例用于基本功能
	return service.NewOrchestratorService(
		nil, // workflowRepo
		nil, // stepRepo  
		nil, // triggerRepo
		nil, // executionRepo
		nil, // stepExecutionRepo
		nil, // eventBus
		logger,
		metrics,
	)
}
