package service

import (
	"context"
	"fmt"
	"time"
	
	"github.com/google/uuid"
	"github.com/noah-loop/backend/modules/agent/internal/domain"
	"github.com/noah-loop/backend/shared/pkg/application"
	"github.com/noah-loop/backend/shared/pkg/infrastructure"
	"go.uber.org/zap"
)

// AgentService 智能体应用服务
type AgentService struct {
	agentRepo           domain.AgentRepository
	toolRepo            domain.ToolRepository
	toolExecutionRepo   domain.ToolExecutionRepository
	eventBus            application.EventBus
	logger              infrastructure.Logger
	metrics             *infrastructure.MetricsRegistry
	toolExecutors       map[domain.ToolType]ToolExecutor
}

// NewAgentService 创建智能体服务
func NewAgentService(
	agentRepo domain.AgentRepository,
	toolRepo domain.ToolRepository,
	toolExecutionRepo domain.ToolExecutionRepository,
	eventBus application.EventBus,
	logger infrastructure.Logger,
	metrics *infrastructure.MetricsRegistry,
) *AgentService {
	return &AgentService{
		agentRepo:         agentRepo,
		toolRepo:          toolRepo,
		toolExecutionRepo: toolExecutionRepo,
		eventBus:          eventBus,
		logger:            logger,
		metrics:           metrics,
		toolExecutors:     make(map[domain.ToolType]ToolExecutor),
	}
}

// RegisterToolExecutor 注册工具执行器
func (s *AgentService) RegisterToolExecutor(toolType domain.ToolType, executor ToolExecutor) {
	s.toolExecutors[toolType] = executor
}

// CreateAgent 创建智能体
func (s *AgentService) CreateAgent(ctx context.Context, cmd *CreateAgentCommand) (*application.Result, error) {
	if err := cmd.Validate(); err != nil {
		return &application.Result{Success: false, Error: err.Error()}, err
	}
	
	// 创建智能体
	agent := domain.NewAgent(cmd.Name, cmd.Type, cmd.OwnerID)
	agent.Description = cmd.Description
	agent.SystemPrompt = cmd.SystemPrompt
	agent.Config = cmd.Config
	agent.Capabilities = cmd.Capabilities
	
	// 创建记忆系统
	memory := domain.NewAgentMemory(agent.ID)
	memory.Capacity = cmd.MemoryCapacity
	agent.Memory = memory
	
	// 保存智能体
	if err := s.agentRepo.Save(ctx, agent); err != nil {
		s.logger.Error("Failed to save agent", zap.Error(err))
		return &application.Result{Success: false, Error: "failed to save agent"}, err
	}
	
	// 发布事件
	for _, event := range agent.GetDomainEvents() {
		if err := s.eventBus.Publish(ctx, event); err != nil {
			s.logger.Warn("Failed to publish event", zap.Error(err))
		}
	}
	agent.ClearDomainEvents()
	
	// 记录智能体创建指标
	if s.metrics != nil {
		// 这里可以异步更新活跃智能体统计
		go s.updateAgentMetrics()
	}
	
	return &application.Result{Success: true, Data: agent}, nil
}

// ExecuteTool 执行工具
func (s *AgentService) ExecuteTool(ctx context.Context, cmd *ExecuteToolCommand) (*application.Result, error) {
	if err := cmd.Validate(); err != nil {
		return &application.Result{Success: false, Error: err.Error()}, err
	}
	
	// 获取智能体
	agent, err := s.agentRepo.FindByID(ctx, cmd.AgentID)
	if err != nil {
		return &application.Result{Success: false, Error: "agent not found"}, err
	}
	
	// 获取工具
	tool, err := s.toolRepo.FindByID(ctx, cmd.ToolID)
	if err != nil {
		return &application.Result{Success: false, Error: "tool not found"}, err
	}
	
	// 检查工具是否启用
	if !tool.IsEnabled {
		return &application.Result{Success: false, Error: "tool is disabled"}, fmt.Errorf("tool is disabled")
	}
	
	// 检查智能体是否可以使用该工具
	if !agent.CanUse(tool.Name) {
		return &application.Result{Success: false, Error: "agent cannot use this tool"}, fmt.Errorf("agent cannot use this tool")
	}
	
	// 验证输入
	if err := tool.ValidateInput(cmd.Input); err != nil {
		return &application.Result{Success: false, Error: err.Error()}, err
	}
	
	// 创建执行记录
	execution := domain.NewToolExecution(tool.ID, agent.ID, cmd.Input)
	execution.Status = domain.ExecutionStatusRunning
	
	// 保存执行记录
	if err := s.toolExecutionRepo.Save(ctx, execution); err != nil {
		return &application.Result{Success: false, Error: "failed to save execution"}, err
	}
	
	// 获取执行器
	executor, exists := s.toolExecutors[tool.Type]
	if !exists {
		execution.Fail("no executor found for tool type", 0)
		s.toolExecutionRepo.Save(ctx, execution)
		return &application.Result{Success: false, Error: "no executor found"}, fmt.Errorf("no executor found")
	}
	
	// 根据执行模式处理
	switch tool.ExecutionMode {
	case domain.ExecutionModeSync:
		return s.executeSyncTool(ctx, tool, agent, execution, executor)
	case domain.ExecutionModeAsync:
		return s.executeAsyncTool(ctx, tool, agent, execution, executor)
	default:
		return &application.Result{Success: false, Error: "unsupported execution mode"}, fmt.Errorf("unsupported execution mode")
	}
}

// executeSyncTool 同步执行工具
func (s *AgentService) executeSyncTool(ctx context.Context, tool *domain.Tool, agent *domain.Agent, execution *domain.ToolExecution, executor ToolExecutor) (*application.Result, error) {
	startTime := time.Now()
	
	// 执行工具
	result, err := executor.Execute(ctx, &ToolExecutionRequest{
		Tool:      tool,
		Agent:     agent,
		Input:     execution.Input,
		Context:   execution.Context,
	})
	
	duration := time.Since(startTime)
	
	if err != nil {
		// 执行失败
		execution.Fail(err.Error(), duration)
		tool.RecordUsage(duration, false)
		
		s.toolExecutionRepo.Save(ctx, execution)
		s.toolRepo.Save(ctx, tool)
		
		return &application.Result{Success: false, Error: err.Error()}, err
	}
	
	// 执行成功
	execution.Complete(result.Output, duration)
		tool.RecordUsage(duration, true)
		
		s.toolExecutionRepo.Save(ctx, execution)
		s.toolRepo.Save(ctx, tool)
		
		// 记录工具使用指标
		if s.metrics != nil {
			// 可以添加更多工具使用指标，比如按类型、按智能体等
			s.logger.Info("Tool executed successfully", 
				zap.String("tool_name", tool.Name),
				zap.String("tool_type", string(tool.Type)),
				zap.Duration("duration", duration),
			)
		}
		
		// 让智能体学习执行结果
		if result.ShouldLearn {
			knowledge := fmt.Sprintf("Used tool %s with result: %v", tool.Name, result.Output)
			agent.Learn(knowledge, 0.5)
			s.agentRepo.Save(ctx, agent)
		}
		
		return &application.Result{Success: true, Data: execution}, nil
}

// executeAsyncTool 异步执行工具
func (s *AgentService) executeAsyncTool(ctx context.Context, tool *domain.Tool, agent *domain.Agent, execution *domain.ToolExecution, executor ToolExecutor) (*application.Result, error) {
	// 异步执行
	go func() {
		defer func() {
			if r := recover(); r != nil {
				s.logger.Error("Panic in executeAsyncTool", zap.Any("panic", r))
				execution.Fail(fmt.Sprintf("panic: %v", r), 0)
				s.toolExecutionRepo.Save(context.Background(), execution)
			}
		}()
		
		startTime := time.Now()
		
		result, err := executor.Execute(context.Background(), &ToolExecutionRequest{
			Tool:    tool,
			Agent:   agent,
			Input:   execution.Input,
			Context: execution.Context,
		})
		
		duration := time.Since(startTime)
		
		if err != nil {
			execution.Fail(err.Error(), duration)
			tool.RecordUsage(duration, false)
		} else {
			execution.Complete(result.Output, duration)
			tool.RecordUsage(duration, true)
			
			// 让智能体学习
			if result.ShouldLearn {
				knowledge := fmt.Sprintf("Used tool %s with result: %v", tool.Name, result.Output)
				agent.Learn(knowledge, 0.5)
				s.agentRepo.Save(context.Background(), agent)
			}
		}
		
		s.toolExecutionRepo.Save(context.Background(), execution)
		s.toolRepo.Save(context.Background(), tool)
		
		// 发布完成事件
		if s.eventBus != nil {
			event := map[string]interface{}{
				"execution_id": execution.ID,
				"agent_id":     agent.ID,
				"tool_id":      tool.ID,
				"status":       execution.Status,
			}
			s.eventBus.Publish(context.Background(), 
				&application.BaseDomainEvent{
					EventType:   "tool.execution.completed",
					AggregateID: execution.ID,
					EventData:   event,
				})
		}
	}()
	
	return &application.Result{Success: true, Data: map[string]interface{}{
		"execution_id": execution.ID,
		"status":       "async_started",
	}}, nil
}

// ChatWithAgent 与智能体对话
func (s *AgentService) ChatWithAgent(ctx context.Context, cmd *ChatCommand) (*application.Result, error) {
	if err := cmd.Validate(); err != nil {
		return &application.Result{Success: false, Error: err.Error()}, err
	}
	
	// 获取智能体
	agent, err := s.agentRepo.FindByID(ctx, cmd.AgentID)
	if err != nil {
		return &application.Result{Success: false, Error: "agent not found"}, err
	}
	
	// 更新智能体状态
	agent.ChangeStatus(domain.AgentStatusBusy)
	
	// 将对话添加到记忆中
	if agent.Memory != nil {
		conversationMemory := domain.NewMemory(
			fmt.Sprintf("User: %s", cmd.Message),
			domain.MemoryTypeConversation,
			0.7,
		)
		agent.Memory.AddMemory(conversationMemory)
	}
	
	// TODO: 实现与大模型的对话逻辑
	// 这里应该调用LLM服务进行对话处理
	
	response := "这是一个模拟回复" // 临时回复
	
	// 将回复添加到记忆中
	if agent.Memory != nil {
		responseMemory := domain.NewMemory(
			fmt.Sprintf("Assistant: %s", response),
			domain.MemoryTypeConversation,
			0.7,
		)
		agent.Memory.AddMemory(responseMemory)
	}
	
	// 更新智能体状态
	agent.ChangeStatus(domain.AgentStatusIdle)
	
	// 保存智能体
	if err := s.agentRepo.Save(ctx, agent); err != nil {
		s.logger.Error("Failed to save agent", zap.Error(err))
	}
	
	return &application.Result{Success: true, Data: map[string]interface{}{
		"response": response,
		"agent_id": agent.ID,
	}}, nil
}

// ToolExecutor 工具执行器接口
type ToolExecutor interface {
	Execute(ctx context.Context, request *ToolExecutionRequest) (*ToolExecutionResult, error)
	GetSupportedType() domain.ToolType
}

// ToolExecutionRequest 工具执行请求
type ToolExecutionRequest struct {
	Tool    *domain.Tool
	Agent   *domain.Agent
	Input   map[string]interface{}
	Context map[string]interface{}
}

// ToolExecutionResult 工具执行结果
type ToolExecutionResult struct {
	Output      map[string]interface{}
	ShouldLearn bool                   // 是否应该让智能体学习这个结果
	Metadata    map[string]interface{}
}
