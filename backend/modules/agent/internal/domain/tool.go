package domain

import (
	"encoding/json"
	"time"
	
	"github.com/google/uuid"
	"github.com/noah-loop/backend/shared/pkg/domain"
)

// ToolType 工具类型
type ToolType string

const (
	ToolTypeFunction  ToolType = "function"  // 函数工具
	ToolTypeAPI       ToolType = "api"       // API工具
	ToolTypeDatabase  ToolType = "database"  // 数据库工具
	ToolTypeFile      ToolType = "file"      // 文件工具
	ToolTypeCalculator ToolType = "calculator" // 计算器工具
	ToolTypeWeb       ToolType = "web"       // 网络工具
	ToolTypeCustom    ToolType = "custom"    // 自定义工具
)

// ToolExecutionMode 工具执行模式
type ToolExecutionMode string

const (
	ExecutionModeSync  ToolExecutionMode = "sync"  // 同步执行
	ExecutionModeAsync ToolExecutionMode = "async" // 异步执行
	ExecutionModeStream ToolExecutionMode = "stream" // 流式执行
)

// Tool 工具实体
type Tool struct {
	domain.BaseEntity
	Name         string                 `json:"name" gorm:"not null;index"`
	Type         ToolType               `json:"type" gorm:"not null"`
	Description  string                 `json:"description"`
	Schema       map[string]interface{} `json:"schema" gorm:"type:jsonb"` // JSON Schema
	Config       map[string]interface{} `json:"config" gorm:"type:jsonb"`
	ExecutionMode ToolExecutionMode     `json:"execution_mode" gorm:"default:'sync'"`
	IsEnabled    bool                   `json:"is_enabled" gorm:"default:true"`
	IsPublic     bool                   `json:"is_public" gorm:"default:false"`
	OwnerID      uuid.UUID              `json:"owner_id" gorm:"type:uuid;index"`
	
	// 使用统计
	UsageCount   int       `json:"usage_count" gorm:"default:0"`
	LastUsed     time.Time `json:"last_used"`
	SuccessRate  float64   `json:"success_rate" gorm:"default:1.0"`
	
	// 性能指标
	AvgExecutionTime time.Duration `json:"avg_execution_time"`
	
	// 关联
	Agents []*Agent `json:"agents,omitempty" gorm:"many2many:agent_tools;"`
	
	// 聚合根实现
	domainEvents []domain.DomainEvent `gorm:"-"`
}

func (t *Tool) GetID() uuid.UUID {
	return t.ID
}

func (t *Tool) GetVersion() int {
	return t.Version
}

func (t *Tool) MarkAsModified() {
	t.UpdatedAt = time.Now()
}

// NewTool 创建新工具
func NewTool(name string, toolType ToolType, ownerID uuid.UUID) *Tool {
	tool := &Tool{
		BaseEntity: domain.BaseEntity{
			ID:        domain.NewEntityID(),
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		},
		Name:            name,
		Type:            toolType,
		OwnerID:         ownerID,
		Schema:          make(map[string]interface{}),
		Config:          make(map[string]interface{}),
		ExecutionMode:   ExecutionModeSync,
		IsEnabled:       true,
		IsPublic:        false,
		UsageCount:      0,
		SuccessRate:     1.0,
		domainEvents:    make([]domain.DomainEvent, 0),
	}
	
	// 发布工具创建事件
	event := domain.NewDomainEvent("tool.created", tool.ID, tool)
	tool.domainEvents = append(tool.domainEvents, event)
	
	return tool
}

// UpdateSchema 更新工具模式
func (t *Tool) UpdateSchema(schema map[string]interface{}) error {
	// 验证schema格式
	if _, err := json.Marshal(schema); err != nil {
		return NewToolError("invalid schema format")
	}
	
	t.Schema = schema
	t.MarkAsModified()
	
	event := domain.NewDomainEvent("tool.schema.updated", t.ID, map[string]interface{}{
		"tool_id": t.ID,
		"schema":  schema,
	})
	t.domainEvents = append(t.domainEvents, event)
	
	return nil
}

// Enable 启用工具
func (t *Tool) Enable() {
	if !t.IsEnabled {
		t.IsEnabled = true
		t.MarkAsModified()
		
		event := domain.NewDomainEvent("tool.enabled", t.ID, t.ID)
		t.domainEvents = append(t.domainEvents, event)
	}
}

// Disable 禁用工具
func (t *Tool) Disable() {
	if t.IsEnabled {
		t.IsEnabled = false
		t.MarkAsModified()
		
		event := domain.NewDomainEvent("tool.disabled", t.ID, t.ID)
		t.domainEvents = append(t.domainEvents, event)
	}
}

// RecordUsage 记录使用情况
func (t *Tool) RecordUsage(executionTime time.Duration, success bool) {
	t.UsageCount++
	t.LastUsed = time.Now()
	
	// 更新平均执行时间
	if t.AvgExecutionTime == 0 {
		t.AvgExecutionTime = executionTime
	} else {
		t.AvgExecutionTime = time.Duration(
			(int64(t.AvgExecutionTime)*int64(t.UsageCount-1) + int64(executionTime)) / int64(t.UsageCount),
		)
	}
	
	// 更新成功率
	if success {
		t.SuccessRate = (t.SuccessRate*float64(t.UsageCount-1) + 1.0) / float64(t.UsageCount)
	} else {
		t.SuccessRate = (t.SuccessRate * float64(t.UsageCount-1)) / float64(t.UsageCount)
	}
	
	t.MarkAsModified()
	
	event := domain.NewDomainEvent("tool.usage.recorded", t.ID, map[string]interface{}{
		"tool_id":        t.ID,
		"execution_time": executionTime,
		"success":        success,
		"usage_count":    t.UsageCount,
	})
	t.domainEvents = append(t.domainEvents, event)
}

// ValidateInput 验证输入参数
func (t *Tool) ValidateInput(input map[string]interface{}) error {
	// 基于schema验证输入
	// 这里简化实现，实际应用中应该使用JSON Schema验证库
	if len(t.Schema) == 0 {
		return nil // 没有schema则跳过验证
	}
	
	// TODO: 实现完整的JSON Schema验证
	return nil
}

// GetDomainEvents 获取领域事件
func (t *Tool) GetDomainEvents() []domain.DomainEvent {
	return t.domainEvents
}

// ClearDomainEvents 清理领域事件
func (t *Tool) ClearDomainEvents() {
	t.domainEvents = make([]domain.DomainEvent, 0)
}

// ToolExecution 工具执行记录
type ToolExecution struct {
	domain.BaseEntity
	ToolID      uuid.UUID              `json:"tool_id" gorm:"type:uuid;not null;index"`
	AgentID     uuid.UUID              `json:"agent_id" gorm:"type:uuid;not null;index"`
	Input       map[string]interface{} `json:"input" gorm:"type:jsonb"`
	Output      map[string]interface{} `json:"output" gorm:"type:jsonb"`
	Status      ExecutionStatus        `json:"status" gorm:"not null"`
	Error       string                 `json:"error"`
	Duration    time.Duration          `json:"duration"`
	Context     map[string]interface{} `json:"context" gorm:"type:jsonb"`
	
	// 关联
	Tool  *Tool  `json:"tool,omitempty" gorm:"foreignKey:ToolID"`
	Agent *Agent `json:"agent,omitempty" gorm:"foreignKey:AgentID"`
}

// ExecutionStatus 执行状态
type ExecutionStatus string

const (
	ExecutionStatusPending   ExecutionStatus = "pending"
	ExecutionStatusRunning   ExecutionStatus = "running"
	ExecutionStatusCompleted ExecutionStatus = "completed"
	ExecutionStatusFailed    ExecutionStatus = "failed"
	ExecutionStatusTimeout   ExecutionStatus = "timeout"
	ExecutionStatusCancelled ExecutionStatus = "cancelled"
)

// NewToolExecution 创建工具执行记录
func NewToolExecution(toolID, agentID uuid.UUID, input map[string]interface{}) *ToolExecution {
	return &ToolExecution{
		BaseEntity: domain.BaseEntity{
			ID:        domain.NewEntityID(),
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		},
		ToolID:  toolID,
		AgentID: agentID,
		Input:   input,
		Status:  ExecutionStatusPending,
		Context: make(map[string]interface{}),
	}
}

// Complete 完成执行
func (te *ToolExecution) Complete(output map[string]interface{}, duration time.Duration) {
	te.Status = ExecutionStatusCompleted
	te.Output = output
	te.Duration = duration
	te.UpdatedAt = time.Now()
}

// Fail 执行失败
func (te *ToolExecution) Fail(error string, duration time.Duration) {
	te.Status = ExecutionStatusFailed
	te.Error = error
	te.Duration = duration
	te.UpdatedAt = time.Now()
}

// ToolError 工具错误
type ToolError struct {
	message string
}

func NewToolError(message string) *ToolError {
	return &ToolError{message: message}
}

func (e *ToolError) Error() string {
	return e.message
}

// ToolRepository 工具仓储接口
type ToolRepository interface {
	domain.Repository[*Tool]
	FindByOwnerID(ctx context.Context, ownerID uuid.UUID) ([]*Tool, error)
	FindByType(ctx context.Context, toolType ToolType) ([]*Tool, error)
	FindPublicTools(ctx context.Context) ([]*Tool, error)
	FindEnabledTools(ctx context.Context) ([]*Tool, error)
	FindByAgentID(ctx context.Context, agentID uuid.UUID) ([]*Tool, error)
}

// ToolExecutionRepository 工具执行仓储接口
type ToolExecutionRepository interface {
	domain.Repository[*ToolExecution]
	FindByToolID(ctx context.Context, toolID uuid.UUID, offset, limit int) ([]*ToolExecution, error)
	FindByAgentID(ctx context.Context, agentID uuid.UUID, offset, limit int) ([]*ToolExecution, error)
	FindByStatus(ctx context.Context, status ExecutionStatus) ([]*ToolExecution, error)
}
