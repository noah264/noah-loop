package domain

import (
	"regexp"
	"strings"
	"time"

	"github.com/noah-loop/backend/shared/pkg/domain"
)

// RecipientType 接收者类型
type RecipientType string

const (
	RecipientTypeUser   RecipientType = "user"   // 用户
	RecipientTypeGroup  RecipientType = "group"  // 用户组
	RecipientTypeRole   RecipientType = "role"   // 角色
	RecipientTypeEmail  RecipientType = "email"  // 邮箱地址
	RecipientTypePhone  RecipientType = "phone"  // 手机号
	RecipientTypeDevice RecipientType = "device" // 设备
)

// Recipient 通知接收者实体
type Recipient struct {
	domain.Entity
	NotificationID string            `gorm:"not null;index" json:"notification_id"`
	Type           RecipientType     `gorm:"not null" json:"type"`
	Identifier     string            `gorm:"not null" json:"identifier"` // 接收者标识（用户ID、邮箱、手机号等）
	Name           string            `json:"name"`                       // 接收者名称
	Channel        NotificationChannel `gorm:"not null" json:"channel"`
	Address        string            `json:"address"`                    // 接收地址（邮箱、手机号等）
	Variables      map[string]string `gorm:"serializer:json" json:"variables,omitempty"` // 个性化变量
	Status         RecipientStatus   `gorm:"not null;default:'pending'" json:"status"`
	SentAt         *time.Time        `json:"sent_at,omitempty"`
	DeliveredAt    *time.Time        `json:"delivered_at,omitempty"`
	FailedAt       *time.Time        `json:"failed_at,omitempty"`
	ErrorMessage   string            `json:"error_message,omitempty"`
	RetryCount     int               `json:"retry_count"`
	CreatedAt      time.Time         `json:"created_at"`
	UpdatedAt      time.Time         `json:"updated_at"`
}

// RecipientStatus 接收者状态
type RecipientStatus string

const (
	RecipientStatusPending   RecipientStatus = "pending"   // 待发送
	RecipientStatusSending   RecipientStatus = "sending"   // 发送中
	RecipientStatusSent      RecipientStatus = "sent"      // 已发送
	RecipientStatusDelivered RecipientStatus = "delivered" // 已送达
	RecipientStatusFailed    RecipientStatus = "failed"    // 发送失败
	RecipientStatusSkipped   RecipientStatus = "skipped"   // 跳过
)

// UpdateStatus 更新接收者状态
func (r *Recipient) UpdateStatus(status RecipientStatus) error {
	r.Status = status
	r.UpdatedAt = time.Now()
	
	switch status {
	case RecipientStatusSent:
		now := time.Now()
		r.SentAt = &now
	case RecipientStatusDelivered:
		now := time.Now()
		r.DeliveredAt = &now
	case RecipientStatusFailed:
		now := time.Now()
		r.FailedAt = &now
		r.RetryCount++
	}
	
	return nil
}

// SetError 设置错误信息
func (r *Recipient) SetError(err error) {
	r.ErrorMessage = err.Error()
	r.UpdateStatus(RecipientStatusFailed)
}

// IsValid 验证接收者信息是否有效
func (r *Recipient) IsValid() error {
	if r.Identifier == "" {
		return NewDomainError("INVALID_IDENTIFIER", "recipient identifier cannot be empty")
	}
	
	switch r.Type {
	case RecipientTypeEmail:
		if !isValidEmail(r.Address) {
			return NewDomainError("INVALID_EMAIL", "invalid email address")
		}
	case RecipientTypePhone:
		if !isValidPhone(r.Address) {
			return NewDomainError("INVALID_PHONE", "invalid phone number")
		}
	}
	
	return nil
}

// GetEffectiveAddress 获取有效的接收地址
func (r *Recipient) GetEffectiveAddress() string {
	if r.Address != "" {
		return r.Address
	}
	
	// 如果没有地址，尝试从标识符中提取
	switch r.Type {
	case RecipientTypeEmail:
		if isValidEmail(r.Identifier) {
			return r.Identifier
		}
	case RecipientTypePhone:
		if isValidPhone(r.Identifier) {
			return r.Identifier
		}
	}
	
	return r.Identifier
}

// NewRecipient 创建新接收者
func NewRecipient(notificationID string, recipientType RecipientType, identifier string, channel NotificationChannel) (*Recipient, error) {
	if notificationID == "" {
		return nil, NewDomainError("INVALID_NOTIFICATION_ID", "notification ID cannot be empty")
	}
	
	if identifier == "" {
		return nil, NewDomainError("INVALID_IDENTIFIER", "recipient identifier cannot be empty")
	}
	
	recipient := &Recipient{
		Entity:         domain.NewEntity(),
		NotificationID: notificationID,
		Type:           recipientType,
		Identifier:     identifier,
		Channel:        channel,
		Variables:      make(map[string]string),
		Status:         RecipientStatusPending,
		RetryCount:     0,
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
	}
	
	// 自动设置地址
	switch recipientType {
	case RecipientTypeEmail:
		if isValidEmail(identifier) {
			recipient.Address = identifier
		}
	case RecipientTypePhone:
		if isValidPhone(identifier) {
			recipient.Address = identifier
		}
	}
	
	// 验证接收者信息
	if err := recipient.IsValid(); err != nil {
		return nil, err
	}
	
	return recipient, nil
}

// 验证邮箱地址
func isValidEmail(email string) bool {
	if email == "" {
		return false
	}
	
	// 简单的邮箱正则验证
	pattern := `^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`
	matched, _ := regexp.MatchString(pattern, email)
	return matched
}

// 验证手机号
func isValidPhone(phone string) bool {
	if phone == "" {
		return false
	}
	
	// 移除所有非数字字符
	cleaned := regexp.MustCompile(`[^\d]`).ReplaceAllString(phone, "")
	
	// 简单验证：长度应该在10-15位之间
	return len(cleaned) >= 10 && len(cleaned) <= 15
}

// FormatPhone 格式化手机号
func FormatPhone(phone string) string {
	// 移除所有非数字字符
	cleaned := regexp.MustCompile(`[^\d]`).ReplaceAllString(phone, "")
	
	// 如果是中国手机号，确保以+86开头
	if len(cleaned) == 11 && strings.HasPrefix(cleaned, "1") {
		return "+86" + cleaned
	}
	
	// 如果已经包含国家代码
	if len(cleaned) > 11 {
		return "+" + cleaned
	}
	
	return cleaned
}
