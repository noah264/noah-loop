package domain

import (
	"time"
	
	"github.com/google/uuid"
	"github.com/noah-loop/backend/shared/pkg/domain"
)

// ContextType 上下文类型
type ContextType string

const (
	ContextTypeConversation ContextType = "conversation" // 对话上下文
	ContextTypeDocument     ContextType = "document"     // 文档上下文
	ContextTypeCode         ContextType = "code"         // 代码上下文
	ContextTypeTask         ContextType = "task"         // 任务上下文
	ContextTypeKnowledge    ContextType = "knowledge"    // 知识上下文
)

// CompressionLevel 压缩级别
type CompressionLevel int

const (
	CompressionNone   CompressionLevel = 0
	CompressionLight  CompressionLevel = 1
	CompressionMedium CompressionLevel = 2
	CompressionHeavy  CompressionLevel = 3
)

// Context 上下文实体
type Context struct {
	domain.BaseEntity
	SessionID      uuid.UUID                 `json:"session_id" gorm:"type:uuid;not null;index"`
	Type           ContextType               `json:"type" gorm:"not null"`
	Title          string                    `json:"title"`
	Content        string                    `json:"content" gorm:"type:text"`
	Metadata       map[string]interface{}    `json:"metadata" gorm:"type:jsonb"`
	TokenCount     int                       `json:"token_count"`
	Priority       int                       `json:"priority" gorm:"default:1"`
	IsCompressed   bool                      `json:"is_compressed" gorm:"default:false"`
	CompressionLevel CompressionLevel        `json:"compression_level" gorm:"default:0"`
	OriginalSize   int                       `json:"original_size"`
	CompressedSize int                       `json:"compressed_size"`
	LastAccessed   time.Time                 `json:"last_accessed"`
	AccessCount    int                       `json:"access_count" gorm:"default:0"`
	
	// 关联
	Session *Session `json:"session,omitempty" gorm:"foreignKey:SessionID"`
	
	// 聚合根实现
	domainEvents []domain.DomainEvent `gorm:"-"`
}

func (c *Context) GetID() uuid.UUID {
	return c.ID
}

func (c *Context) GetVersion() int {
	return c.Version
}

func (c *Context) MarkAsModified() {
	c.UpdatedAt = time.Now()
}

// NewContext 创建新上下文
func NewContext(sessionID uuid.UUID, contextType ContextType, title, content string) *Context {
	context := &Context{
		BaseEntity: domain.BaseEntity{
			ID:        domain.NewEntityID(),
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		},
		SessionID:    sessionID,
		Type:         contextType,
		Title:        title,
		Content:      content,
		Metadata:     make(map[string]interface{}),
		Priority:     1,
		IsCompressed: false,
		CompressionLevel: CompressionNone,
		LastAccessed: time.Now(),
		AccessCount:  0,
		domainEvents: make([]domain.DomainEvent, 0),
	}
	
	// 计算token数量（简化实现）
	context.TokenCount = len(content) / 4 // 粗略估算
	context.OriginalSize = len(content)
	
	// 发布上下文创建事件
	event := domain.NewDomainEvent("context.created", context.ID, context)
	context.domainEvents = append(context.domainEvents, event)
	
	return context
}

// Access 访问上下文
func (c *Context) Access() {
	c.AccessCount++
	c.LastAccessed = time.Now()
	c.MarkAsModified()
	
	event := domain.NewDomainEvent("context.accessed", c.ID, map[string]interface{}{
		"context_id":   c.ID,
		"access_count": c.AccessCount,
	})
	c.domainEvents = append(c.domainEvents, event)
}

// Compress 压缩上下文
func (c *Context) Compress(level CompressionLevel, compressedContent string) error {
	if c.IsCompressed {
		return NewContextError("context is already compressed")
	}
	
	c.IsCompressed = true
	c.CompressionLevel = level
	c.Content = compressedContent
	c.CompressedSize = len(compressedContent)
	c.MarkAsModified()
	
	event := domain.NewDomainEvent("context.compressed", c.ID, map[string]interface{}{
		"context_id":        c.ID,
		"compression_level": level,
		"original_size":     c.OriginalSize,
		"compressed_size":   c.CompressedSize,
		"compression_ratio": float64(c.CompressedSize) / float64(c.OriginalSize),
	})
	c.domainEvents = append(c.domainEvents, event)
	
	return nil
}

// Decompress 解压缩上下文
func (c *Context) Decompress(originalContent string) error {
	if !c.IsCompressed {
		return NewContextError("context is not compressed")
	}
	
	c.IsCompressed = false
	c.CompressionLevel = CompressionNone
	c.Content = originalContent
	c.MarkAsModified()
	
	event := domain.NewDomainEvent("context.decompressed", c.ID, c.ID)
	c.domainEvents = append(c.domainEvents, event)
	
	return nil
}

// UpdatePriority 更新优先级
func (c *Context) UpdatePriority(priority int) {
	if priority < 1 {
		priority = 1
	}
	if priority > 10 {
		priority = 10
	}
	
	oldPriority := c.Priority
	c.Priority = priority
	c.MarkAsModified()
	
	event := domain.NewDomainEvent("context.priority.updated", c.ID, map[string]interface{}{
		"context_id":   c.ID,
		"old_priority": oldPriority,
		"new_priority": priority,
	})
	c.domainEvents = append(c.domainEvents, event)
}

// GetRelevanceScore 获取相关性评分
func (c *Context) GetRelevanceScore() float64 {
	// 基于访问频率、优先级、时效性计算相关性
	recencyScore := 1.0 / (1.0 + time.Since(c.LastAccessed).Hours()/24)
	frequencyScore := min(1.0, float64(c.AccessCount)/10.0)
	priorityScore := float64(c.Priority) / 10.0
	
	return (recencyScore + frequencyScore + priorityScore) / 3.0
}

// GetDomainEvents 获取领域事件
func (c *Context) GetDomainEvents() []domain.DomainEvent {
	return c.domainEvents
}

// ClearDomainEvents 清理领域事件
func (c *Context) ClearDomainEvents() {
	c.domainEvents = make([]domain.DomainEvent, 0)
}

// ContextError 上下文错误
type ContextError struct {
	message string
}

func NewContextError(message string) *ContextError {
	return &ContextError{message: message}
}

func (e *ContextError) Error() string {
	return e.message
}

// ContextRepository 上下文仓储接口
type ContextRepository interface {
	domain.Repository[*Context]
	FindBySessionID(ctx context.Context, sessionID uuid.UUID) ([]*Context, error)
	FindByType(ctx context.Context, contextType ContextType) ([]*Context, error)
	FindByPriority(ctx context.Context, minPriority int) ([]*Context, error)
	FindExpiredContexts(ctx context.Context, before time.Time) ([]*Context, error)
	GetSessionContextSize(ctx context.Context, sessionID uuid.UUID) (int, error)
}

func min(a, b float64) float64 {
	if a < b {
		return a
	}
	return b
}
