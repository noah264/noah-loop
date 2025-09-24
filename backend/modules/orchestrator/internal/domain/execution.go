package domain

import (
	"time"
	
	"github.com/google/uuid"
	"github.com/noah-loop/backend/shared/pkg/domain"
)

// ExecutionStatus 执行状态
type ExecutionStatus string

const (
	ExecutionStatusPending   ExecutionStatus = "pending"   // 待执行
	ExecutionStatusRunning   ExecutionStatus = "running"   // 执行中
	ExecutionStatusCompleted ExecutionStatus = "completed" // 已完成
	ExecutionStatusFailed    ExecutionStatus = "failed"    // 失败
	ExecutionStatusCancelled ExecutionStatus = "cancelled" // 取消
	ExecutionStatusTimeout   ExecutionStatus = "timeout"   // 超时
)

// Execution 工作流执行实体
type Execution struct {
	domain.BaseEntity
	WorkflowID   uuid.UUID              `json:"workflow_id" gorm:"type:uuid;not null;index"`
	TriggerID    uuid.UUID              `json:"trigger_id" gorm:"type:uuid;index"`
	Status       ExecutionStatus        `json:"status" gorm:"not null;default:'pending'"`
	Input        map[string]interface{} `json:"input" gorm:"type:jsonb"`
	Output       map[string]interface{} `json:"output" gorm:"type:jsonb"`
	Context      map[string]interface{} `json:"context" gorm:"type:jsonb"`
	ErrorMessage string                 `json:"error_message"`
	
	// 执行时间
	StartedAt   *time.Time    `json:"started_at"`
	CompletedAt *time.Time    `json:"completed_at"`
	Duration    time.Duration `json:"duration"`
	
	// 步骤执行
	StepExecutions []*StepExecution `json:"step_executions" gorm:"foreignKey:ExecutionID"`
	CurrentStep    *uuid.UUID       `json:"current_step" gorm:"type:uuid"`
	
	// 关联
	Workflow *Workflow `json:"workflow,omitempty" gorm:"foreignKey:WorkflowID"`
	Trigger  *Trigger  `json:"trigger,omitempty" gorm:"foreignKey:TriggerID"`
	
	// 聚合根实现
	domainEvents []domain.DomainEvent `gorm:"-"`
}

func (e *Execution) GetID() uuid.UUID {
	return e.ID
}

func (e *Execution) GetVersion() int {
	return e.Version
}

func (e *Execution) MarkAsModified() {
	e.UpdatedAt = time.Now()
}

// NewExecution 创建新执行
func NewExecution(workflowID, triggerID uuid.UUID, input map[string]interface{}) *Execution {
	execution := &Execution{
		BaseEntity: domain.BaseEntity{
			ID:        domain.NewEntityID(),
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		},
		WorkflowID:     workflowID,
		TriggerID:      triggerID,
		Status:         ExecutionStatusPending,
		Input:          input,
		Output:         make(map[string]interface{}),
		Context:        make(map[string]interface{}),
		StepExecutions: make([]*StepExecution, 0),
		domainEvents:   make([]domain.DomainEvent, 0),
	}
	
	// 发布执行创建事件
	event := domain.NewDomainEvent("execution.created", execution.ID, execution)
	execution.domainEvents = append(execution.domainEvents, event)
	
	return execution
}

// Start 开始执行
func (e *Execution) Start() error {
	if e.Status != ExecutionStatusPending {
		return NewExecutionError("execution is not in pending status")
	}
	
	e.Status = ExecutionStatusRunning
	now := time.Now()
	e.StartedAt = &now
	e.MarkAsModified()
	
	event := domain.NewDomainEvent("execution.started", e.ID, map[string]interface{}{
		"execution_id": e.ID,
		"workflow_id":  e.WorkflowID,
		"started_at":   e.StartedAt,
	})
	e.domainEvents = append(e.domainEvents, event)
	
	return nil
}

// Complete 完成执行
func (e *Execution) Complete(output map[string]interface{}) error {
	if e.Status != ExecutionStatusRunning {
		return NewExecutionError("execution is not in running status")
	}
	
	e.Status = ExecutionStatusCompleted
	e.Output = output
	now := time.Now()
	e.CompletedAt = &now
	
	if e.StartedAt != nil {
		e.Duration = now.Sub(*e.StartedAt)
	}
	
	e.MarkAsModified()
	
	event := domain.NewDomainEvent("execution.completed", e.ID, map[string]interface{}{
		"execution_id": e.ID,
		"workflow_id":  e.WorkflowID,
		"completed_at": e.CompletedAt,
		"duration":     e.Duration,
		"output":       output,
	})
	e.domainEvents = append(e.domainEvents, event)
	
	return nil
}

// Fail 执行失败
func (e *Execution) Fail(errorMessage string) error {
	if e.Status != ExecutionStatusRunning {
		return NewExecutionError("execution is not in running status")
	}
	
	e.Status = ExecutionStatusFailed
	e.ErrorMessage = errorMessage
	now := time.Now()
	e.CompletedAt = &now
	
	if e.StartedAt != nil {
		e.Duration = now.Sub(*e.StartedAt)
	}
	
	e.MarkAsModified()
	
	event := domain.NewDomainEvent("execution.failed", e.ID, map[string]interface{}{
		"execution_id": e.ID,
		"workflow_id":  e.WorkflowID,
		"error":        errorMessage,
		"completed_at": e.CompletedAt,
		"duration":     e.Duration,
	})
	e.domainEvents = append(e.domainEvents, event)
	
	return nil
}

// Cancel 取消执行
func (e *Execution) Cancel() {
	if e.Status == ExecutionStatusCompleted || e.Status == ExecutionStatusFailed {
		return
	}
	
	e.Status = ExecutionStatusCancelled
	now := time.Now()
	e.CompletedAt = &now
	
	if e.StartedAt != nil {
		e.Duration = now.Sub(*e.StartedAt)
	}
	
	e.MarkAsModified()
	
	event := domain.NewDomainEvent("execution.cancelled", e.ID, map[string]interface{}{
		"execution_id": e.ID,
		"workflow_id":  e.WorkflowID,
	})
	e.domainEvents = append(e.domainEvents, event)
}

// SetCurrentStep 设置当前步骤
func (e *Execution) SetCurrentStep(stepID uuid.UUID) {
	e.CurrentStep = &stepID
	e.MarkAsModified()
	
	event := domain.NewDomainEvent("execution.step.current", e.ID, map[string]interface{}{
		"execution_id": e.ID,
		"step_id":      stepID,
	})
	e.domainEvents = append(e.domainEvents, event)
}

// AddStepExecution 添加步骤执行
func (e *Execution) AddStepExecution(stepExecution *StepExecution) {
	stepExecution.ExecutionID = e.ID
	e.StepExecutions = append(e.StepExecutions, stepExecution)
	e.MarkAsModified()
}

// GetDomainEvents 获取领域事件
func (e *Execution) GetDomainEvents() []domain.DomainEvent {
	return e.domainEvents
}

// ClearDomainEvents 清理领域事件
func (e *Execution) ClearDomainEvents() {
	e.domainEvents = make([]domain.DomainEvent, 0)
}

// StepExecution 步骤执行记录
type StepExecution struct {
	domain.BaseEntity
	ExecutionID  uuid.UUID              `json:"execution_id" gorm:"type:uuid;not null;index"`
	StepID       uuid.UUID              `json:"step_id" gorm:"type:uuid;not null;index"`
	Status       StepStatus             `json:"status" gorm:"not null;default:'pending'"`
	Input        map[string]interface{} `json:"input" gorm:"type:jsonb"`
	Output       map[string]interface{} `json:"output" gorm:"type:jsonb"`
	ErrorMessage string                 `json:"error_message"`
	StartedAt    *time.Time             `json:"started_at"`
	CompletedAt  *time.Time             `json:"completed_at"`
	Duration     time.Duration          `json:"duration"`
	RetryCount   int                    `json:"retry_count" gorm:"default:0"`
	
	// 关联
	Execution *Execution `json:"execution,omitempty" gorm:"foreignKey:ExecutionID"`
	Step      *Step      `json:"step,omitempty" gorm:"foreignKey:StepID"`
}

// NewStepExecution 创建新步骤执行
func NewStepExecution(executionID, stepID uuid.UUID, input map[string]interface{}) *StepExecution {
	return &StepExecution{
		BaseEntity: domain.BaseEntity{
			ID:        domain.NewEntityID(),
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		},
		ExecutionID: executionID,
		StepID:      stepID,
		Status:      StepStatusPending,
		Input:       input,
		Output:      make(map[string]interface{}),
		RetryCount:  0,
	}
}

// ExecutionError 执行错误
type ExecutionError struct {
	message string
}

func NewExecutionError(message string) *ExecutionError {
	return &ExecutionError{message: message}
}

func (e *ExecutionError) Error() string {
	return e.message
}

// ExecutionRepository 执行仓储接口
type ExecutionRepository interface {
	domain.Repository[*Execution]
	FindByWorkflowID(ctx context.Context, workflowID uuid.UUID, offset, limit int) ([]*Execution, error)
	FindByStatus(ctx context.Context, status ExecutionStatus) ([]*Execution, error)
	FindRunningExecutions(ctx context.Context) ([]*Execution, error)
	FindByTriggerID(ctx context.Context, triggerID uuid.UUID) ([]*Execution, error)
}

// StepExecutionRepository 步骤执行仓储接口
type StepExecutionRepository interface {
	domain.Repository[*StepExecution]
	FindByExecutionID(ctx context.Context, executionID uuid.UUID) ([]*StepExecution, error)
	FindByStepID(ctx context.Context, stepID uuid.UUID) ([]*StepExecution, error)
}
