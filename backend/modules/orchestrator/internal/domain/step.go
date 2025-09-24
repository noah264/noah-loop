package domain

import (
	"time"
	
	"github.com/google/uuid"
	"github.com/noah-loop/backend/shared/pkg/domain"
)

// StepType 步骤类型
type StepType string

const (
	StepTypeAction      StepType = "action"      // 动作步骤
	StepTypeCondition   StepType = "condition"   // 条件步骤
	StepTypeLoop        StepType = "loop"        // 循环步骤
	StepTypeParallel    StepType = "parallel"    // 并行步骤
	StepTypeWait        StepType = "wait"        // 等待步骤
	StepTypeHuman       StepType = "human"       // 人工步骤
	StepTypeSubworkflow StepType = "subworkflow" // 子工作流步骤
)

// StepStatus 步骤状态
type StepStatus string

const (
	StepStatusPending    StepStatus = "pending"    // 待执行
	StepStatusRunning    StepStatus = "running"    // 执行中
	StepStatusCompleted  StepStatus = "completed"  // 已完成
	StepStatusFailed     StepStatus = "failed"     // 失败
	StepStatusSkipped    StepStatus = "skipped"    // 跳过
	StepStatusTimeout    StepStatus = "timeout"    // 超时
	StepStatusCancelled  StepStatus = "cancelled"  // 取消
)

// Step 步骤实体
type Step struct {
	domain.BaseEntity
	WorkflowID   uuid.UUID              `json:"workflow_id" gorm:"type:uuid;not null;index"`
	Name         string                 `json:"name" gorm:"not null"`
	Type         StepType               `json:"type" gorm:"not null"`
	Status       StepStatus             `json:"status" gorm:"not null;default:'pending'"`
	Description  string                 `json:"description"`
	Config       map[string]interface{} `json:"config" gorm:"type:jsonb"`
	Input        map[string]interface{} `json:"input" gorm:"type:jsonb"`
	Output       map[string]interface{} `json:"output" gorm:"type:jsonb"`
	ErrorMessage string                 `json:"error_message"`
	
	// 执行配置
	Order        int           `json:"order" gorm:"not null;index"` // 执行顺序
	Timeout      time.Duration `json:"timeout"`                    // 超时时间
	RetryCount   int           `json:"retry_count" gorm:"default:0"`
	MaxRetries   int           `json:"max_retries" gorm:"default:3"`
	
	// 依赖关系
	Dependencies []uuid.UUID `json:"dependencies" gorm:"type:uuid[]"` // 依赖的步骤ID
	
	// 执行信息
	StartedAt   *time.Time    `json:"started_at"`
	CompletedAt *time.Time    `json:"completed_at"`
	Duration    time.Duration `json:"duration"`
	
	// 关联
	Workflow   *Workflow           `json:"workflow,omitempty" gorm:"foreignKey:WorkflowID"`
	Executions []*StepExecution    `json:"executions,omitempty" gorm:"foreignKey:StepID"`
	
	// 聚合根实现
	domainEvents []domain.DomainEvent `gorm:"-"`
}

func (s *Step) GetID() uuid.UUID {
	return s.ID
}

func (s *Step) GetVersion() int {
	return s.Version
}

func (s *Step) MarkAsModified() {
	s.UpdatedAt = time.Now()
}

// NewStep 创建新步骤
func NewStep(workflowID uuid.UUID, name string, stepType StepType, order int) *Step {
	step := &Step{
		BaseEntity: domain.BaseEntity{
			ID:        domain.NewEntityID(),
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		},
		WorkflowID:   workflowID,
		Name:         name,
		Type:         stepType,
		Status:       StepStatusPending,
		Config:       make(map[string]interface{}),
		Input:        make(map[string]interface{}),
		Output:       make(map[string]interface{}),
		Order:        order,
		Timeout:      30 * time.Minute, // 默认30分钟超时
		RetryCount:   0,
		MaxRetries:   3,
		Dependencies: make([]uuid.UUID, 0),
		domainEvents: make([]domain.DomainEvent, 0),
	}
	
	// 发布步骤创建事件
	event := domain.NewDomainEvent("step.created", step.ID, step)
	step.domainEvents = append(step.domainEvents, event)
	
	return step
}

// Start 开始执行步骤
func (s *Step) Start() error {
	if s.Status != StepStatusPending {
		return NewStepError("step is not in pending status")
	}
	
	s.Status = StepStatusRunning
	now := time.Now()
	s.StartedAt = &now
	s.MarkAsModified()
	
	event := domain.NewDomainEvent("step.started", s.ID, map[string]interface{}{
		"step_id":     s.ID,
		"workflow_id": s.WorkflowID,
		"started_at":  s.StartedAt,
	})
	s.domainEvents = append(s.domainEvents, event)
	
	return nil
}

// Complete 完成步骤
func (s *Step) Complete(output map[string]interface{}) error {
	if s.Status != StepStatusRunning {
		return NewStepError("step is not in running status")
	}
	
	s.Status = StepStatusCompleted
	s.Output = output
	now := time.Now()
	s.CompletedAt = &now
	
	if s.StartedAt != nil {
		s.Duration = now.Sub(*s.StartedAt)
	}
	
	s.MarkAsModified()
	
	event := domain.NewDomainEvent("step.completed", s.ID, map[string]interface{}{
		"step_id":      s.ID,
		"workflow_id":  s.WorkflowID,
		"completed_at": s.CompletedAt,
		"duration":     s.Duration,
		"output":       output,
	})
	s.domainEvents = append(s.domainEvents, event)
	
	return nil
}

// Fail 步骤失败
func (s *Step) Fail(errorMessage string) error {
	if s.Status != StepStatusRunning {
		return NewStepError("step is not in running status")
	}
	
	s.Status = StepStatusFailed
	s.ErrorMessage = errorMessage
	now := time.Now()
	s.CompletedAt = &now
	
	if s.StartedAt != nil {
		s.Duration = now.Sub(*s.StartedAt)
	}
	
	s.MarkAsModified()
	
	event := domain.NewDomainEvent("step.failed", s.ID, map[string]interface{}{
		"step_id":      s.ID,
		"workflow_id":  s.WorkflowID,
		"error":        errorMessage,
		"completed_at": s.CompletedAt,
		"duration":     s.Duration,
	})
	s.domainEvents = append(s.domainEvents, event)
	
	return nil
}

// Skip 跳过步骤
func (s *Step) Skip(reason string) {
	s.Status = StepStatusSkipped
	s.ErrorMessage = reason
	s.MarkAsModified()
	
	event := domain.NewDomainEvent("step.skipped", s.ID, map[string]interface{}{
		"step_id":     s.ID,
		"workflow_id": s.WorkflowID,
		"reason":      reason,
	})
	s.domainEvents = append(s.domainEvents, event)
}

// Timeout 步骤超时
func (s *Step) Timeout() {
	s.Status = StepStatusTimeout
	now := time.Now()
	s.CompletedAt = &now
	
	if s.StartedAt != nil {
		s.Duration = now.Sub(*s.StartedAt)
	}
	
	s.MarkAsModified()
	
	event := domain.NewDomainEvent("step.timeout", s.ID, map[string]interface{}{
		"step_id":     s.ID,
		"workflow_id": s.WorkflowID,
		"duration":    s.Duration,
	})
	s.domainEvents = append(s.domainEvents, event)
}

// Cancel 取消步骤
func (s *Step) Cancel() {
	if s.Status == StepStatusCompleted || s.Status == StepStatusFailed {
		return
	}
	
	s.Status = StepStatusCancelled
	now := time.Now()
	s.CompletedAt = &now
	
	if s.StartedAt != nil {
		s.Duration = now.Sub(*s.StartedAt)
	}
	
	s.MarkAsModified()
	
	event := domain.NewDomainEvent("step.cancelled", s.ID, map[string]interface{}{
		"step_id":     s.ID,
		"workflow_id": s.WorkflowID,
	})
	s.domainEvents = append(s.domainEvents, event)
}

// Retry 重试步骤
func (s *Step) Retry() error {
	if s.RetryCount >= s.MaxRetries {
		return NewStepError("maximum retry count exceeded")
	}
	
	s.RetryCount++
	s.Status = StepStatusPending
	s.ErrorMessage = ""
	s.StartedAt = nil
	s.CompletedAt = nil
	s.Duration = 0
	s.MarkAsModified()
	
	event := domain.NewDomainEvent("step.retried", s.ID, map[string]interface{}{
		"step_id":     s.ID,
		"workflow_id": s.WorkflowID,
		"retry_count": s.RetryCount,
	})
	s.domainEvents = append(s.domainEvents, event)
	
	return nil
}

// CanExecute 检查是否可以执行
func (s *Step) CanExecute(completedSteps []uuid.UUID) bool {
	if s.Status != StepStatusPending {
		return false
	}
	
	// 检查依赖是否满足
	for _, depID := range s.Dependencies {
		found := false
		for _, completedID := range completedSteps {
			if depID == completedID {
				found = true
				break
			}
		}
		if !found {
			return false
		}
	}
	
	return true
}

// GetDomainEvents 获取领域事件
func (s *Step) GetDomainEvents() []domain.DomainEvent {
	return s.domainEvents
}

// ClearDomainEvents 清理领域事件
func (s *Step) ClearDomainEvents() {
	s.domainEvents = make([]domain.DomainEvent, 0)
}

// StepError 步骤错误
type StepError struct {
	message string
}

func NewStepError(message string) *StepError {
	return &StepError{message: message}
}

func (e *StepError) Error() string {
	return e.message
}

// StepRepository 步骤仓储接口
type StepRepository interface {
	domain.Repository[*Step]
	FindByWorkflowID(ctx context.Context, workflowID uuid.UUID) ([]*Step, error)
	FindByStatus(ctx context.Context, status StepStatus) ([]*Step, error)
	FindExecutableSteps(ctx context.Context, workflowID uuid.UUID) ([]*Step, error)
}
