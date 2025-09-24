package domain

import (
	"time"
	
	"github.com/google/uuid"
	"github.com/noah-loop/backend/shared/pkg/domain"
)

// SessionStatus 会话状态
type SessionStatus string

const (
	SessionStatusActive    SessionStatus = "active"
	SessionStatusIdle      SessionStatus = "idle"
	SessionStatusArchived  SessionStatus = "archived"
	SessionStatusExpired   SessionStatus = "expired"
)

// Session 会话实体
type Session struct {
	domain.BaseEntity
	UserID         uuid.UUID                 `json:"user_id" gorm:"type:uuid;not null;index"`
	AgentID        uuid.UUID                 `json:"agent_id" gorm:"type:uuid;index"`
	Status         SessionStatus             `json:"status" gorm:"not null;default:'active'"`
	Title          string                    `json:"title"`
	Description    string                    `json:"description"`
	Metadata       map[string]interface{}    `json:"metadata" gorm:"type:jsonb"`
	MaxContextSize int                       `json:"max_context_size" gorm:"default:8192"`
	CurrentSize    int                       `json:"current_size" gorm:"default:0"`
	MessageCount   int                       `json:"message_count" gorm:"default:0"`
	LastActivity   time.Time                 `json:"last_activity"`
	ExpiresAt      *time.Time                `json:"expires_at"`
	
	// 关联
	Contexts []*Context `json:"contexts,omitempty" gorm:"foreignKey:SessionID"`
	
	// 聚合根实现
	domainEvents []domain.DomainEvent `gorm:"-"`
}

func (s *Session) GetID() uuid.UUID {
	return s.ID
}

func (s *Session) GetVersion() int {
	return s.Version
}

func (s *Session) MarkAsModified() {
	s.UpdatedAt = time.Now()
	s.LastActivity = time.Now()
}

// NewSession 创建新会话
func NewSession(userID, agentID uuid.UUID, title string) *Session {
	session := &Session{
		BaseEntity: domain.BaseEntity{
			ID:        domain.NewEntityID(),
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		},
		UserID:         userID,
		AgentID:        agentID,
		Status:         SessionStatusActive,
		Title:          title,
		Metadata:       make(map[string]interface{}),
		MaxContextSize: 8192,
		CurrentSize:    0,
		MessageCount:   0,
		LastActivity:   time.Now(),
		domainEvents:   make([]domain.DomainEvent, 0),
	}
	
	// 设置默认过期时间（24小时）
	expiresAt := time.Now().Add(24 * time.Hour)
	session.ExpiresAt = &expiresAt
	
	// 发布会话创建事件
	event := domain.NewDomainEvent("session.created", session.ID, session)
	session.domainEvents = append(session.domainEvents, event)
	
	return session
}

// AddContext 添加上下文
func (s *Session) AddContext(context *Context) error {
	if s.Status != SessionStatusActive {
		return NewSessionError("cannot add context to inactive session")
	}
	
	// 检查是否超出最大上下文大小
	if s.CurrentSize+context.TokenCount > s.MaxContextSize {
		// 需要进行上下文管理
		if err := s.manageContextSize(); err != nil {
			return err
		}
	}
	
	s.Contexts = append(s.Contexts, context)
	s.CurrentSize += context.TokenCount
	s.MessageCount++
	s.MarkAsModified()
	
	event := domain.NewDomainEvent("session.context.added", s.ID, map[string]interface{}{
		"session_id":  s.ID,
		"context_id":  context.ID,
		"token_count": context.TokenCount,
		"total_size":  s.CurrentSize,
	})
	s.domainEvents = append(s.domainEvents, event)
	
	return nil
}

// manageContextSize 管理上下文大小
func (s *Session) manageContextSize() error {
	// 按相关性排序，移除不重要的上下文
	var contextsToRemove []*Context
	targetSize := s.MaxContextSize * 70 / 100 // 保持在70%以下
	
	// 找出低优先级和低相关性的上下文
	for _, context := range s.Contexts {
		if context.Priority <= 2 && context.GetRelevanceScore() < 0.3 {
			contextsToRemove = append(contextsToRemove, context)
			s.CurrentSize -= context.TokenCount
			if s.CurrentSize <= targetSize {
				break
			}
		}
	}
	
	// 移除选中的上下文
	for _, contextToRemove := range contextsToRemove {
		s.removeContext(contextToRemove)
	}
	
	return nil
}

// removeContext 移除上下文
func (s *Session) removeContext(context *Context) {
	for i, c := range s.Contexts {
		if c.ID == context.ID {
			s.Contexts = append(s.Contexts[:i], s.Contexts[i+1:]...)
			s.CurrentSize -= context.TokenCount
			break
		}
	}
	
	event := domain.NewDomainEvent("session.context.removed", s.ID, map[string]interface{}{
		"session_id": s.ID,
		"context_id": context.ID,
		"reason":     "size_management",
	})
	s.domainEvents = append(s.domainEvents, event)
}

// UpdateActivity 更新活动时间
func (s *Session) UpdateActivity() {
	s.LastActivity = time.Now()
	
	// 如果是空闲状态，恢复为活跃状态
	if s.Status == SessionStatusIdle {
		s.Status = SessionStatusActive
		
		event := domain.NewDomainEvent("session.activated", s.ID, s.ID)
		s.domainEvents = append(s.domainEvents, event)
	}
	
	s.MarkAsModified()
}

// SetIdle 设置为空闲状态
func (s *Session) SetIdle() {
	if s.Status == SessionStatusActive {
		s.Status = SessionStatusIdle
		s.MarkAsModified()
		
		event := domain.NewDomainEvent("session.idle", s.ID, s.ID)
		s.domainEvents = append(s.domainEvents, event)
	}
}

// Archive 归档会话
func (s *Session) Archive() {
	s.Status = SessionStatusArchived
	s.MarkAsModified()
	
	event := domain.NewDomainEvent("session.archived", s.ID, s.ID)
	s.domainEvents = append(s.domainEvents, event)
}

// Expire 过期会话
func (s *Session) Expire() {
	s.Status = SessionStatusExpired
	s.MarkAsModified()
	
	event := domain.NewDomainEvent("session.expired", s.ID, s.ID)
	s.domainEvents = append(s.domainEvents, event)
}

// IsExpired 检查是否过期
func (s *Session) IsExpired() bool {
	if s.ExpiresAt == nil {
		return false
	}
	return time.Now().After(*s.ExpiresAt)
}

// ExtendExpiry 延长过期时间
func (s *Session) ExtendExpiry(duration time.Duration) {
	newExpiryTime := time.Now().Add(duration)
	s.ExpiresAt = &newExpiryTime
	s.MarkAsModified()
	
	event := domain.NewDomainEvent("session.expiry.extended", s.ID, map[string]interface{}{
		"session_id": s.ID,
		"expires_at": newExpiryTime,
		"duration":   duration.String(),
	})
	s.domainEvents = append(s.domainEvents, event)
}

// GetContextsByType 根据类型获取上下文
func (s *Session) GetContextsByType(contextType ContextType) []*Context {
	var contexts []*Context
	for _, context := range s.Contexts {
		if context.Type == contextType {
			contexts = append(contexts, context)
		}
	}
	return contexts
}

// GetRecentContexts 获取最近的上下文
func (s *Session) GetRecentContexts(limit int) []*Context {
	if len(s.Contexts) <= limit {
		return s.Contexts
	}
	
	// 按创建时间排序，返回最近的
	contexts := make([]*Context, len(s.Contexts))
	copy(contexts, s.Contexts)
	
	// 简单排序（实际应用中可能需要更复杂的排序逻辑）
	return contexts[len(contexts)-limit:]
}

// GetDomainEvents 获取领域事件
func (s *Session) GetDomainEvents() []domain.DomainEvent {
	return s.domainEvents
}

// ClearDomainEvents 清理领域事件
func (s *Session) ClearDomainEvents() {
	s.domainEvents = make([]domain.DomainEvent, 0)
}

// SessionError 会话错误
type SessionError struct {
	message string
}

func NewSessionError(message string) *SessionError {
	return &SessionError{message: message}
}

func (e *SessionError) Error() string {
	return e.message
}

// SessionRepository 会话仓储接口
type SessionRepository interface {
	domain.Repository[*Session]
	FindByUserID(ctx context.Context, userID uuid.UUID) ([]*Session, error)
	FindByAgentID(ctx context.Context, agentID uuid.UUID) ([]*Session, error)
	FindByStatus(ctx context.Context, status SessionStatus) ([]*Session, error)
	FindExpiredSessions(ctx context.Context) ([]*Session, error)
	FindIdleSessions(ctx context.Context, idleThreshold time.Duration) ([]*Session, error)
}
