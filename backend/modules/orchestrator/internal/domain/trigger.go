package domain

import (
	"time"
	
	"github.com/google/uuid"
	"github.com/noah-loop/backend/shared/pkg/domain"
)

// Trigger 触发器实体
type Trigger struct {
	domain.BaseEntity
	WorkflowID  uuid.UUID              `json:"workflow_id" gorm:"type:uuid;not null;index"`
	Type        TriggerType            `json:"type" gorm:"not null"`
	Name        string                 `json:"name" gorm:"not null"`
	Description string                 `json:"description"`
	Config      map[string]interface{} `json:"config" gorm:"type:jsonb"`
	IsEnabled   bool                   `json:"is_enabled" gorm:"default:true"`
	
	// 调度配置（针对定时触发）
	Schedule     string     `json:"schedule"`      // Cron表达式
	Timezone     string     `json:"timezone"`      // 时区
	NextRun      *time.Time `json:"next_run"`      // 下次运行时间
	
	// 条件配置（针对条件触发）
	Conditions   []TriggerCondition `json:"conditions" gorm:"type:jsonb"`
	
	// 统计信息
	TriggerCount int       `json:"trigger_count" gorm:"default:0"`
	LastTriggered *time.Time `json:"last_triggered"`
	
	// 关联
	Workflow *Workflow `json:"workflow,omitempty" gorm:"foreignKey:WorkflowID"`
	
	// 聚合根实现
	domainEvents []domain.DomainEvent `gorm:"-"`
}

// TriggerCondition 触发条件
type TriggerCondition struct {
	Field    string      `json:"field"`
	Operator string      `json:"operator"` // eq, ne, gt, lt, gte, lte, in, contains
	Value    interface{} `json:"value"`
}

func (t *Trigger) GetID() uuid.UUID {
	return t.ID
}

func (t *Trigger) GetVersion() int {
	return t.Version
}

func (t *Trigger) MarkAsModified() {
	t.UpdatedAt = time.Now()
}

// NewTrigger 创建新触发器
func NewTrigger(workflowID uuid.UUID, triggerType TriggerType, name string) *Trigger {
	trigger := &Trigger{
		BaseEntity: domain.BaseEntity{
			ID:        domain.NewEntityID(),
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		},
		WorkflowID:   workflowID,
		Type:         triggerType,
		Name:         name,
		Config:       make(map[string]interface{}),
		IsEnabled:    true,
		Conditions:   make([]TriggerCondition, 0),
		TriggerCount: 0,
		domainEvents: make([]domain.DomainEvent, 0),
	}
	
	// 发布触发器创建事件
	event := domain.NewDomainEvent("trigger.created", trigger.ID, trigger)
	trigger.domainEvents = append(trigger.domainEvents, event)
	
	return trigger
}

// Enable 启用触发器
func (t *Trigger) Enable() {
	if !t.IsEnabled {
		t.IsEnabled = true
		t.MarkAsModified()
		
		event := domain.NewDomainEvent("trigger.enabled", t.ID, t.ID)
		t.domainEvents = append(t.domainEvents, event)
	}
}

// Disable 禁用触发器
func (t *Trigger) Disable() {
	if t.IsEnabled {
		t.IsEnabled = false
		t.MarkAsModified()
		
		event := domain.NewDomainEvent("trigger.disabled", t.ID, t.ID)
		t.domainEvents = append(t.domainEvents, event)
	}
}

// Trigger 执行触发
func (t *Trigger) Fire() {
	t.TriggerCount++
	now := time.Now()
	t.LastTriggered = &now
	t.MarkAsModified()
	
	event := domain.NewDomainEvent("trigger.fired", t.ID, map[string]interface{}{
		"trigger_id":     t.ID,
		"workflow_id":    t.WorkflowID,
		"trigger_count":  t.TriggerCount,
		"triggered_at":   now,
	})
	t.domainEvents = append(t.domainEvents, event)
}

// UpdateSchedule 更新调度配置
func (t *Trigger) UpdateSchedule(schedule, timezone string) error {
	if t.Type != TriggerTypeSchedule {
		return NewTriggerError("can only update schedule for schedule triggers")
	}
	
	t.Schedule = schedule
	t.Timezone = timezone
	// TODO: 计算下次运行时间
	t.MarkAsModified()
	
	event := domain.NewDomainEvent("trigger.schedule.updated", t.ID, map[string]interface{}{
		"trigger_id": t.ID,
		"schedule":   schedule,
		"timezone":   timezone,
	})
	t.domainEvents = append(t.domainEvents, event)
	
	return nil
}

// UpdateConditions 更新条件
func (t *Trigger) UpdateConditions(conditions []TriggerCondition) error {
	if t.Type != TriggerTypeCondition {
		return NewTriggerError("can only update conditions for condition triggers")
	}
	
	t.Conditions = conditions
	t.MarkAsModified()
	
	event := domain.NewDomainEvent("trigger.conditions.updated", t.ID, map[string]interface{}{
		"trigger_id": t.ID,
		"conditions": conditions,
	})
	t.domainEvents = append(t.domainEvents, event)
	
	return nil
}

// CheckConditions 检查条件是否满足
func (t *Trigger) CheckConditions(data map[string]interface{}) bool {
	if t.Type != TriggerTypeCondition {
		return false
	}
	
	for _, condition := range t.Conditions {
		if !evaluateCondition(condition, data) {
			return false
		}
	}
	
	return len(t.Conditions) > 0
}

// GetDomainEvents 获取领域事件
func (t *Trigger) GetDomainEvents() []domain.DomainEvent {
	return t.domainEvents
}

// ClearDomainEvents 清理领域事件
func (t *Trigger) ClearDomainEvents() {
	t.domainEvents = make([]domain.DomainEvent, 0)
}

// evaluateCondition 评估条件
func evaluateCondition(condition TriggerCondition, data map[string]interface{}) bool {
	fieldValue, exists := data[condition.Field]
	if !exists {
		return false
	}
	
	switch condition.Operator {
	case "eq":
		return fieldValue == condition.Value
	case "ne":
		return fieldValue != condition.Value
	// TODO: 实现更多操作符
	default:
		return false
	}
}

// TriggerError 触发器错误
type TriggerError struct {
	message string
}

func NewTriggerError(message string) *TriggerError {
	return &TriggerError{message: message}
}

func (e *TriggerError) Error() string {
	return e.message
}

// TriggerRepository 触发器仓储接口
type TriggerRepository interface {
	domain.Repository[*Trigger]
	FindByWorkflowID(ctx context.Context, workflowID uuid.UUID) ([]*Trigger, error)
	FindByType(ctx context.Context, triggerType TriggerType) ([]*Trigger, error)
	FindEnabledTriggers(ctx context.Context) ([]*Trigger, error)
	FindScheduledTriggers(ctx context.Context, before time.Time) ([]*Trigger, error)
}
