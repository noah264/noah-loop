package repository

import (
	"context"

	"github.com/noah-loop/backend/modules/notify/internal/domain"
)

// TemplateRepository 模板仓储接口
type TemplateRepository interface {
	// 基本CRUD操作
	Save(ctx context.Context, template *domain.NotificationTemplate) error
	FindByID(ctx context.Context, id string) (*domain.NotificationTemplate, error)
	FindByCode(ctx context.Context, code string) (*domain.NotificationTemplate, error)
	Update(ctx context.Context, template *domain.NotificationTemplate) error
	Delete(ctx context.Context, id string) error

	// 查询操作
	FindByStatus(ctx context.Context, status domain.TemplateStatus) ([]*domain.NotificationTemplate, error)
	FindByType(ctx context.Context, templateType domain.TemplateType) ([]*domain.NotificationTemplate, error)
	FindByCreatedBy(ctx context.Context, createdBy string) ([]*domain.NotificationTemplate, error)
	FindByCategory(ctx context.Context, category string) ([]*domain.NotificationTemplate, error)
	FindByTags(ctx context.Context, tags []string) ([]*domain.NotificationTemplate, error)

	// 分页查询
	FindWithPagination(ctx context.Context, offset, limit int) ([]*domain.NotificationTemplate, int64, error)
	FindByStatusWithPagination(ctx context.Context, status domain.TemplateStatus, offset, limit int) ([]*domain.NotificationTemplate, int64, error)
	FindByCreatedByWithPagination(ctx context.Context, createdBy string, offset, limit int) ([]*domain.NotificationTemplate, int64, error)

	// 搜索操作
	SearchByName(ctx context.Context, query string, limit int) ([]*domain.NotificationTemplate, error)
	SearchByContent(ctx context.Context, query string, limit int) ([]*domain.NotificationTemplate, error)

	// 版本管理
	SaveVersion(ctx context.Context, version *domain.TemplateVersion) error
	FindVersionsByTemplateID(ctx context.Context, templateID string) ([]*domain.TemplateVersion, error)
	FindActiveVersion(ctx context.Context, templateID string) (*domain.TemplateVersion, error)
	UpdateVersionStatus(ctx context.Context, templateID, version string, isActive bool) error

	// 渠道模板
	SaveChannelTemplate(ctx context.Context, channelTemplate *domain.TemplateChannel) error
	FindChannelTemplates(ctx context.Context, templateID string) ([]*domain.TemplateChannel, error)
	FindChannelTemplate(ctx context.Context, templateID string, channel domain.NotificationChannel) (*domain.TemplateChannel, error)

	// 变量管理
	SaveVariables(ctx context.Context, variables []*domain.TemplateVariable) error
	FindVariablesByTemplateID(ctx context.Context, templateID string) ([]*domain.TemplateVariable, error)
	DeleteVariablesByTemplateID(ctx context.Context, templateID string) error

	// 批量操作
	SaveBatch(ctx context.Context, templates []*domain.NotificationTemplate) error
	UpdateBatch(ctx context.Context, templates []*domain.NotificationTemplate) error
	UpdateStatusBatch(ctx context.Context, ids []string, status domain.TemplateStatus) error

	// 统计操作
	CountByStatus(ctx context.Context, status domain.TemplateStatus) (int64, error)
	CountByType(ctx context.Context, templateType domain.TemplateType) (int64, error)
	CountByCreatedBy(ctx context.Context, createdBy string) (int64, error)
	GetUsageStats(ctx context.Context, templateID string) (*TemplateUsageStats, error)

	// 清理操作
	DeleteArchivedTemplates(ctx context.Context, beforeTime int64) (int64, error)
	CleanupOrphanedVersions(ctx context.Context) (int64, error)
}

// TemplateUsageStats 模板使用统计
type TemplateUsageStats struct {
	TemplateID      string                                     `json:"template_id"`
	TotalUsage      int64                                      `json:"total_usage"`
	ChannelUsage    map[domain.NotificationChannel]int64      `json:"channel_usage"`
	SuccessCount    int64                                      `json:"success_count"`
	FailedCount     int64                                      `json:"failed_count"`
	LastUsedAt      string                                     `json:"last_used_at"`
	AverageRenderTime float64                                  `json:"average_render_time"`
}
