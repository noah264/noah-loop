package service

import (
	"context"
	"fmt"
	"time"
	
	"github.com/google/uuid"
	"github.com/noah-loop/backend/modules/llm/internal/domain"
	"github.com/noah-loop/backend/shared/pkg/application"
	"github.com/noah-loop/backend/shared/pkg/infrastructure"
	"go.uber.org/zap"
)

// LLMService 大模型应用服务
type LLMService struct {
	modelRepo   domain.ModelRepository
	requestRepo domain.RequestRepository
	providers   map[domain.ModelProvider]Provider
	eventBus    application.EventBus
	logger      infrastructure.Logger
	metrics     *infrastructure.MetricsRegistry
}

// NewLLMService 创建大模型服务
func NewLLMService(
	modelRepo domain.ModelRepository,
	requestRepo domain.RequestRepository,
	eventBus application.EventBus,
	logger infrastructure.Logger,
	metrics *infrastructure.MetricsRegistry,
) *LLMService {
	return &LLMService{
		modelRepo:   modelRepo,
		requestRepo: requestRepo,
		providers:   make(map[domain.ModelProvider]Provider),
		eventBus:    eventBus,
		logger:      logger,
		metrics:     metrics,
	}
}

// RegisterProvider 注册模型提供商
func (s *LLMService) RegisterProvider(provider domain.ModelProvider, impl Provider) {
	s.providers[provider] = impl
}

// CreateModel 创建模型
func (s *LLMService) CreateModel(ctx context.Context, cmd *CreateModelCommand) (*application.Result, error) {
	if err := cmd.Validate(); err != nil {
		return &application.Result{Success: false, Error: err.Error()}, err
	}
	
	// 检查模型是否已存在
	existing, err := s.modelRepo.FindByNameAndProvider(ctx, cmd.Name, cmd.Provider)
	if err == nil && existing != nil {
		return &application.Result{Success: false, Error: "model already exists"}, fmt.Errorf("model already exists")
	}
	
	// 创建模型
	model := domain.NewModel(cmd.Name, cmd.Provider, cmd.Type)
	model.Version = cmd.Version
	model.Description = cmd.Description
	model.Config = cmd.Config
	model.Capabilities = cmd.Capabilities
	model.MaxTokens = cmd.MaxTokens
	model.PricePerK = cmd.PricePerK
	
	// 保存模型
	if err := s.modelRepo.Save(ctx, model); err != nil {
		s.logger.Error("Failed to save model", zap.Error(err))
		return &application.Result{Success: false, Error: "failed to save model"}, err
	}
	
	// 发布事件
	for _, event := range model.GetDomainEvents() {
		if err := s.eventBus.Publish(ctx, event); err != nil {
			s.logger.Warn("Failed to publish event", zap.Error(err))
		}
	}
	model.ClearDomainEvents()
	
	return &application.Result{Success: true, Data: model}, nil
}

// ProcessRequest 处理请求
func (s *LLMService) ProcessRequest(ctx context.Context, cmd *ProcessRequestCommand) (*application.Result, error) {
	if err := cmd.Validate(); err != nil {
		return &application.Result{Success: false, Error: err.Error()}, err
	}
	
	// 获取模型
	model, err := s.modelRepo.FindByID(ctx, cmd.ModelID)
	if err != nil {
		return &application.Result{Success: false, Error: "model not found"}, err
	}
	
	if !model.IsActive {
		return &application.Result{Success: false, Error: "model is not active"}, fmt.Errorf("model is not active")
	}
	
	// 创建请求
	request := domain.NewRequest(cmd.ModelID, cmd.UserID, cmd.SessionID, cmd.RequestType, cmd.Input)
	
	// 保存请求
	if err := s.requestRepo.Save(ctx, request); err != nil {
		return &application.Result{Success: false, Error: "failed to save request"}, err
	}
	
	// 开始处理
	if err := request.StartProcessing(); err != nil {
		return &application.Result{Success: false, Error: err.Error()}, err
	}
	
	// 更新请求状态
	if err := s.requestRepo.Save(ctx, request); err != nil {
		return &application.Result{Success: false, Error: "failed to update request"}, err
	}
	
	// 获取提供商
	provider, exists := s.providers[model.Provider]
	if !exists {
		request.Fail("provider not found")
		s.requestRepo.Save(ctx, request)
		return &application.Result{Success: false, Error: "provider not found"}, fmt.Errorf("provider not found")
	}
	
	// 异步处理请求
	go s.processRequestAsync(ctx, request, model, provider)
	
	return &application.Result{Success: true, Data: request}, nil
}

// processRequestAsync 异步处理请求
func (s *LLMService) processRequestAsync(ctx context.Context, request *domain.Request, model *domain.Model, provider Provider) {
	startTime := time.Now()
	
	defer func() {
		if r := recover(); r != nil {
			s.logger.Error("Panic in processRequestAsync", zap.Any("panic", r))
			request.Fail(fmt.Sprintf("internal error: %v", r))
			s.requestRepo.Save(ctx, request)
		}
	}()
	
	// 调用提供商处理
	response, err := provider.Process(ctx, &ProviderRequest{
		Model:  model,
		Input:  request.Input,
		Config: model.Config,
	})
	
	duration := time.Since(startTime)
	
	if err != nil {
		s.logger.Error("Provider processing failed", zap.Error(err))
		request.Fail(err.Error())
	} else {
		// 计算成本
		cost := float64(response.TokensUsed) / 1000 * model.PricePerK
		request.Complete(response.Output, response.TokensUsed, cost, duration)
		
		// 记录Token消耗指标
		if s.metrics != nil {
			s.metrics.RecordTokenConsumption(
				string(model.Provider),
				model.Name,
				request.UserID.String(),
				request.RequestType,
				response.TokensUsed,
				cost,
			)
		}
	}
	
	// 保存结果
	if err := s.requestRepo.Save(ctx, request); err != nil {
		s.logger.Error("Failed to save request result", zap.Error(err))
	}
	
	// 发布事件
	for _, event := range request.GetDomainEvents() {
		if err := s.eventBus.Publish(ctx, event); err != nil {
			s.logger.Warn("Failed to publish event", zap.Error(err))
		}
	}
	request.ClearDomainEvents()
}

// Provider 模型提供商接口
type Provider interface {
	Process(ctx context.Context, request *ProviderRequest) (*ProviderResponse, error)
	Health(ctx context.Context) error
}

// ProviderRequest 提供商请求
type ProviderRequest struct {
	Model  *domain.Model
	Input  map[string]interface{}
	Config map[string]interface{}
}

// ProviderResponse 提供商响应
type ProviderResponse struct {
	Output     map[string]interface{}
	TokensUsed int
	Metadata   map[string]interface{}
}
