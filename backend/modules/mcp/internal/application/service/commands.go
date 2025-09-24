package service

import (
	"errors"
	"time"
	
	"github.com/google/uuid"
	"github.com/noah-loop/backend/modules/mcp/internal/domain"
	"github.com/noah-loop/backend/shared/pkg/application"
)

// CreateSessionCommand 创建会话命令
type CreateSessionCommand struct {
	application.BaseCommand
	UserID         uuid.UUID                 `json:"user_id" binding:"required"`
	AgentID        uuid.UUID                 `json:"agent_id" binding:"required"`
	Title          string                    `json:"title" binding:"required"`
	Description    string                    `json:"description"`
	Metadata       map[string]interface{}    `json:"metadata"`
	MaxContextSize int                       `json:"max_context_size"`
	ExpiresIn      time.Duration             `json:"expires_in"` // 过期时间间隔
}

func NewCreateSessionCommand() *CreateSessionCommand {
	return &CreateSessionCommand{
		BaseCommand: application.BaseCommand{
			CommandID:   uuid.New(),
			CommandType: "create_session",
		},
		Metadata: make(map[string]interface{}),
	}
}

func (c *CreateSessionCommand) Validate() error {
	if c.UserID == uuid.Nil {
		return errors.New("user ID is required")
	}
	
	if c.AgentID == uuid.Nil {
		return errors.New("agent ID is required")
	}
	
	if c.Title == "" {
		return errors.New("title is required")
	}
	
	if c.MaxContextSize > 0 && c.MaxContextSize < 1024 {
		return errors.New("max context size must be at least 1024 tokens")
	}
	
	return nil
}

// AddContextCommand 添加上下文命令
type AddContextCommand struct {
	application.BaseCommand
	SessionID        uuid.UUID                   `json:"session_id" binding:"required"`
	Type             domain.ContextType          `json:"type" binding:"required"`
	Title            string                      `json:"title" binding:"required"`
	Content          string                      `json:"content" binding:"required"`
	Metadata         map[string]interface{}      `json:"metadata"`
	Priority         int                         `json:"priority"`
	CompressionLevel domain.CompressionLevel     `json:"compression_level"`
}

func NewAddContextCommand() *AddContextCommand {
	return &AddContextCommand{
		BaseCommand: application.BaseCommand{
			CommandID:   uuid.New(),
			CommandType: "add_context",
		},
		Metadata:         make(map[string]interface{}),
		Priority:         1,
		CompressionLevel: domain.CompressionNone,
	}
}

func (c *AddContextCommand) Validate() error {
	if c.SessionID == uuid.Nil {
		return errors.New("session ID is required")
	}
	
	if c.Title == "" {
		return errors.New("title is required")
	}
	
	if c.Content == "" {
		return errors.New("content is required")
	}
	
	// 验证上下文类型
	switch c.Type {
	case domain.ContextTypeConversation, domain.ContextTypeDocument, domain.ContextTypeCode, domain.ContextTypeTask, domain.ContextTypeKnowledge:
		// valid
	default:
		return errors.New("invalid context type")
	}
	
	if c.Priority < 1 || c.Priority > 10 {
		return errors.New("priority must be between 1 and 10")
	}
	
	// 验证压缩级别
	switch c.CompressionLevel {
	case domain.CompressionNone, domain.CompressionLight, domain.CompressionMedium, domain.CompressionHeavy:
		// valid
	default:
		return errors.New("invalid compression level")
	}
	
	return nil
}

// UpdateContextCommand 更新上下文命令
type UpdateContextCommand struct {
	application.BaseCommand
	ContextID   uuid.UUID                 `json:"context_id" binding:"required"`
	Title       *string                   `json:"title"`
	Content     *string                   `json:"content"`
	Priority    *int                      `json:"priority"`
	Metadata    map[string]interface{}    `json:"metadata"`
}

func NewUpdateContextCommand(contextID uuid.UUID) *UpdateContextCommand {
	return &UpdateContextCommand{
		BaseCommand: application.BaseCommand{
			CommandID:   uuid.New(),
			CommandType: "update_context",
		},
		ContextID: contextID,
	}
}

func (c *UpdateContextCommand) Validate() error {
	if c.ContextID == uuid.Nil {
		return errors.New("context ID is required")
	}
	
	if c.Priority != nil && (*c.Priority < 1 || *c.Priority > 10) {
		return errors.New("priority must be between 1 and 10")
	}
	
	return nil
}

// UpdateSessionCommand 更新会话命令
type UpdateSessionCommand struct {
	application.BaseCommand
	SessionID      uuid.UUID                 `json:"session_id" binding:"required"`
	Title          *string                   `json:"title"`
	Description    *string                   `json:"description"`
	MaxContextSize *int                      `json:"max_context_size"`
	Metadata       map[string]interface{}    `json:"metadata"`
	ExtendExpiry   *time.Duration            `json:"extend_expiry"`
}

func NewUpdateSessionCommand(sessionID uuid.UUID) *UpdateSessionCommand {
	return &UpdateSessionCommand{
		BaseCommand: application.BaseCommand{
			CommandID:   uuid.New(),
			CommandType: "update_session",
		},
		SessionID: sessionID,
	}
}

func (c *UpdateSessionCommand) Validate() error {
	if c.SessionID == uuid.Nil {
		return errors.New("session ID is required")
	}
	
	if c.MaxContextSize != nil && *c.MaxContextSize < 1024 {
		return errors.New("max context size must be at least 1024 tokens")
	}
	
	return nil
}

// 查询对象

// GetSessionQuery 获取会话查询
type GetSessionQuery struct {
	application.BaseQuery
	SessionID uuid.UUID `form:"session_id" binding:"required"`
}

func NewGetSessionQuery() *GetSessionQuery {
	return &GetSessionQuery{
		BaseQuery: application.BaseQuery{
			QueryID:   uuid.New(),
			QueryType: "get_session",
		},
	}
}

func (q *GetSessionQuery) Validate() error {
	if q.SessionID == uuid.Nil {
		return errors.New("session ID is required")
	}
	return nil
}

// GetContextQuery 获取上下文查询
type GetContextQuery struct {
	application.BaseQuery
	ContextID  uuid.UUID `form:"context_id" binding:"required"`
	Decompress bool      `form:"decompress,default=false"`
}

func NewGetContextQuery() *GetContextQuery {
	return &GetContextQuery{
		BaseQuery: application.BaseQuery{
			QueryID:   uuid.New(),
			QueryType: "get_context",
		},
		Decompress: false,
	}
}

func (q *GetContextQuery) Validate() error {
	if q.ContextID == uuid.Nil {
		return errors.New("context ID is required")
	}
	return nil
}

// GetSessionContextsQuery 获取会话上下文查询
type GetSessionContextsQuery struct {
	application.BaseQuery
	SessionID   uuid.UUID           `form:"session_id" binding:"required"`
	Type        *domain.ContextType `form:"type"`
	MinPriority int                 `form:"min_priority,default=0"`
	Page        int                 `form:"page,default=1"`
	PageSize    int                 `form:"page_size,default=20"`
}

func NewGetSessionContextsQuery() *GetSessionContextsQuery {
	return &GetSessionContextsQuery{
		BaseQuery: application.BaseQuery{
			QueryID:   uuid.New(),
			QueryType: "get_session_contexts",
		},
		MinPriority: 0,
		Page:        1,
		PageSize:    20,
	}
}

func (q *GetSessionContextsQuery) Validate() error {
	if q.SessionID == uuid.Nil {
		return errors.New("session ID is required")
	}
	
	if q.Page <= 0 {
		return errors.New("page must be greater than 0")
	}
	
	if q.PageSize <= 0 || q.PageSize > 100 {
		return errors.New("page size must be between 1 and 100")
	}
	
	if q.MinPriority < 0 || q.MinPriority > 10 {
		return errors.New("min priority must be between 0 and 10")
	}
	
	return nil
}

// GetSessionsQuery 获取会话列表查询
type GetSessionsQuery struct {
	application.BaseQuery
	UserID   *uuid.UUID             `form:"user_id"`
	AgentID  *uuid.UUID             `form:"agent_id"`
	Status   *domain.SessionStatus  `form:"status"`
	Page     int                    `form:"page,default=1"`
	PageSize int                    `form:"page_size,default=20"`
}

func NewGetSessionsQuery() *GetSessionsQuery {
	return &GetSessionsQuery{
		BaseQuery: application.BaseQuery{
			QueryID:   uuid.New(),
			QueryType: "get_sessions",
		},
		Page:     1,
		PageSize: 20,
	}
}

func (q *GetSessionsQuery) Validate() error {
	if q.Page <= 0 {
		return errors.New("page must be greater than 0")
	}
	
	if q.PageSize <= 0 || q.PageSize > 100 {
		return errors.New("page size must be between 1 and 100")
	}
	
	return nil
}
