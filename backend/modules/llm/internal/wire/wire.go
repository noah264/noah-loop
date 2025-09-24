//go:build wireinject

package wire

import (
	"github.com/google/wire"
	"github.com/noah-loop/backend/modules/llm/internal/application/service"
	httpHandler "github.com/noah-loop/backend/modules/llm/internal/interface/http"
	"github.com/noah-loop/backend/modules/llm/internal/infrastructure/providers"
	"github.com/noah-loop/backend/modules/llm/internal/infrastructure/repository"
	"github.com/noah-loop/backend/shared/pkg/infrastructure"
)

// LLMApp LLM应用结构
type LLMApp struct {
	LLMService *service.LLMService
	Handler    *httpHandler.LLMHandler
	Router     *httpHandler.Router
	Metrics    *infrastructure.MetricsRegistry
	Database   *infrastructure.Database
}

// InitializeLLMApp 初始化LLM应用
func InitializeLLMApp() (*LLMApp, func(), error) {
	wire.Build(
		// 基础设施
		infrastructure.InfrastructureProviderSet,
		
		// 仓储
		LLMRepositoryProviderSet,
		
		// 应用服务
		LLMServiceProviderSet,
		
		// HTTP处理器和路由
		LLMHandlerProviderSet,
		
		// 应用结构
		wire.Struct(new(LLMApp), "*"),
		
		// 提供服务名称
		wire.Value("llm"),
	)
	
	return &LLMApp{}, nil, nil
}

// LLMRepositoryProviderSet 仓储提供者集合
var LLMRepositoryProviderSet = wire.NewSet(
	repository.NewGormModelRepository,
	repository.NewGormRequestRepository,
)

// LLMServiceProviderSet 应用服务提供者集合
var LLMServiceProviderSet = wire.NewSet(
	service.NewLLMService,
	// 事件总线暂时为nil
	wire.Value((interface{})(nil)),
	wire.Bind(new(interface{}), new(interface{})),
)

// LLMHandlerProviderSet HTTP处理器提供者集合
var LLMHandlerProviderSet = wire.NewSet(
	httpHandler.NewLLMHandler,
	httpHandler.NewRouter,
)

// ProvidersProviderSet Provider提供者集合
var ProvidersProviderSet = wire.NewSet(
	ProvideOpenAIProvider,
)

// ProvideOpenAIProvider 提供OpenAI Provider
func ProvideOpenAIProvider(logger infrastructure.Logger) *providers.OpenAIProvider {
	// 从环境变量获取API Key
	apiKey := "sk-test" // 实际应该从配置或环境变量获取
	return providers.NewOpenAIProvider(apiKey, logger)
}
