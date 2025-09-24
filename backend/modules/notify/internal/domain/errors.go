package domain

import "fmt"

// DomainError 通知领域错误
type DomainError struct {
	Code    string `json:"code"`
	Message string `json:"message"`
	Details string `json:"details,omitempty"`
}

func (e *DomainError) Error() string {
	if e.Details != "" {
		return fmt.Sprintf("%s: %s (%s)", e.Code, e.Message, e.Details)
	}
	return fmt.Sprintf("%s: %s", e.Code, e.Message)
}

// NewDomainError 创建领域错误
func NewDomainError(code, message string) *DomainError {
	return &DomainError{
		Code:    code,
		Message: message,
	}
}

// NewDomainErrorWithDetails 创建带详情的领域错误
func NewDomainErrorWithDetails(code, message, details string) *DomainError {
	return &DomainError{
		Code:    code,
		Message: message,
		Details: details,
	}
}

// 预定义错误代码
const (
	// 通知相关错误
	ErrNotificationNotFound        = "NOTIFICATION_NOT_FOUND"
	ErrNotificationAlreadyExists   = "NOTIFICATION_ALREADY_EXISTS"
	ErrNotificationInvalidStatus   = "NOTIFICATION_INVALID_STATUS"
	ErrNotificationSendFailed      = "NOTIFICATION_SEND_FAILED"
	ErrNotificationCancelled       = "NOTIFICATION_CANCELLED"
	ErrNotificationExpired         = "NOTIFICATION_EXPIRED"

	// 模板相关错误
	ErrTemplateNotFound            = "TEMPLATE_NOT_FOUND"
	ErrTemplateAlreadyExists       = "TEMPLATE_ALREADY_EXISTS"
	ErrTemplateInvalidFormat       = "TEMPLATE_INVALID_FORMAT"
	ErrTemplateRenderFailed        = "TEMPLATE_RENDER_FAILED"
	ErrTemplateMissingVariable     = "TEMPLATE_MISSING_VARIABLE"

	// 渠道相关错误
	ErrChannelNotFound             = "CHANNEL_NOT_FOUND"
	ErrChannelDisabled             = "CHANNEL_DISABLED"
	ErrChannelConfigInvalid        = "CHANNEL_CONFIG_INVALID"
	ErrChannelRateLimitExceeded    = "CHANNEL_RATE_LIMIT_EXCEEDED"
	ErrChannelConnectionFailed     = "CHANNEL_CONNECTION_FAILED"

	// 接收者相关错误
	ErrRecipientNotFound           = "RECIPIENT_NOT_FOUND"
	ErrRecipientInvalidAddress     = "RECIPIENT_INVALID_ADDRESS"
	ErrRecipientDeliveryFailed     = "RECIPIENT_DELIVERY_FAILED"

	// 验证相关错误
	ErrInvalidEmail                = "INVALID_EMAIL"
	ErrInvalidPhone                = "INVALID_PHONE"
	ErrInvalidTemplate             = "INVALID_TEMPLATE"
	ErrInvalidChannel              = "INVALID_CHANNEL"
	ErrInvalidPriority             = "INVALID_PRIORITY"

	// 权限相关错误
	ErrPermissionDenied            = "PERMISSION_DENIED"
	ErrUnauthorized                = "UNAUTHORIZED"
	ErrForbidden                   = "FORBIDDEN"

	// 系统相关错误
	ErrInternalError               = "INTERNAL_ERROR"
	ErrServiceUnavailable          = "SERVICE_UNAVAILABLE"
	ErrTimeout                     = "TIMEOUT"
	ErrResourceExhausted           = "RESOURCE_EXHAUSTED"

	// 配置相关错误
	ErrMissingConfig               = "MISSING_CONFIG"
	ErrInvalidConfig               = "INVALID_CONFIG"
	ErrConfigNotFound              = "CONFIG_NOT_FOUND"
)

// 常用错误创建函数
func ErrNotificationNotFoundf(notificationID string) *DomainError {
	return NewDomainErrorWithDetails(ErrNotificationNotFound, "Notification not found", fmt.Sprintf("notification_id: %s", notificationID))
}

func ErrTemplateNotFoundf(templateID string) *DomainError {
	return NewDomainErrorWithDetails(ErrTemplateNotFound, "Template not found", fmt.Sprintf("template_id: %s", templateID))
}

func ErrChannelNotFoundf(channel string) *DomainError {
	return NewDomainErrorWithDetails(ErrChannelNotFound, "Channel not found", fmt.Sprintf("channel: %s", channel))
}

func ErrRecipientNotFoundf(recipientID string) *DomainError {
	return NewDomainErrorWithDetails(ErrRecipientNotFound, "Recipient not found", fmt.Sprintf("recipient_id: %s", recipientID))
}

func ErrInvalidEmailf(email string) *DomainError {
	return NewDomainErrorWithDetails(ErrInvalidEmail, "Invalid email address", fmt.Sprintf("email: %s", email))
}

func ErrInvalidPhonef(phone string) *DomainError {
	return NewDomainErrorWithDetails(ErrInvalidPhone, "Invalid phone number", fmt.Sprintf("phone: %s", phone))
}

func ErrChannelDisabledf(channel string) *DomainError {
	return NewDomainErrorWithDetails(ErrChannelDisabled, "Channel is disabled", fmt.Sprintf("channel: %s", channel))
}

func ErrMissingConfigf(field string) *DomainError {
	return NewDomainErrorWithDetails(ErrMissingConfig, "Missing required configuration", fmt.Sprintf("field: %s", field))
}

func ErrRateLimitExceededf(channel string, limit string) *DomainError {
	return NewDomainErrorWithDetails(ErrChannelRateLimitExceeded, "Rate limit exceeded", fmt.Sprintf("channel: %s, limit: %s", channel, limit))
}
