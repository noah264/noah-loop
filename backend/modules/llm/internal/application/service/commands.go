package service

import (
	"errors"
	
	"github.com/google/uuid"
	"github.com/noah-loop/backend/modules/llm/internal/domain"
	"github.com/noah-loop/backend/shared/pkg/application"
)

// CreateModelCommand 创建模型命令
type CreateModelCommand struct {
	application.BaseCommand
	Name         string                        `json:"name" binding:"required"`
	Provider     domain.ModelProvider          `json:"provider" binding:"required"`
	Type         domain.ModelType              `json:"type" binding:"required"`
	Version      string                        `json:"version"`
	Description  string                        `json:"description"`
	Config       map[string]interface{}        `json:"config"`
	Capabilities []string                      `json:"capabilities"`
	MaxTokens    int                           `json:"max_tokens"`
	PricePerK    float64                       `json:"price_per_k"`
}

func NewCreateModelCommand() *CreateModelCommand {
	return &CreateModelCommand{
		BaseCommand: application.BaseCommand{
			CommandID:   uuid.New(),
			CommandType: "create_model",
		},
		Config:       make(map[string]interface{}),
		Capabilities: make([]string, 0),
	}
}

func (c *CreateModelCommand) Validate() error {
	if c.Name == "" {
		return errors.New("model name is required")
	}
	
	if c.Provider == "" {
		return errors.New("model provider is required")
	}
	
	if c.Type == "" {
		return errors.New("model type is required")
	}
	
	// 验证提供商
	switch c.Provider {
	case domain.ProviderOpenAI, domain.ProviderAnthropic, domain.ProviderLocal, domain.ProviderCustom:
		// valid
	default:
		return errors.New("invalid model provider")
	}
	
	// 验证模型类型
	switch c.Type {
	case domain.ModelTypeChat, domain.ModelTypeCompletion, domain.ModelTypeEmbedding, domain.ModelTypeImage, domain.ModelTypeAudio:
		// valid
	default:
		return errors.New("invalid model type")
	}
	
	if c.MaxTokens <= 0 {
		return errors.New("max tokens must be greater than 0")
	}
	
	if c.PricePerK < 0 {
		return errors.New("price per k tokens cannot be negative")
	}
	
	return nil
}

// ProcessRequestCommand 处理请求命令
type ProcessRequestCommand struct {
	application.BaseCommand
	ModelID     uuid.UUID                 `json:"model_id" binding:"required"`
	UserID      uuid.UUID                 `json:"user_id" binding:"required"`
	SessionID   uuid.UUID                 `json:"session_id"`
	RequestType string                    `json:"request_type" binding:"required"`
	Input       map[string]interface{}    `json:"input" binding:"required"`
	Metadata    map[string]interface{}    `json:"metadata"`
}

func NewProcessRequestCommand() *ProcessRequestCommand {
	return &ProcessRequestCommand{
		BaseCommand: application.BaseCommand{
			CommandID:   uuid.New(),
			CommandType: "process_request",
		},
		SessionID: uuid.New(), // 默认生成新的会话ID
		Input:     make(map[string]interface{}),
		Metadata:  make(map[string]interface{}),
	}
}

func (c *ProcessRequestCommand) Validate() error {
	if c.ModelID == uuid.Nil {
		return errors.New("model ID is required")
	}
	
	if c.UserID == uuid.Nil {
		return errors.New("user ID is required")
	}
	
	if c.RequestType == "" {
		return errors.New("request type is required")
	}
	
	if len(c.Input) == 0 {
		return errors.New("input is required")
	}
	
	// 根据请求类型验证输入
	switch c.RequestType {
	case "chat":
		if _, ok := c.Input["messages"]; !ok {
			return errors.New("messages field is required for chat requests")
		}
	case "completion":
		if _, ok := c.Input["prompt"]; !ok {
			return errors.New("prompt field is required for completion requests")
		}
	case "embedding":
		if _, ok := c.Input["text"]; !ok {
			return errors.New("text field is required for embedding requests")
		}
	}
	
	return nil
}

// UpdateModelCommand 更新模型命令
type UpdateModelCommand struct {
	application.BaseCommand
	ModelID     uuid.UUID                 `json:"model_id" binding:"required"`
	Description *string                   `json:"description"`
	Config      map[string]interface{}    `json:"config"`
	MaxTokens   *int                      `json:"max_tokens"`
	PricePerK   *float64                  `json:"price_per_k"`
	IsActive    *bool                     `json:"is_active"`
}

func NewUpdateModelCommand(modelID uuid.UUID) *UpdateModelCommand {
	return &UpdateModelCommand{
		BaseCommand: application.BaseCommand{
			CommandID:   uuid.New(),
			CommandType: "update_model",
		},
		ModelID: modelID,
	}
}

func (c *UpdateModelCommand) Validate() error {
	if c.ModelID == uuid.Nil {
		return errors.New("model ID is required")
	}
	
	if c.MaxTokens != nil && *c.MaxTokens <= 0 {
		return errors.New("max tokens must be greater than 0")
	}
	
	if c.PricePerK != nil && *c.PricePerK < 0 {
		return errors.New("price per k tokens cannot be negative")
	}
	
	return nil
}

// GetModelsQuery 获取模型查询
type GetModelsQuery struct {
	application.BaseQuery
	Provider   *domain.ModelProvider `form:"provider"`
	Type       *domain.ModelType     `form:"type"`
	IsActive   *bool                 `form:"is_active"`
	Page       int                   `form:"page,default=1"`
	PageSize   int                   `form:"page_size,default=20"`
}

func NewGetModelsQuery() *GetModelsQuery {
	return &GetModelsQuery{
		BaseQuery: application.BaseQuery{
			QueryID:   uuid.New(),
			QueryType: "get_models",
		},
		Page:     1,
		PageSize: 20,
	}
}

func (q *GetModelsQuery) Validate() error {
	if q.Page <= 0 {
		return errors.New("page must be greater than 0")
	}
	
	if q.PageSize <= 0 || q.PageSize > 100 {
		return errors.New("page size must be between 1 and 100")
	}
	
	return nil
}

// GetRequestsQuery 获取请求查询
type GetRequestsQuery struct {
	application.BaseQuery
	UserID    *uuid.UUID              `form:"user_id"`
	ModelID   *uuid.UUID              `form:"model_id"`
	SessionID *uuid.UUID              `form:"session_id"`
	Status    *domain.RequestStatus   `form:"status"`
	Page      int                     `form:"page,default=1"`
	PageSize  int                     `form:"page_size,default=20"`
}

func NewGetRequestsQuery() *GetRequestsQuery {
	return &GetRequestsQuery{
		BaseQuery: application.BaseQuery{
			QueryID:   uuid.New(),
			QueryType: "get_requests",
		},
		Page:     1,
		PageSize: 20,
	}
}

func (q *GetRequestsQuery) Validate() error {
	if q.Page <= 0 {
		return errors.New("page must be greater than 0")
	}
	
	if q.PageSize <= 0 || q.PageSize > 100 {
		return errors.New("page size must be between 1 and 100")
	}
	
	return nil
}
