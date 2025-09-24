package domain

import (
	"time"
	
	"github.com/google/uuid"
	"github.com/noah-loop/backend/shared/pkg/domain"
)

// RequestStatus 请求状态
type RequestStatus string

const (
	RequestStatusPending    RequestStatus = "pending"
	RequestStatusProcessing RequestStatus = "processing"
	RequestStatusCompleted  RequestStatus = "completed"
	RequestStatusFailed     RequestStatus = "failed"
	RequestStatusCancelled  RequestStatus = "cancelled"
)

// Request 大模型请求实体
type Request struct {
	domain.BaseEntity
	ModelID      uuid.UUID              `json:"model_id" gorm:"type:uuid;not null;index"`
	UserID       uuid.UUID              `json:"user_id" gorm:"type:uuid;index"`
	SessionID    uuid.UUID              `json:"session_id" gorm:"type:uuid;index"`
	Status       RequestStatus          `json:"status" gorm:"not null;index"`
	RequestType  string                 `json:"request_type" gorm:"not null"` // chat, completion, embedding等
	Input        map[string]interface{} `json:"input" gorm:"type:jsonb;not null"`
	Output       map[string]interface{} `json:"output" gorm:"type:jsonb"`
	Metadata     map[string]interface{} `json:"metadata" gorm:"type:jsonb"`
	TokensUsed   int                    `json:"tokens_used"`
	Cost         float64                `json:"cost"`
	Duration     time.Duration          `json:"duration"`
	ErrorMessage string                 `json:"error_message"`
	
	// 关联
	Model *Model `json:"model,omitempty" gorm:"foreignKey:ModelID"`
	
	// 聚合根实现
	domainEvents []domain.DomainEvent `gorm:"-"`
}

func (r *Request) GetID() uuid.UUID {
	return r.ID
}

func (r *Request) GetVersion() int {
	return r.Version
}

func (r *Request) MarkAsModified() {
	r.UpdatedAt = time.Now()
}

// NewRequest 创建新请求
func NewRequest(modelID, userID, sessionID uuid.UUID, requestType string, input map[string]interface{}) *Request {
	request := &Request{
		BaseEntity: domain.BaseEntity{
			ID:        domain.NewEntityID(),
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		},
		ModelID:     modelID,
		UserID:      userID,
		SessionID:   sessionID,
		Status:      RequestStatusPending,
		RequestType: requestType,
		Input:       input,
		Metadata:    make(map[string]interface{}),
		domainEvents: make([]domain.DomainEvent, 0),
	}
	
	// 发布请求创建事件
	event := domain.NewDomainEvent("request.created", request.ID, request)
	request.domainEvents = append(request.domainEvents, event)
	
	return request
}

// StartProcessing 开始处理
func (r *Request) StartProcessing() error {
	if r.Status != RequestStatusPending {
		return NewRequestError("request is not in pending status")
	}
	
	r.Status = RequestStatusProcessing
	r.MarkAsModified()
	
	event := domain.NewDomainEvent("request.processing", r.ID, r.ID)
	r.domainEvents = append(r.domainEvents, event)
	
	return nil
}

// Complete 完成请求
func (r *Request) Complete(output map[string]interface{}, tokensUsed int, cost float64, duration time.Duration) error {
	if r.Status != RequestStatusProcessing {
		return NewRequestError("request is not in processing status")
	}
	
	r.Status = RequestStatusCompleted
	r.Output = output
	r.TokensUsed = tokensUsed
	r.Cost = cost
	r.Duration = duration
	r.MarkAsModified()
	
	event := domain.NewDomainEvent("request.completed", r.ID, map[string]interface{}{
		"request_id":   r.ID,
		"tokens_used":  tokensUsed,
		"cost":         cost,
		"duration":     duration,
	})
	r.domainEvents = append(r.domainEvents, event)
	
	return nil
}

// Fail 请求失败
func (r *Request) Fail(errorMessage string) error {
	if r.Status != RequestStatusProcessing && r.Status != RequestStatusPending {
		return NewRequestError("cannot fail request in current status")
	}
	
	r.Status = RequestStatusFailed
	r.ErrorMessage = errorMessage
	r.MarkAsModified()
	
	event := domain.NewDomainEvent("request.failed", r.ID, map[string]interface{}{
		"request_id": r.ID,
		"error":      errorMessage,
	})
	r.domainEvents = append(r.domainEvents, event)
	
	return nil
}

// Cancel 取消请求
func (r *Request) Cancel() error {
	if r.Status == RequestStatusCompleted || r.Status == RequestStatusFailed {
		return NewRequestError("cannot cancel completed or failed request")
	}
	
	r.Status = RequestStatusCancelled
	r.MarkAsModified()
	
	event := domain.NewDomainEvent("request.cancelled", r.ID, r.ID)
	r.domainEvents = append(r.domainEvents, event)
	
	return nil
}

// GetDomainEvents 获取领域事件
func (r *Request) GetDomainEvents() []domain.DomainEvent {
	return r.domainEvents
}

// ClearDomainEvents 清理领域事件
func (r *Request) ClearDomainEvents() {
	r.domainEvents = make([]domain.DomainEvent, 0)
}

// RequestError 请求错误
type RequestError struct {
	message string
}

func NewRequestError(message string) *RequestError {
	return &RequestError{message: message}
}

func (e *RequestError) Error() string {
	return e.message
}

// RequestRepository 请求仓储接口
type RequestRepository interface {
	domain.Repository[*Request]
	FindByModelID(ctx context.Context, modelID uuid.UUID, offset, limit int) ([]*Request, error)
	FindByUserID(ctx context.Context, userID uuid.UUID, offset, limit int) ([]*Request, error)
	FindBySessionID(ctx context.Context, sessionID uuid.UUID) ([]*Request, error)
	FindByStatus(ctx context.Context, status RequestStatus, offset, limit int) ([]*Request, error)
	GetUsageStats(ctx context.Context, userID uuid.UUID, start, end time.Time) (*UsageStats, error)
}

// UsageStats 使用统计
type UsageStats struct {
	TotalRequests int     `json:"total_requests"`
	TotalTokens   int     `json:"total_tokens"`
	TotalCost     float64 `json:"total_cost"`
	AvgDuration   float64 `json:"avg_duration"`
}
