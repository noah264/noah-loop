package domain

import (
	"time"
	
	"github.com/google/uuid"
	"github.com/noah-loop/backend/shared/pkg/domain"
)

// WorkflowStatus 工作流状态
type WorkflowStatus string

const (
	WorkflowStatusDraft      WorkflowStatus = "draft"      // 草稿
	WorkflowStatusActive     WorkflowStatus = "active"     // 活跃
	WorkflowStatusPaused     WorkflowStatus = "paused"     // 暂停
	WorkflowStatusCompleted  WorkflowStatus = "completed"  // 已完成
	WorkflowStatusFailed     WorkflowStatus = "failed"     // 失败
	WorkflowStatusCancelled  WorkflowStatus = "cancelled"  // 取消
)

// TriggerType 触发器类型
type TriggerType string

const (
	TriggerTypeManual     TriggerType = "manual"     // 手动触发
	TriggerTypeSchedule   TriggerType = "schedule"   // 定时触发
	TriggerTypeEvent      TriggerType = "event"      // 事件触发
	TriggerTypeWebhook    TriggerType = "webhook"    // Webhook触发
	TriggerTypeCondition  TriggerType = "condition"  // 条件触发
)

// Workflow 工作流实体
type Workflow struct {
	domain.BaseEntity
	Name        string                 `json:"name" gorm:"not null;index"`
	Description string                 `json:"description"`
	Status      WorkflowStatus         `json:"status" gorm:"not null;default:'draft'"`
	Definition  map[string]interface{} `json:"definition" gorm:"type:jsonb;not null"`
	Triggers    []*Trigger            `json:"triggers" gorm:"foreignKey:WorkflowID"`
	Steps       []*Step               `json:"steps" gorm:"foreignKey:WorkflowID"`
	Variables   map[string]interface{} `json:"variables" gorm:"type:jsonb"`
	Tags        []string              `json:"tags" gorm:"type:text[]"`
	OwnerID     uuid.UUID             `json:"owner_id" gorm:"type:uuid;not null;index"`
	IsTemplate  bool                  `json:"is_template" gorm:"default:false"`
	
	// 统计信息
	ExecutionCount int       `json:"execution_count" gorm:"default:0"`
	LastExecuted   time.Time `json:"last_executed"`
	SuccessRate    float64   `json:"success_rate" gorm:"default:0"`
	
	// 聚合根实现
	domainEvents []domain.DomainEvent `gorm:"-"`
}

func (w *Workflow) GetID() uuid.UUID {
	return w.ID
}

func (w *Workflow) GetVersion() int {
	return w.Version
}

func (w *Workflow) MarkAsModified() {
	w.UpdatedAt = time.Now()
}

// NewWorkflow 创建新工作流
func NewWorkflow(name, description string, ownerID uuid.UUID) *Workflow {
	workflow := &Workflow{
		BaseEntity: domain.BaseEntity{
			ID:        domain.NewEntityID(),
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		},
		Name:        name,
		Description: description,
		Status:      WorkflowStatusDraft,
		Definition:  make(map[string]interface{}),
		Variables:   make(map[string]interface{}),
		Tags:        make([]string, 0),
		OwnerID:     ownerID,
		IsTemplate:  false,
		ExecutionCount: 0,
		SuccessRate: 0,
		domainEvents: make([]domain.DomainEvent, 0),
	}
	
	// 发布工作流创建事件
	event := domain.NewDomainEvent("workflow.created", workflow.ID, workflow)
	workflow.domainEvents = append(workflow.domainEvents, event)
	
	return workflow
}

// Activate 激活工作流
func (w *Workflow) Activate() error {
	if w.Status == WorkflowStatusActive {
		return nil
	}
	
	if len(w.Steps) == 0 {
		return NewWorkflowError("cannot activate workflow without steps")
	}
	
	oldStatus := w.Status
	w.Status = WorkflowStatusActive
	w.MarkAsModified()
	
	event := domain.NewDomainEvent("workflow.activated", w.ID, map[string]interface{}{
		"workflow_id": w.ID,
		"old_status":  oldStatus,
		"new_status":  WorkflowStatusActive,
	})
	w.domainEvents = append(w.domainEvents, event)
	
	return nil
}

// Pause 暂停工作流
func (w *Workflow) Pause() error {
	if w.Status != WorkflowStatusActive {
		return NewWorkflowError("can only pause active workflow")
	}
	
	w.Status = WorkflowStatusPaused
	w.MarkAsModified()
	
	event := domain.NewDomainEvent("workflow.paused", w.ID, w.ID)
	w.domainEvents = append(w.domainEvents, event)
	
	return nil
}

// Resume 恢复工作流
func (w *Workflow) Resume() error {
	if w.Status != WorkflowStatusPaused {
		return NewWorkflowError("can only resume paused workflow")
	}
	
	w.Status = WorkflowStatusActive
	w.MarkAsModified()
	
	event := domain.NewDomainEvent("workflow.resumed", w.ID, w.ID)
	w.domainEvents = append(w.domainEvents, event)
	
	return nil
}

// Complete 完成工作流
func (w *Workflow) Complete() {
	w.Status = WorkflowStatusCompleted
	w.MarkAsModified()
	
	event := domain.NewDomainEvent("workflow.completed", w.ID, w.ID)
	w.domainEvents = append(w.domainEvents, event)
}

// Fail 工作流失败
func (w *Workflow) Fail(reason string) {
	w.Status = WorkflowStatusFailed
	w.MarkAsModified()
	
	event := domain.NewDomainEvent("workflow.failed", w.ID, map[string]interface{}{
		"workflow_id": w.ID,
		"reason":      reason,
	})
	w.domainEvents = append(w.domainEvents, event)
}

// Cancel 取消工作流
func (w *Workflow) Cancel() {
	if w.Status == WorkflowStatusCompleted || w.Status == WorkflowStatusFailed {
		return
	}
	
	w.Status = WorkflowStatusCancelled
	w.MarkAsModified()
	
	event := domain.NewDomainEvent("workflow.cancelled", w.ID, w.ID)
	w.domainEvents = append(w.domainEvents, event)
}

// UpdateDefinition 更新工作流定义
func (w *Workflow) UpdateDefinition(definition map[string]interface{}) error {
	if w.Status == WorkflowStatusActive {
		return NewWorkflowError("cannot update definition of active workflow")
	}
	
	w.Definition = definition
	w.MarkAsModified()
	
	event := domain.NewDomainEvent("workflow.definition.updated", w.ID, map[string]interface{}{
		"workflow_id": w.ID,
		"definition":  definition,
	})
	w.domainEvents = append(w.domainEvents, event)
	
	return nil
}

// AddStep 添加步骤
func (w *Workflow) AddStep(step *Step) error {
	if w.Status == WorkflowStatusActive {
		return NewWorkflowError("cannot add step to active workflow")
	}
	
	step.WorkflowID = w.ID
	w.Steps = append(w.Steps, step)
	w.MarkAsModified()
	
	event := domain.NewDomainEvent("workflow.step.added", w.ID, map[string]interface{}{
		"workflow_id": w.ID,
		"step_id":     step.ID,
		"step_name":   step.Name,
	})
	w.domainEvents = append(w.domainEvents, event)
	
	return nil
}

// AddTrigger 添加触发器
func (w *Workflow) AddTrigger(trigger *Trigger) error {
	trigger.WorkflowID = w.ID
	w.Triggers = append(w.Triggers, trigger)
	w.MarkAsModified()
	
	event := domain.NewDomainEvent("workflow.trigger.added", w.ID, map[string]interface{}{
		"workflow_id": w.ID,
		"trigger_id":  trigger.ID,
		"trigger_type": trigger.Type,
	})
	w.domainEvents = append(w.domainEvents, event)
	
	return nil
}

// RecordExecution 记录执行
func (w *Workflow) RecordExecution(success bool) {
	w.ExecutionCount++
	w.LastExecuted = time.Now()
	
	// 更新成功率
	if success {
		w.SuccessRate = (w.SuccessRate*float64(w.ExecutionCount-1) + 1.0) / float64(w.ExecutionCount)
	} else {
		w.SuccessRate = (w.SuccessRate * float64(w.ExecutionCount-1)) / float64(w.ExecutionCount)
	}
	
	w.MarkAsModified()
	
	event := domain.NewDomainEvent("workflow.execution.recorded", w.ID, map[string]interface{}{
		"workflow_id":     w.ID,
		"execution_count": w.ExecutionCount,
		"success":         success,
		"success_rate":    w.SuccessRate,
	})
	w.domainEvents = append(w.domainEvents, event)
}

// GetDomainEvents 获取领域事件
func (w *Workflow) GetDomainEvents() []domain.DomainEvent {
	return w.domainEvents
}

// ClearDomainEvents 清理领域事件
func (w *Workflow) ClearDomainEvents() {
	w.domainEvents = make([]domain.DomainEvent, 0)
}

// WorkflowError 工作流错误
type WorkflowError struct {
	message string
}

func NewWorkflowError(message string) *WorkflowError {
	return &WorkflowError{message: message}
}

func (e *WorkflowError) Error() string {
	return e.message
}

// WorkflowRepository 工作流仓储接口
type WorkflowRepository interface {
	domain.Repository[*Workflow]
	FindByOwnerID(ctx context.Context, ownerID uuid.UUID) ([]*Workflow, error)
	FindByStatus(ctx context.Context, status WorkflowStatus) ([]*Workflow, error)
	FindActiveWorkflows(ctx context.Context) ([]*Workflow, error)
	FindByTags(ctx context.Context, tags []string) ([]*Workflow, error)
	FindTemplates(ctx context.Context) ([]*Workflow, error)
}
