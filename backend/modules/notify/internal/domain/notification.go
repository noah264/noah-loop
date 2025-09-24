package domain

import (
	"time"

	"github.com/noah-loop/backend/shared/pkg/domain"
)

// NotificationStatus 通知状态
type NotificationStatus string

const (
	NotificationStatusPending   NotificationStatus = "pending"   // 待发送
	NotificationStatusSending   NotificationStatus = "sending"   // 发送中
	NotificationStatusSent      NotificationStatus = "sent"      // 已发送
	NotificationStatusDelivered NotificationStatus = "delivered" // 已送达
	NotificationStatusFailed    NotificationStatus = "failed"    // 发送失败
	NotificationStatusCancelled NotificationStatus = "cancelled" // 已取消
)

// NotificationPriority 通知优先级
type NotificationPriority string

const (
	NotificationPriorityLow    NotificationPriority = "low"    // 低优先级
	NotificationPriorityNormal NotificationPriority = "normal" // 普通优先级
	NotificationPriorityHigh   NotificationPriority = "high"   // 高优先级
	NotificationPriorityUrgent NotificationPriority = "urgent" // 紧急
)

// NotificationType 通知类型
type NotificationType string

const (
	NotificationTypeSystem     NotificationType = "system"     // 系统通知
	NotificationTypeMarketing  NotificationType = "marketing"  // 营销通知
	NotificationTypeReminder   NotificationType = "reminder"   // 提醒通知
	NotificationTypeAlert      NotificationType = "alert"      // 警报通知
	NotificationTypeVerify     NotificationType = "verify"     // 验证通知
	NotificationTypeWorkflow   NotificationType = "workflow"   // 工作流通知
)

// Notification 通知聚合根
type Notification struct {
	domain.Entity
	Title            string               `gorm:"not null" json:"title"`
	Content          string               `gorm:"type:text;not null" json:"content"`
	Type             NotificationType     `gorm:"not null" json:"type"`
	Priority         NotificationPriority `gorm:"not null;default:'normal'" json:"priority"`
	Status           NotificationStatus   `gorm:"not null;default:'pending'" json:"status"`
	Channel          NotificationChannel  `gorm:"not null" json:"channel"`
	Recipients       []Recipient          `json:"recipients"`
	TemplateID       string               `gorm:"index" json:"template_id,omitempty"`
	Variables        map[string]string    `gorm:"serializer:json" json:"variables,omitempty"`
	Metadata         NotificationMetadata `gorm:"embedded" json:"metadata"`
	ScheduledAt      *time.Time           `json:"scheduled_at,omitempty"`
	SentAt           *time.Time           `json:"sent_at,omitempty"`
	DeliveredAt      *time.Time           `json:"delivered_at,omitempty"`
	FailedAt         *time.Time           `json:"failed_at,omitempty"`
	ErrorMessage     string               `json:"error_message,omitempty"`
	RetryCount       int                  `json:"retry_count"`
	MaxRetries       int                  `gorm:"default:3" json:"max_retries"`
	CreatedBy        string               `gorm:"index" json:"created_by"`
	CreatedAt        time.Time            `json:"created_at"`
	UpdatedAt        time.Time            `json:"updated_at"`
}

// NotificationMetadata 通知元数据
type NotificationMetadata struct {
	Source          string            `json:"source,omitempty"`          // 来源系统
	Reference       string            `json:"reference,omitempty"`       // 关联引用
	Tags            []string          `gorm:"serializer:json" json:"tags,omitempty"`
	Category        string            `json:"category,omitempty"`        // 分类
	TrackingID      string            `json:"tracking_id,omitempty"`     // 跟踪ID
	ExternalID      string            `json:"external_id,omitempty"`     // 外部ID
	Custom          map[string]string `gorm:"serializer:json" json:"custom,omitempty"`
}

// UpdateStatus 更新通知状态
func (n *Notification) UpdateStatus(status NotificationStatus) error {
	if !n.isValidStatusTransition(n.Status, status) {
		return NewDomainError("INVALID_STATUS_TRANSITION", "invalid status transition")
	}
	
	n.Status = status
	n.UpdatedAt = time.Now()
	
	switch status {
	case NotificationStatusSent:
		now := time.Now()
		n.SentAt = &now
	case NotificationStatusDelivered:
		now := time.Now()
		n.DeliveredAt = &now
	case NotificationStatusFailed:
		now := time.Now()
		n.FailedAt = &now
		n.RetryCount++
	}
	
	return nil
}

// CanRetry 是否可以重试
func (n *Notification) CanRetry() bool {
	return n.Status == NotificationStatusFailed && n.RetryCount < n.MaxRetries
}

// IsScheduled 是否为定时通知
func (n *Notification) IsScheduled() bool {
	return n.ScheduledAt != nil && n.ScheduledAt.After(time.Now())
}

// ShouldSend 是否应该发送
func (n *Notification) ShouldSend() bool {
	if n.Status != NotificationStatusPending {
		return false
	}
	
	if n.ScheduledAt != nil {
		return !n.ScheduledAt.After(time.Now())
	}
	
	return true
}

// AddRecipient 添加接收者
func (n *Notification) AddRecipient(recipient Recipient) {
	n.Recipients = append(n.Recipients, recipient)
	n.UpdatedAt = time.Now()
}

// SetError 设置错误信息
func (n *Notification) SetError(err error) {
	n.ErrorMessage = err.Error()
	n.UpdateStatus(NotificationStatusFailed)
}

// isValidStatusTransition 检查状态转换是否有效
func (n *Notification) isValidStatusTransition(from, to NotificationStatus) bool {
	validTransitions := map[NotificationStatus][]NotificationStatus{
		NotificationStatusPending: {NotificationStatusSending, NotificationStatusCancelled},
		NotificationStatusSending: {NotificationStatusSent, NotificationStatusFailed},
		NotificationStatusSent:    {NotificationStatusDelivered, NotificationStatusFailed},
		NotificationStatusFailed:  {NotificationStatusSending}, // 可以重试
		NotificationStatusDelivered: {}, // 终态
		NotificationStatusCancelled: {}, // 终态
	}
	
	allowedStates, exists := validTransitions[from]
	if !exists {
		return false
	}
	
	for _, allowed := range allowedStates {
		if allowed == to {
			return true
		}
	}
	
	return false
}

// NewNotification 创建新通知
func NewNotification(title, content string, notifyType NotificationType, channel NotificationChannel, createdBy string) (*Notification, error) {
	if title == "" {
		return nil, NewDomainError("INVALID_TITLE", "notification title cannot be empty")
	}
	
	if content == "" {
		return nil, NewDomainError("INVALID_CONTENT", "notification content cannot be empty")
	}
	
	if createdBy == "" {
		return nil, NewDomainError("INVALID_CREATOR", "creator cannot be empty")
	}
	
	notification := &Notification{
		Entity:      domain.NewEntity(),
		Title:       title,
		Content:     content,
		Type:        notifyType,
		Priority:    NotificationPriorityNormal,
		Status:      NotificationStatusPending,
		Channel:     channel,
		Recipients:  make([]Recipient, 0),
		Variables:   make(map[string]string),
		Metadata: NotificationMetadata{
			Custom: make(map[string]string),
		},
		RetryCount:  0,
		MaxRetries:  3,
		CreatedBy:   createdBy,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}
	
	return notification, nil
}
