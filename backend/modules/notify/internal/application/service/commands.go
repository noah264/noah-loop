package service

import (
	"time"

	"github.com/noah-loop/backend/modules/notify/internal/domain"
)

// CreateNotificationCommand 创建通知命令
type CreateNotificationCommand struct {
	Title       string                        `json:"title" binding:"required"`
	Content     string                        `json:"content" binding:"required"`
	Type        domain.NotificationType       `json:"type" binding:"required"`
	Channel     domain.NotificationChannel    `json:"channel" binding:"required"`
	Priority    domain.NotificationPriority   `json:"priority,omitempty"`
	TemplateID  string                        `json:"template_id,omitempty"`
	Variables   map[string]string             `json:"variables,omitempty"`
	Recipients  []CreateRecipientCommand      `json:"recipients" binding:"required"`
	Metadata    *domain.NotificationMetadata  `json:"metadata,omitempty"`
	ScheduledAt *time.Time                    `json:"scheduled_at,omitempty"`
	MaxRetries  int                           `json:"max_retries,omitempty"`
	CreatedBy   string                        `json:"created_by" binding:"required"`
}

// CreateRecipientCommand 创建接收者命令
type CreateRecipientCommand struct {
	Type       domain.RecipientType  `json:"type" binding:"required"`
	Identifier string                `json:"identifier" binding:"required"`
	Name       string                `json:"name,omitempty"`
	Address    string                `json:"address,omitempty"`
	Variables  map[string]string     `json:"variables,omitempty"`
}

// CreateNotificationFromTemplateCommand 从模板创建通知命令
type CreateNotificationFromTemplateCommand struct {
	TemplateID  string                        `json:"template_id" binding:"required"`
	Type        domain.NotificationType       `json:"type" binding:"required"`
	Channel     domain.NotificationChannel    `json:"channel" binding:"required"`
	Priority    domain.NotificationPriority   `json:"priority,omitempty"`
	Variables   map[string]string             `json:"variables,omitempty"`
	Recipients  []CreateRecipientCommand      `json:"recipients" binding:"required"`
	Metadata    *domain.NotificationMetadata  `json:"metadata,omitempty"`
	ScheduledAt *time.Time                    `json:"scheduled_at,omitempty"`
	MaxRetries  int                           `json:"max_retries,omitempty"`
	CreatedBy   string                        `json:"created_by" binding:"required"`
}

// SendNotificationCommand 发送通知命令
type SendNotificationCommand struct {
	NotificationID string `json:"notification_id" binding:"required"`
}

// BatchSendNotificationCommand 批量发送通知命令
type BatchSendNotificationCommand struct {
	NotificationIDs []string `json:"notification_ids" binding:"required"`
}

// ListNotificationsCommand 列出通知命令
type ListNotificationsCommand struct {
	Status    string `json:"status,omitempty"`
	Type      string `json:"type,omitempty"`
	Channel   string `json:"channel,omitempty"`
	CreatedBy string `json:"created_by,omitempty"`
	Offset    int    `json:"offset"`
	Limit     int    `json:"limit"`
}

// GetNotificationCommand 获取通知命令
type GetNotificationCommand struct {
	ID               string `json:"id" binding:"required"`
	IncludeRecipients bool   `json:"include_recipients"`
}

// UpdateNotificationCommand 更新通知命令
type UpdateNotificationCommand struct {
	ID          string                       `json:"id" binding:"required"`
	Title       string                       `json:"title,omitempty"`
	Content     string                       `json:"content,omitempty"`
	Priority    domain.NotificationPriority  `json:"priority,omitempty"`
	ScheduledAt *time.Time                   `json:"scheduled_at,omitempty"`
	MaxRetries  int                          `json:"max_retries,omitempty"`
}

// CancelNotificationCommand 取消通知命令
type CancelNotificationCommand struct {
	ID string `json:"id" binding:"required"`
}

// RetryNotificationCommand 重试通知命令
type RetryNotificationCommand struct {
	ID string `json:"id" binding:"required"`
}

// GetNotificationStatsCommand 获取通知统计命令
type GetNotificationStatsCommand struct {
	StartDate string `json:"start_date,omitempty"`
	EndDate   string `json:"end_date,omitempty"`
}

// CreateTemplateCommand 创建模板命令
type CreateTemplateCommand struct {
	Name        string                `json:"name" binding:"required"`
	Code        string                `json:"code" binding:"required"`
	Type        domain.TemplateType   `json:"type" binding:"required"`
	Category    string                `json:"category,omitempty"`
	Description string                `json:"description,omitempty"`
	Subject     string                `json:"subject,omitempty"`
	Content     string                `json:"content" binding:"required"`
	Variables   []TemplateVariableCmd `json:"variables,omitempty"`
	Tags        []string              `json:"tags,omitempty"`
	CreatedBy   string                `json:"created_by" binding:"required"`
}

// TemplateVariableCmd 模板变量命令
type TemplateVariableCmd struct {
	Name         string `json:"name" binding:"required"`
	DisplayName  string `json:"display_name,omitempty"`
	Type         string `json:"type,omitempty"`
	DefaultValue string `json:"default_value,omitempty"`
	Required     bool   `json:"required"`
	Description  string `json:"description,omitempty"`
	Validation   string `json:"validation,omitempty"`
}

// UpdateTemplateCommand 更新模板命令
type UpdateTemplateCommand struct {
	ID          string                `json:"id" binding:"required"`
	Name        string                `json:"name,omitempty"`
	Category    string                `json:"category,omitempty"`
	Description string                `json:"description,omitempty"`
	Status      domain.TemplateStatus `json:"status,omitempty"`
	Tags        []string              `json:"tags,omitempty"`
}

// CreateTemplateVersionCommand 创建模板版本命令
type CreateTemplateVersionCommand struct {
	TemplateID string `json:"template_id" binding:"required"`
	Version    string `json:"version" binding:"required"`
	Subject    string `json:"subject,omitempty"`
	Content    string `json:"content" binding:"required"`
	IsActive   bool   `json:"is_active"`
	ChangeLog  string `json:"change_log,omitempty"`
	CreatedBy  string `json:"created_by" binding:"required"`
}

// RenderTemplateCommand 渲染模板命令
type RenderTemplateCommand struct {
	TemplateID string                     `json:"template_id" binding:"required"`
	Channel    domain.NotificationChannel `json:"channel" binding:"required"`
	Variables  map[string]string          `json:"variables,omitempty"`
}

// ListTemplatesCommand 列出模板命令
type ListTemplatesCommand struct {
	Status    string `json:"status,omitempty"`
	Type      string `json:"type,omitempty"`
	Category  string `json:"category,omitempty"`
	CreatedBy string `json:"created_by,omitempty"`
	Offset    int    `json:"offset"`
	Limit     int    `json:"limit"`
}

// SearchTemplatesCommand 搜索模板命令
type SearchTemplatesCommand struct {
	Query  string `json:"query" binding:"required"`
	Limit  int    `json:"limit"`
}

// CreateChannelConfigCommand 创建渠道配置命令
type CreateChannelConfigCommand struct {
	Channel     domain.NotificationChannel `json:"channel" binding:"required"`
	Name        string                     `json:"name" binding:"required"`
	Description string                     `json:"description,omitempty"`
	Config      map[string]string          `json:"config" binding:"required"`
	OwnerID     string                     `json:"owner_id" binding:"required"`
}

// UpdateChannelConfigCommand 更新渠道配置命令
type UpdateChannelConfigCommand struct {
	ID          string            `json:"id" binding:"required"`
	Name        string            `json:"name,omitempty"`
	Description string            `json:"description,omitempty"`
	Config      map[string]string `json:"config,omitempty"`
	IsEnabled   *bool             `json:"is_enabled,omitempty"`
}

// TestChannelCommand 测试渠道命令
type TestChannelCommand struct {
	ChannelID string `json:"channel_id" binding:"required"`
	TestData  map[string]string `json:"test_data,omitempty"`
}

// ListChannelConfigsCommand 列出渠道配置命令
type ListChannelConfigsCommand struct {
	Channel   string `json:"channel,omitempty"`
	OwnerID   string `json:"owner_id,omitempty"`
	IsEnabled *bool  `json:"is_enabled,omitempty"`
	Offset    int    `json:"offset"`
	Limit     int    `json:"limit"`
}

// BatchCreateNotificationsCommand 批量创建通知命令
type BatchCreateNotificationsCommand struct {
	Notifications []CreateNotificationCommand `json:"notifications" binding:"required"`
}

// BatchCreateFromTemplateCommand 批量从模板创建通知命令
type BatchCreateFromTemplateCommand struct {
	TemplateID  string                      `json:"template_id" binding:"required"`
	Recipients  []CreateRecipientCommand    `json:"recipients" binding:"required"`
	Type        domain.NotificationType     `json:"type" binding:"required"`
	Channel     domain.NotificationChannel  `json:"channel" binding:"required"`
	Priority    domain.NotificationPriority `json:"priority,omitempty"`
	Variables   map[string]string           `json:"variables,omitempty"`
	ScheduledAt *time.Time                  `json:"scheduled_at,omitempty"`
	CreatedBy   string                      `json:"created_by" binding:"required"`
}
