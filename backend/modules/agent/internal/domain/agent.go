package domain

import (
	"time"
	
	"github.com/google/uuid"
	"github.com/noah-loop/backend/shared/pkg/domain"
)

// AgentType 智能体类型
type AgentType string

const (
	AgentTypeConversational AgentType = "conversational" // 对话型
	AgentTypeTask          AgentType = "task"           // 任务型
	AgentTypeReflective    AgentType = "reflective"     // 反思型
	AgentTypePlanning      AgentType = "planning"       // 规划型
	AgentTypeMultiModal    AgentType = "multimodal"     // 多模态
)

// AgentStatus 智能体状态
type AgentStatus string

const (
	AgentStatusIdle       AgentStatus = "idle"       // 空闲
	AgentStatusBusy       AgentStatus = "busy"       // 忙碌
	AgentStatusLearning   AgentStatus = "learning"   // 学习中
	AgentStatusSleeping   AgentStatus = "sleeping"   // 休眠
	AgentStatusMaintenance AgentStatus = "maintenance" // 维护中
)

// Agent 智能体实体
type Agent struct {
	domain.BaseEntity
	Name        string                 `json:"name" gorm:"not null;index"`
	Type        AgentType              `json:"type" gorm:"not null"`
	Status      AgentStatus            `json:"status" gorm:"not null;default:'idle'"`
	Description string                 `json:"description"`
	SystemPrompt string                `json:"system_prompt" gorm:"type:text"`
	Config      map[string]interface{} `json:"config" gorm:"type:jsonb"`
	Capabilities []string              `json:"capabilities" gorm:"type:text[]"`
	Memory      *AgentMemory          `json:"memory,omitempty" gorm:"foreignKey:AgentID"`
	Tools       []*Tool               `json:"tools,omitempty" gorm:"many2many:agent_tools;"`
	OwnerID     uuid.UUID             `json:"owner_id" gorm:"type:uuid;index"`
	IsActive    bool                  `json:"is_active" gorm:"default:true"`
	LastActiveAt time.Time            `json:"last_active_at"`
	
	// 学习和适应相关
	LearningRate    float64 `json:"learning_rate" gorm:"default:0.1"`
	MemoryCapacity  int     `json:"memory_capacity" gorm:"default:1000"`
	ContextWindow   int     `json:"context_window" gorm:"default:4096"`
	
	// 聚合根实现
	domainEvents []domain.DomainEvent `gorm:"-"`
}

func (a *Agent) GetID() uuid.UUID {
	return a.ID
}

func (a *Agent) GetVersion() int {
	return a.Version
}

func (a *Agent) MarkAsModified() {
	a.UpdatedAt = time.Now()
	a.LastActiveAt = time.Now()
}

// NewAgent 创建新智能体
func NewAgent(name string, agentType AgentType, ownerID uuid.UUID) *Agent {
	agent := &Agent{
		BaseEntity: domain.BaseEntity{
			ID:        domain.NewEntityID(),
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		},
		Name:           name,
		Type:           agentType,
		Status:         AgentStatusIdle,
		OwnerID:        ownerID,
		Config:         make(map[string]interface{}),
		Capabilities:   make([]string, 0),
		IsActive:       true,
		LastActiveAt:   time.Now(),
		LearningRate:   0.1,
		MemoryCapacity: 1000,
		ContextWindow:  4096,
		domainEvents:   make([]domain.DomainEvent, 0),
	}
	
	// 发布智能体创建事件
	event := domain.NewDomainEvent("agent.created", agent.ID, agent)
	agent.domainEvents = append(agent.domainEvents, event)
	
	return agent
}

// UpdateSystemPrompt 更新系统提示
func (a *Agent) UpdateSystemPrompt(prompt string) {
	a.SystemPrompt = prompt
	a.MarkAsModified()
	
	event := domain.NewDomainEvent("agent.system_prompt.updated", a.ID, map[string]interface{}{
		"agent_id": a.ID,
		"prompt":   prompt,
	})
	a.domainEvents = append(a.domainEvents, event)
}

// ChangeStatus 改变状态
func (a *Agent) ChangeStatus(status AgentStatus) error {
	if a.Status == status {
		return nil
	}
	
	oldStatus := a.Status
	a.Status = status
	a.MarkAsModified()
	
	event := domain.NewDomainEvent("agent.status.changed", a.ID, map[string]interface{}{
		"agent_id":   a.ID,
		"old_status": oldStatus,
		"new_status": status,
	})
	a.domainEvents = append(a.domainEvents, event)
	
	return nil
}

// AddTool 添加工具
func (a *Agent) AddTool(tool *Tool) error {
	// 检查工具是否已存在
	for _, existingTool := range a.Tools {
		if existingTool.ID == tool.ID {
			return NewAgentError("tool already exists")
		}
	}
	
	a.Tools = append(a.Tools, tool)
	a.MarkAsModified()
	
	event := domain.NewDomainEvent("agent.tool.added", a.ID, map[string]interface{}{
		"agent_id": a.ID,
		"tool_id":  tool.ID,
		"tool_name": tool.Name,
	})
	a.domainEvents = append(a.domainEvents, event)
	
	return nil
}

// RemoveTool 移除工具
func (a *Agent) RemoveTool(toolID uuid.UUID) error {
	for i, tool := range a.Tools {
		if tool.ID == toolID {
			a.Tools = append(a.Tools[:i], a.Tools[i+1:]...)
			a.MarkAsModified()
			
			event := domain.NewDomainEvent("agent.tool.removed", a.ID, map[string]interface{}{
				"agent_id": a.ID,
				"tool_id":  toolID,
			})
			a.domainEvents = append(a.domainEvents, event)
			
			return nil
		}
	}
	return NewAgentError("tool not found")
}

// Learn 学习新知识
func (a *Agent) Learn(knowledge string, importance float64) error {
	if a.Memory == nil {
		return NewAgentError("agent memory not initialized")
	}
	
	// 创建记忆条目
	memory := NewMemory(knowledge, MemoryTypeLearned, importance)
	
	// 添加到记忆中
	if err := a.Memory.AddMemory(memory); err != nil {
		return err
	}
	
	a.MarkAsModified()
	
	event := domain.NewDomainEvent("agent.learned", a.ID, map[string]interface{}{
		"agent_id":   a.ID,
		"knowledge":  knowledge,
		"importance": importance,
	})
	a.domainEvents = append(a.domainEvents, event)
	
	return nil
}

// CanUse 检查是否可以使用工具
func (a *Agent) CanUse(toolName string) bool {
	for _, tool := range a.Tools {
		if tool.Name == toolName && tool.IsEnabled {
			return true
		}
	}
	return false
}

// GetDomainEvents 获取领域事件
func (a *Agent) GetDomainEvents() []domain.DomainEvent {
	return a.domainEvents
}

// ClearDomainEvents 清理领域事件
func (a *Agent) ClearDomainEvents() {
	a.domainEvents = make([]domain.DomainEvent, 0)
}

// AgentError 智能体错误
type AgentError struct {
	message string
}

func NewAgentError(message string) *AgentError {
	return &AgentError{message: message}
}

func (e *AgentError) Error() string {
	return e.message
}

// AgentRepository 智能体仓储接口
type AgentRepository interface {
	domain.Repository[*Agent]
	FindByOwnerID(ctx context.Context, ownerID uuid.UUID) ([]*Agent, error)
	FindByType(ctx context.Context, agentType AgentType) ([]*Agent, error)
	FindActiveAgents(ctx context.Context) ([]*Agent, error)
	FindByStatus(ctx context.Context, status AgentStatus) ([]*Agent, error)
}
