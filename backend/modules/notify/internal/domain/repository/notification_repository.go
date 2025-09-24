package repository

import (
	"context"

	"github.com/noah-loop/backend/modules/notify/internal/domain"
)

// NotificationRepository 通知仓储接口
type NotificationRepository interface {
	// 基本CRUD操作
	Save(ctx context.Context, notification *domain.Notification) error
	FindByID(ctx context.Context, id string) (*domain.Notification, error)
	Update(ctx context.Context, notification *domain.Notification) error
	Delete(ctx context.Context, id string) error

	// 查询操作
	FindByStatus(ctx context.Context, status domain.NotificationStatus) ([]*domain.Notification, error)
	FindByChannel(ctx context.Context, channel domain.NotificationChannel) ([]*domain.Notification, error)
	FindByType(ctx context.Context, notifyType domain.NotificationType) ([]*domain.Notification, error)
	FindByCreatedBy(ctx context.Context, createdBy string) ([]*domain.Notification, error)
	FindByPriority(ctx context.Context, priority domain.NotificationPriority) ([]*domain.Notification, error)

	// 分页查询
	FindWithPagination(ctx context.Context, offset, limit int) ([]*domain.Notification, int64, error)
	FindByStatusWithPagination(ctx context.Context, status domain.NotificationStatus, offset, limit int) ([]*domain.Notification, int64, error)
	FindByCreatedByWithPagination(ctx context.Context, createdBy string, offset, limit int) ([]*domain.Notification, int64, error)

	// 定时任务相关
	FindScheduledNotifications(ctx context.Context, beforeTime int64) ([]*domain.Notification, error)
	FindPendingNotifications(ctx context.Context, limit int) ([]*domain.Notification, error)
	FindFailedNotifications(ctx context.Context, limit int) ([]*domain.Notification, error)
	FindRetryableNotifications(ctx context.Context, limit int) ([]*domain.Notification, error)

	// 搜索操作
	SearchByContent(ctx context.Context, query string, limit int) ([]*domain.Notification, error)
	SearchByTitle(ctx context.Context, query string, limit int) ([]*domain.Notification, error)

	// 批量操作
	SaveBatch(ctx context.Context, notifications []*domain.Notification) error
	UpdateBatch(ctx context.Context, notifications []*domain.Notification) error
	UpdateStatusBatch(ctx context.Context, ids []string, status domain.NotificationStatus) error

	// 统计操作
	CountByStatus(ctx context.Context, status domain.NotificationStatus) (int64, error)
	CountByChannel(ctx context.Context, channel domain.NotificationChannel) (int64, error)
	CountByCreatedBy(ctx context.Context, createdBy string) (int64, error)
	GetStatsByDateRange(ctx context.Context, startDate, endDate string) (*NotificationStats, error)
	GetChannelStats(ctx context.Context) ([]ChannelStats, error)

	// 清理操作
	DeleteOldNotifications(ctx context.Context, beforeTime int64) (int64, error)
	DeleteCancelledNotifications(ctx context.Context, beforeTime int64) (int64, error)
}

// NotificationStats 通知统计信息
type NotificationStats struct {
	TotalCount       int64                                      `json:"total_count"`
	StatusCounts     map[domain.NotificationStatus]int64       `json:"status_counts"`
	TypeCounts       map[domain.NotificationType]int64         `json:"type_counts"`
	ChannelCounts    map[domain.NotificationChannel]int64      `json:"channel_counts"`
	PriorityCounts   map[domain.NotificationPriority]int64     `json:"priority_counts"`
	SuccessRate      float64                                    `json:"success_rate"`
	AverageRetries   float64                                    `json:"average_retries"`
	LastHourCount    int64                                      `json:"last_hour_count"`
	LastDayCount     int64                                      `json:"last_day_count"`
}

// ChannelStats 渠道统计信息
type ChannelStats struct {
	Channel      domain.NotificationChannel `json:"channel"`
	TotalCount   int64                      `json:"total_count"`
	SuccessCount int64                      `json:"success_count"`
	FailedCount  int64                      `json:"failed_count"`
	SuccessRate  float64                    `json:"success_rate"`
	LastSentAt   string                     `json:"last_sent_at"`
}
