package repository

import (
	"context"

	"github.com/noah-loop/backend/modules/notify/internal/domain"
)

// ChannelRepository 渠道仓储接口
type ChannelRepository interface {
	// 基本CRUD操作
	Save(ctx context.Context, config *domain.ChannelConfig) error
	FindByID(ctx context.Context, id string) (*domain.ChannelConfig, error)
	FindByChannelAndOwner(ctx context.Context, channel domain.NotificationChannel, ownerID string) (*domain.ChannelConfig, error)
	Update(ctx context.Context, config *domain.ChannelConfig) error
	Delete(ctx context.Context, id string) error

	// 查询操作
	FindByChannel(ctx context.Context, channel domain.NotificationChannel) ([]*domain.ChannelConfig, error)
	FindByOwner(ctx context.Context, ownerID string) ([]*domain.ChannelConfig, error)
	FindEnabledChannels(ctx context.Context) ([]*domain.ChannelConfig, error)
	FindEnabledChannelsByOwner(ctx context.Context, ownerID string) ([]*domain.ChannelConfig, error)

	// 分页查询
	FindWithPagination(ctx context.Context, offset, limit int) ([]*domain.ChannelConfig, int64, error)
	FindByOwnerWithPagination(ctx context.Context, ownerID string, offset, limit int) ([]*domain.ChannelConfig, int64, error)

	// 搜索操作
	SearchByName(ctx context.Context, query string, limit int) ([]*domain.ChannelConfig, error)

	// 批量操作
	SaveBatch(ctx context.Context, configs []*domain.ChannelConfig) error
	UpdateBatch(ctx context.Context, configs []*domain.ChannelConfig) error
	EnableBatch(ctx context.Context, ids []string) error
	DisableBatch(ctx context.Context, ids []string) error

	// 统计操作
	CountByChannel(ctx context.Context, channel domain.NotificationChannel) (int64, error)
	CountByOwner(ctx context.Context, ownerID string) (int64, error)
	CountEnabledChannels(ctx context.Context) (int64, error)

	// 配置验证
	ValidateChannelConfig(ctx context.Context, channel domain.NotificationChannel, config map[string]string) error
	TestChannelConnection(ctx context.Context, configID string) error
}

// RecipientRepository 接收者仓储接口
type RecipientRepository interface {
	// 基本CRUD操作
	Save(ctx context.Context, recipient *domain.Recipient) error
	FindByID(ctx context.Context, id string) (*domain.Recipient, error)
	Update(ctx context.Context, recipient *domain.Recipient) error
	Delete(ctx context.Context, id string) error

	// 查询操作
	FindByNotificationID(ctx context.Context, notificationID string) ([]*domain.Recipient, error)
	FindByStatus(ctx context.Context, status domain.RecipientStatus) ([]*domain.Recipient, error)
	FindByType(ctx context.Context, recipientType domain.RecipientType) ([]*domain.Recipient, error)
	FindByChannel(ctx context.Context, channel domain.NotificationChannel) ([]*domain.Recipient, error)

	// 分页查询
	FindWithPagination(ctx context.Context, offset, limit int) ([]*domain.Recipient, int64, error)
	FindByNotificationIDWithPagination(ctx context.Context, notificationID string, offset, limit int) ([]*domain.Recipient, int64, error)

	// 批量操作
	SaveBatch(ctx context.Context, recipients []*domain.Recipient) error
	UpdateBatch(ctx context.Context, recipients []*domain.Recipient) error
	UpdateStatusBatch(ctx context.Context, ids []string, status domain.RecipientStatus) error
	DeleteByNotificationID(ctx context.Context, notificationID string) error

	// 统计操作
	CountByNotificationID(ctx context.Context, notificationID string) (int64, error)
	CountByStatus(ctx context.Context, status domain.RecipientStatus) (int64, error)
	CountByChannel(ctx context.Context, channel domain.NotificationChannel) (int64, error)

	// 地址验证
	ValidateEmailAddress(ctx context.Context, email string) (bool, error)
	ValidatePhoneNumber(ctx context.Context, phone string) (bool, error)
	FormatPhoneNumber(ctx context.Context, phone, countryCode string) (string, error)
}
