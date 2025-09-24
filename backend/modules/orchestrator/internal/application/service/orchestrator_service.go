package service

import (
	"context"
	"fmt"
	"sort"
	"time"
	
	"github.com/google/uuid"
	"github.com/noah-loop/backend/modules/orchestrator/internal/domain"
	"github.com/noah-loop/backend/shared/pkg/application"
	"github.com/noah-loop/backend/shared/pkg/infrastructure"
	"go.uber.org/zap"
)

// OrchestratorService 编排服务
type OrchestratorService struct {
	workflowRepo      domain.WorkflowRepository
	stepRepo          domain.StepRepository
	triggerRepo       domain.TriggerRepository
	executionRepo     domain.ExecutionRepository
	stepExecutionRepo domain.StepExecutionRepository
	eventBus          application.EventBus
	logger            infrastructure.Logger
	metrics           *infrastructure.MetricsRegistry
	stepExecutors     map[domain.StepType]StepExecutor
}

// NewOrchestratorService 创建编排服务
func NewOrchestratorService(
	workflowRepo domain.WorkflowRepository,
	stepRepo domain.StepRepository,
	triggerRepo domain.TriggerRepository,
	executionRepo domain.ExecutionRepository,
	stepExecutionRepo domain.StepExecutionRepository,
	eventBus application.EventBus,
	logger infrastructure.Logger,
	metrics *infrastructure.MetricsRegistry,
) *OrchestratorService {
	return &OrchestratorService{
		workflowRepo:      workflowRepo,
		stepRepo:          stepRepo,
		triggerRepo:       triggerRepo,
		executionRepo:     executionRepo,
		stepExecutionRepo: stepExecutionRepo,
		eventBus:          eventBus,
		logger:            logger,
		metrics:           metrics,
		stepExecutors:     make(map[domain.StepType]StepExecutor),
	}
}

// RegisterStepExecutor 注册步骤执行器
func (s *OrchestratorService) RegisterStepExecutor(stepType domain.StepType, executor StepExecutor) {
	s.stepExecutors[stepType] = executor
}

// CreateWorkflow 创建工作流
func (s *OrchestratorService) CreateWorkflow(ctx context.Context, cmd *CreateWorkflowCommand) (*application.Result, error) {
	if err := cmd.Validate(); err != nil {
		return &application.Result{Success: false, Error: err.Error()}, err
	}
	
	// 创建工作流
	workflow := domain.NewWorkflow(cmd.Name, cmd.Description, cmd.OwnerID)
	workflow.Definition = cmd.Definition
	workflow.Variables = cmd.Variables
	workflow.Tags = cmd.Tags
	workflow.IsTemplate = cmd.IsTemplate
	
	// 保存工作流
	if err := s.workflowRepo.Save(ctx, workflow); err != nil {
		s.logger.Error("Failed to save workflow", zap.Error(err))
		return &application.Result{Success: false, Error: "failed to save workflow"}, err
	}
	
	// 发布事件
	for _, event := range workflow.GetDomainEvents() {
		if err := s.eventBus.Publish(ctx, event); err != nil {
			s.logger.Warn("Failed to publish event", zap.Error(err))
		}
	}
	workflow.ClearDomainEvents()
	
	return &application.Result{Success: true, Data: workflow}, nil
}

// ExecuteWorkflow 执行工作流
func (s *OrchestratorService) ExecuteWorkflow(ctx context.Context, cmd *ExecuteWorkflowCommand) (*application.Result, error) {
	if err := cmd.Validate(); err != nil {
		return &application.Result{Success: false, Error: err.Error()}, err
	}
	
	// 获取工作流
	workflow, err := s.workflowRepo.FindByID(ctx, cmd.WorkflowID)
	if err != nil {
		return &application.Result{Success: false, Error: "workflow not found"}, err
	}
	
	// 检查工作流状态
	if workflow.Status != domain.WorkflowStatusActive {
		return &application.Result{Success: false, Error: "workflow is not active"}, fmt.Errorf("workflow is not active")
	}
	
	// 创建执行
	execution := domain.NewExecution(workflow.ID, cmd.TriggerID, cmd.Input)
	execution.Context = cmd.Context
	
	// 保存执行
	if err := s.executionRepo.Save(ctx, execution); err != nil {
		s.logger.Error("Failed to save execution", zap.Error(err))
		return &application.Result{Success: false, Error: "failed to save execution"}, err
	}
	
	// 异步执行工作流
	go s.executeWorkflowAsync(ctx, workflow, execution)
	
	// 记录工作流执行
	workflow.RecordExecution(true) // 先记录为成功，失败时会更新
	if err := s.workflowRepo.Save(ctx, workflow); err != nil {
		s.logger.Warn("Failed to update workflow execution stats", zap.Error(err))
	}
	
	return &application.Result{Success: true, Data: execution}, nil
}

// executeWorkflowAsync 异步执行工作流
func (s *OrchestratorService) executeWorkflowAsync(ctx context.Context, workflow *domain.Workflow, execution *domain.Execution) {
	defer func() {
		if r := recover(); r != nil {
			s.logger.Error("Panic in executeWorkflowAsync", zap.Any("panic", r))
			execution.Fail(fmt.Sprintf("internal error: %v", r))
			s.executionRepo.Save(ctx, execution)
		}
	}()
	
	// 开始执行
	if err := execution.Start(); err != nil {
		s.logger.Error("Failed to start execution", zap.Error(err))
		return
	}
	s.executionRepo.Save(ctx, execution)
	
	// 获取工作流步骤
	steps, err := s.stepRepo.FindByWorkflowID(ctx, workflow.ID)
	if err != nil {
		s.logger.Error("Failed to get workflow steps", zap.Error(err))
		execution.Fail("failed to get workflow steps")
		s.executionRepo.Save(ctx, execution)
		return
	}
	
	// 按顺序排序步骤
	sort.Slice(steps, func(i, j int) bool {
		return steps[i].Order < steps[j].Order
	})
	
	// 执行步骤
	completedSteps := make([]uuid.UUID, 0)
	
	for {
		// 找到可执行的步骤
		executableSteps := s.findExecutableSteps(steps, completedSteps)
		if len(executableSteps) == 0 {
			break // 没有可执行的步骤，结束执行
		}
		
		// 并行执行可执行的步骤
		stepResults := make(chan *stepExecutionResult, len(executableSteps))
		
		for _, step := range executableSteps {
			go s.executeStepAsync(ctx, execution, step, stepResults)
		}
		
		// 等待步骤执行完成
		for i := 0; i < len(executableSteps); i++ {
			result := <-stepResults
			if result.Success {
				completedSteps = append(completedSteps, result.StepID)
			} else {
				// 有步骤失败，整个工作流失败
				execution.Fail(fmt.Sprintf("step %s failed: %s", result.StepID, result.Error))
				s.executionRepo.Save(ctx, execution)
				return
			}
		}
	}
	
	// 检查是否所有步骤都执行完成
	if len(completedSteps) == len(steps) {
		// 所有步骤完成，工作流成功
		execution.Complete(map[string]interface{}{
			"completed_steps": completedSteps,
			"total_steps":     len(steps),
		})
		
		// 记录工作流执行成功指标
		if s.metrics != nil {
			duration := time.Since(*execution.StartedAt)
			s.metrics.RecordWorkflowExecution(workflow.ID.String(), "completed", duration)
		}
	} else {
		// 有未完成的步骤，可能存在循环依赖
		execution.Fail("workflow contains circular dependencies or unreachable steps")
		
		// 记录工作流执行失败指标
		if s.metrics != nil {
			duration := time.Since(*execution.StartedAt)
			s.metrics.RecordWorkflowExecution(workflow.ID.String(), "failed", duration)
		}
	}
	
	s.executionRepo.Save(ctx, execution)
}

// stepExecutionResult 步骤执行结果
type stepExecutionResult struct {
	StepID  uuid.UUID
	Success bool
	Error   string
	Output  map[string]interface{}
}

// executeStepAsync 异步执行步骤
func (s *OrchestratorService) executeStepAsync(ctx context.Context, execution *domain.Execution, step *domain.Step, result chan<- *stepExecutionResult) {
	defer func() {
		if r := recover(); r != nil {
			result <- &stepExecutionResult{
				StepID:  step.ID,
				Success: false,
				Error:   fmt.Sprintf("panic: %v", r),
			}
		}
	}()
	
	// 设置当前步骤
	execution.SetCurrentStep(step.ID)
	s.executionRepo.Save(ctx, execution)
	
	// 开始执行步骤
	if err := step.Start(); err != nil {
		s.logger.Error("Failed to start step", zap.Error(err))
		result <- &stepExecutionResult{
			StepID:  step.ID,
			Success: false,
			Error:   err.Error(),
		}
		return
	}
	s.stepRepo.Save(ctx, step)
	
	// 创建步骤执行记录
	stepExecution := domain.NewStepExecution(execution.ID, step.ID, step.Input)
	execution.AddStepExecution(stepExecution)
	s.stepExecutionRepo.Save(ctx, stepExecution)
	
	// 获取步骤执行器
	executor, exists := s.stepExecutors[step.Type]
	if !exists {
		step.Fail("no executor found for step type")
		s.stepRepo.Save(ctx, step)
		result <- &stepExecutionResult{
			StepID:  step.ID,
			Success: false,
			Error:   "no executor found",
		}
		return
	}
	
	// 执行步骤
	stepResult, err := executor.Execute(ctx, &StepExecutionRequest{
		Step:      step,
		Execution: execution,
		Input:     step.Input,
		Context:   execution.Context,
	})
	
	if err != nil {
		step.Fail(err.Error())
		s.stepRepo.Save(ctx, step)
		result <- &stepExecutionResult{
			StepID:  step.ID,
			Success: false,
			Error:   err.Error(),
		}
		return
	}
	
	// 步骤执行成功
	step.Complete(stepResult.Output)
	s.stepRepo.Save(ctx, step)
	
	result <- &stepExecutionResult{
		StepID:  step.ID,
		Success: true,
		Output:  stepResult.Output,
	}
}

// findExecutableSteps 找到可执行的步骤
func (s *OrchestratorService) findExecutableSteps(allSteps []*domain.Step, completedSteps []uuid.UUID) []*domain.Step {
	var executableSteps []*domain.Step
	
	for _, step := range allSteps {
		if step.CanExecute(completedSteps) {
			executableSteps = append(executableSteps, step)
		}
	}
	
	return executableSteps
}

// AddStep 添加步骤
func (s *OrchestratorService) AddStep(ctx context.Context, cmd *AddStepCommand) (*application.Result, error) {
	if err := cmd.Validate(); err != nil {
		return &application.Result{Success: false, Error: err.Error()}, err
	}
	
	// 获取工作流
	workflow, err := s.workflowRepo.FindByID(ctx, cmd.WorkflowID)
	if err != nil {
		return &application.Result{Success: false, Error: "workflow not found"}, err
	}
	
	// 创建步骤
	step := domain.NewStep(workflow.ID, cmd.Name, cmd.Type, cmd.Order)
	step.Description = cmd.Description
	step.Config = cmd.Config
	step.Input = cmd.Input
	step.Timeout = cmd.Timeout
	step.MaxRetries = cmd.MaxRetries
	step.Dependencies = cmd.Dependencies
	
	// 添加到工作流
	if err := workflow.AddStep(step); err != nil {
		return &application.Result{Success: false, Error: err.Error()}, err
	}
	
	// 保存步骤和工作流
	if err := s.stepRepo.Save(ctx, step); err != nil {
		s.logger.Error("Failed to save step", zap.Error(err))
		return &application.Result{Success: false, Error: "failed to save step"}, err
	}
	
	if err := s.workflowRepo.Save(ctx, workflow); err != nil {
		s.logger.Error("Failed to update workflow", zap.Error(err))
		return &application.Result{Success: false, Error: "failed to update workflow"}, err
	}
	
	// 发布事件
	for _, event := range step.GetDomainEvents() {
		if err := s.eventBus.Publish(ctx, event); err != nil {
			s.logger.Warn("Failed to publish event", zap.Error(err))
		}
	}
	step.ClearDomainEvents()
	
	for _, event := range workflow.GetDomainEvents() {
		if err := s.eventBus.Publish(ctx, event); err != nil {
			s.logger.Warn("Failed to publish event", zap.Error(err))
		}
	}
	workflow.ClearDomainEvents()
	
	return &application.Result{Success: true, Data: step}, nil
}

// StepExecutor 步骤执行器接口
type StepExecutor interface {
	Execute(ctx context.Context, request *StepExecutionRequest) (*StepExecutionResult, error)
	GetSupportedType() domain.StepType
}

// StepExecutionRequest 步骤执行请求
type StepExecutionRequest struct {
	Step      *domain.Step
	Execution *domain.Execution
	Input     map[string]interface{}
	Context   map[string]interface{}
}

// StepExecutionResult 步骤执行结果
type StepExecutionResult struct {
	Output   map[string]interface{}
	Metadata map[string]interface{}
}
