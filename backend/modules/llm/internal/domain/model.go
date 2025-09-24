package domain

import (
	"context"
	"time"
	
	"github.com/google/uuid"
	"github.com/noah-loop/backend/shared/pkg/domain"
)

// ModelProvider 模型提供商
type ModelProvider string

const (
	ProviderOpenAI    ModelProvider = "openai"
	ProviderAnthropic ModelProvider = "anthropic"
	ProviderLocal     ModelProvider = "local"
	ProviderCustom    ModelProvider = "custom"
)

// ModelType 模型类型
type ModelType string

const (
	ModelTypeChat       ModelType = "chat"
	ModelTypeCompletion ModelType = "completion"
	ModelTypeEmbedding  ModelType = "embedding"
	ModelTypeImage      ModelType = "image"
	ModelTypeAudio      ModelType = "audio"
)

// Model 大模型实体
type Model struct {
	domain.BaseEntity
	Name         string                 `json:"name" gorm:"not null;index"`
	Provider     ModelProvider          `json:"provider" gorm:"not null"`
	Type         ModelType              `json:"type" gorm:"not null"`
	Version      string                 `json:"version"`
	Description  string                 `json:"description"`
	Config       map[string]interface{} `json:"config" gorm:"type:jsonb"`
	Capabilities []string               `json:"capabilities" gorm:"type:text[]"`
	MaxTokens    int                    `json:"max_tokens"`
	PricePerK    float64                `json:"price_per_k"` // 每千token价格
	IsActive     bool                   `json:"is_active" gorm:"default:true"`
	
	// 聚合根实现
	domainEvents []domain.DomainEvent `gorm:"-"`
}

func (m *Model) GetID() uuid.UUID {
	return m.ID
}

func (m *Model) GetVersion() int {
	return m.Version
}

func (m *Model) MarkAsModified() {
	m.UpdatedAt = time.Now()
}

// NewModel 创建新模型
func NewModel(name string, provider ModelProvider, modelType ModelType) *Model {
	model := &Model{
		BaseEntity: domain.BaseEntity{
			ID:        domain.NewEntityID(),
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		},
		Name:         name,
		Provider:     provider,
		Type:         modelType,
		Config:       make(map[string]interface{}),
		Capabilities: make([]string, 0),
		IsActive:     true,
		domainEvents: make([]domain.DomainEvent, 0),
	}
	
	// 发布模型创建事件
	event := domain.NewDomainEvent("model.created", model.ID, model)
	model.domainEvents = append(model.domainEvents, event)
	
	return model
}

// UpdateConfig 更新模型配置
func (m *Model) UpdateConfig(config map[string]interface{}) error {
	m.Config = config
	m.MarkAsModified()
	
	// 发布配置更新事件
	event := domain.NewDomainEvent("model.config.updated", m.ID, map[string]interface{}{
		"model_id": m.ID,
		"config":   config,
	})
	m.domainEvents = append(m.domainEvents, event)
	
	return nil
}

// Activate 激活模型
func (m *Model) Activate() {
	if !m.IsActive {
		m.IsActive = true
		m.MarkAsModified()
		
		event := domain.NewDomainEvent("model.activated", m.ID, m.ID)
		m.domainEvents = append(m.domainEvents, event)
	}
}

// Deactivate 停用模型
func (m *Model) Deactivate() {
	if m.IsActive {
		m.IsActive = false
		m.MarkAsModified()
		
		event := domain.NewDomainEvent("model.deactivated", m.ID, m.ID)
		m.domainEvents = append(m.domainEvents, event)
	}
}

// GetDomainEvents 获取领域事件
func (m *Model) GetDomainEvents() []domain.DomainEvent {
	return m.domainEvents
}

// ClearDomainEvents 清理领域事件
func (m *Model) ClearDomainEvents() {
	m.domainEvents = make([]domain.DomainEvent, 0)
}

// ModelRepository 模型仓储接口
type ModelRepository interface {
	domain.Repository[*Model]
	FindByProvider(ctx context.Context, provider ModelProvider) ([]*Model, error)
	FindByType(ctx context.Context, modelType ModelType) ([]*Model, error)
	FindActiveModels(ctx context.Context) ([]*Model, error)
	FindByNameAndProvider(ctx context.Context, name string, provider ModelProvider) (*Model, error)
}
