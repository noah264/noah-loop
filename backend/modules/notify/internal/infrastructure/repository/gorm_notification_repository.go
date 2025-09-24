package repository

import (
	"context"

	"github.com/noah-loop/backend/modules/notify/internal/domain"
	"github.com/noah-loop/backend/modules/notify/internal/domain/repository"
	"gorm.io/gorm"
)

// GormNotificationRepository GORM通知仓储实现
type GormNotificationRepository struct {
	db *gorm.DB
}

// NewGormNotificationRepository 创建GORM通知仓储
func NewGormNotificationRepository(db *gorm.DB) repository.NotificationRepository {
	return &GormNotificationRepository{
		db: db,
	}
}

// Save 保存通知
func (r *GormNotificationRepository) Save(ctx context.Context, notification *domain.Notification) error {
	return r.db.WithContext(ctx).Create(notification).Error
}

// FindByID 根据ID查找通知
func (r *GormNotificationRepository) FindByID(ctx context.Context, id string) (*domain.Notification, error) {
	var notification domain.Notification
	err := r.db.WithContext(ctx).
		Preload("Recipients").
		First(&notification, "id = ?", id).Error
	
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}
	
	return &notification, nil
}

// Update 更新通知
func (r *GormNotificationRepository) Update(ctx context.Context, notification *domain.Notification) error {
	return r.db.WithContext(ctx).Save(notification).Error
}

// Delete 删除通知
func (r *GormNotificationRepository) Delete(ctx context.Context, id string) error {
	return r.db.WithContext(ctx).Delete(&domain.Notification{}, "id = ?", id).Error
}

// FindByStatus 根据状态查找通知
func (r *GormNotificationRepository) FindByStatus(ctx context.Context, status domain.NotificationStatus) ([]*domain.Notification, error) {
	var notifications []*domain.Notification
	err := r.db.WithContext(ctx).
		Where("status = ?", status).
		Order("created_at DESC").
		Find(&notifications).Error
	
	return notifications, err
}

// FindByChannel 根据渠道查找通知
func (r *GormNotificationRepository) FindByChannel(ctx context.Context, channel domain.NotificationChannel) ([]*domain.Notification, error) {
	var notifications []*domain.Notification
	err := r.db.WithContext(ctx).
		Where("channel = ?", channel).
		Order("created_at DESC").
		Find(&notifications).Error
	
	return notifications, err
}

// FindByType 根据类型查找通知
func (r *GormNotificationRepository) FindByType(ctx context.Context, notifyType domain.NotificationType) ([]*domain.Notification, error) {
	var notifications []*domain.Notification
	err := r.db.WithContext(ctx).
		Where("type = ?", notifyType).
		Order("created_at DESC").
		Find(&notifications).Error
	
	return notifications, err
}

// FindByCreatedBy 根据创建者查找通知
func (r *GormNotificationRepository) FindByCreatedBy(ctx context.Context, createdBy string) ([]*domain.Notification, error) {
	var notifications []*domain.Notification
	err := r.db.WithContext(ctx).
		Where("created_by = ?", createdBy).
		Order("created_at DESC").
		Find(&notifications).Error
	
	return notifications, err
}

// FindByPriority 根据优先级查找通知
func (r *GormNotificationRepository) FindByPriority(ctx context.Context, priority domain.NotificationPriority) ([]*domain.Notification, error) {
	var notifications []*domain.Notification
	err := r.db.WithContext(ctx).
		Where("priority = ?", priority).
		Order("created_at DESC").
		Find(&notifications).Error
	
	return notifications, err
}

// FindWithPagination 分页查找通知
func (r *GormNotificationRepository) FindWithPagination(ctx context.Context, offset, limit int) ([]*domain.Notification, int64, error) {
	var notifications []*domain.Notification
	var total int64
	
	// 获取总数
	err := r.db.WithContext(ctx).Model(&domain.Notification{}).Count(&total).Error
	if err != nil {
		return nil, 0, err
	}
	
	// 获取分页数据
	err = r.db.WithContext(ctx).
		Offset(offset).
		Limit(limit).
		Order("created_at DESC").
		Find(&notifications).Error
	
	return notifications, total, err
}

// FindByStatusWithPagination 根据状态分页查找通知
func (r *GormNotificationRepository) FindByStatusWithPagination(ctx context.Context, status domain.NotificationStatus, offset, limit int) ([]*domain.Notification, int64, error) {
	var notifications []*domain.Notification
	var total int64
	
	query := r.db.WithContext(ctx).Model(&domain.Notification{}).Where("status = ?", status)
	
	// 获取总数
	err := query.Count(&total).Error
	if err != nil {
		return nil, 0, err
	}
	
	// 获取分页数据
	err = query.
		Offset(offset).
		Limit(limit).
		Order("created_at DESC").
		Find(&notifications).Error
	
	return notifications, total, err
}

// FindByCreatedByWithPagination 根据创建者分页查找通知
func (r *GormNotificationRepository) FindByCreatedByWithPagination(ctx context.Context, createdBy string, offset, limit int) ([]*domain.Notification, int64, error) {
	var notifications []*domain.Notification
	var total int64
	
	query := r.db.WithContext(ctx).Model(&domain.Notification{}).Where("created_by = ?", createdBy)
	
	// 获取总数
	err := query.Count(&total).Error
	if err != nil {
		return nil, 0, err
	}
	
	// 获取分页数据
	err = query.
		Offset(offset).
		Limit(limit).
		Order("created_at DESC").
		Find(&notifications).Error
	
	return notifications, total, err
}

// FindScheduledNotifications 查找定时通知
func (r *GormNotificationRepository) FindScheduledNotifications(ctx context.Context, beforeTime int64) ([]*domain.Notification, error) {
	var notifications []*domain.Notification
	err := r.db.WithContext(ctx).
		Where("status = ? AND scheduled_at IS NOT NULL AND scheduled_at <= FROM_UNIXTIME(?)", 
			domain.NotificationStatusPending, beforeTime).
		Order("scheduled_at ASC").
		Find(&notifications).Error
	
	return notifications, err
}

// FindPendingNotifications 查找待发送通知
func (r *GormNotificationRepository) FindPendingNotifications(ctx context.Context, limit int) ([]*domain.Notification, error) {
	var notifications []*domain.Notification
	err := r.db.WithContext(ctx).
		Where("status = ? AND (scheduled_at IS NULL OR scheduled_at <= NOW())", 
			domain.NotificationStatusPending).
		Limit(limit).
		Order("created_at ASC").
		Find(&notifications).Error
	
	return notifications, err
}

// FindFailedNotifications 查找失败的通知
func (r *GormNotificationRepository) FindFailedNotifications(ctx context.Context, limit int) ([]*domain.Notification, error) {
	var notifications []*domain.Notification
	err := r.db.WithContext(ctx).
		Where("status = ?", domain.NotificationStatusFailed).
		Limit(limit).
		Order("failed_at DESC").
		Find(&notifications).Error
	
	return notifications, err
}

// FindRetryableNotifications 查找可重试的通知
func (r *GormNotificationRepository) FindRetryableNotifications(ctx context.Context, limit int) ([]*domain.Notification, error) {
	var notifications []*domain.Notification
	err := r.db.WithContext(ctx).
		Where("status = ? AND retry_count < max_retries", domain.NotificationStatusFailed).
		Limit(limit).
		Order("failed_at ASC").
		Find(&notifications).Error
	
	return notifications, err
}

// SearchByContent 根据内容搜索通知
func (r *GormNotificationRepository) SearchByContent(ctx context.Context, query string, limit int) ([]*domain.Notification, error) {
	var notifications []*domain.Notification
	err := r.db.WithContext(ctx).
		Where("content ILIKE ?", "%"+query+"%").
		Limit(limit).
		Order("created_at DESC").
		Find(&notifications).Error
	
	return notifications, err
}

// SearchByTitle 根据标题搜索通知
func (r *GormNotificationRepository) SearchByTitle(ctx context.Context, query string, limit int) ([]*domain.Notification, error) {
	var notifications []*domain.Notification
	err := r.db.WithContext(ctx).
		Where("title ILIKE ?", "%"+query+"%").
		Limit(limit).
		Order("created_at DESC").
		Find(&notifications).Error
	
	return notifications, err
}

// SaveBatch 批量保存通知
func (r *GormNotificationRepository) SaveBatch(ctx context.Context, notifications []*domain.Notification) error {
	return r.db.WithContext(ctx).CreateInBatches(notifications, 100).Error
}

// UpdateBatch 批量更新通知
func (r *GormNotificationRepository) UpdateBatch(ctx context.Context, notifications []*domain.Notification) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		for _, notification := range notifications {
			if err := tx.Save(notification).Error; err != nil {
				return err
			}
		}
		return nil
	})
}

// UpdateStatusBatch 批量更新状态
func (r *GormNotificationRepository) UpdateStatusBatch(ctx context.Context, ids []string, status domain.NotificationStatus) error {
	return r.db.WithContext(ctx).
		Model(&domain.Notification{}).
		Where("id IN ?", ids).
		Update("status", status).Error
}

// CountByStatus 根据状态统计数量
func (r *GormNotificationRepository) CountByStatus(ctx context.Context, status domain.NotificationStatus) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).
		Model(&domain.Notification{}).
		Where("status = ?", status).
		Count(&count).Error
	
	return count, err
}

// CountByChannel 根据渠道统计数量
func (r *GormNotificationRepository) CountByChannel(ctx context.Context, channel domain.NotificationChannel) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).
		Model(&domain.Notification{}).
		Where("channel = ?", channel).
		Count(&count).Error
	
	return count, err
}

// CountByCreatedBy 根据创建者统计数量
func (r *GormNotificationRepository) CountByCreatedBy(ctx context.Context, createdBy string) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).
		Model(&domain.Notification{}).
		Where("created_by = ?", createdBy).
		Count(&count).Error
	
	return count, err
}

// GetStatsByDateRange 获取日期范围统计
func (r *GormNotificationRepository) GetStatsByDateRange(ctx context.Context, startDate, endDate string) (*repository.NotificationStats, error) {
	stats := &repository.NotificationStats{
		StatusCounts:   make(map[domain.NotificationStatus]int64),
		TypeCounts:     make(map[domain.NotificationType]int64),
		ChannelCounts:  make(map[domain.NotificationChannel]int64),
		PriorityCounts: make(map[domain.NotificationPriority]int64),
	}
	
	query := r.db.WithContext(ctx).Model(&domain.Notification{})
	
	// 添加日期范围过滤
	if startDate != "" && endDate != "" {
		query = query.Where("created_at BETWEEN ? AND ?", startDate, endDate)
	}
	
	// 获取总数
	err := query.Count(&stats.TotalCount).Error
	if err != nil {
		return nil, err
	}
	
	// 按状态统计
	var statusStats []struct {
		Status domain.NotificationStatus
		Count  int64
	}
	err = query.Select("status, count(*) as count").Group("status").Scan(&statusStats).Error
	if err == nil {
		for _, stat := range statusStats {
			stats.StatusCounts[stat.Status] = stat.Count
		}
	}
	
	// 计算成功率
	if delivered, exists := stats.StatusCounts[domain.NotificationStatusDelivered]; exists {
		if sent, exists := stats.StatusCounts[domain.NotificationStatusSent]; exists {
			delivered += sent // 已发送也算成功
		}
		stats.SuccessRate = float64(delivered) / float64(stats.TotalCount)
	}
	
	return stats, nil
}

// GetChannelStats 获取渠道统计
func (r *GormNotificationRepository) GetChannelStats(ctx context.Context) ([]repository.ChannelStats, error) {
	var stats []repository.ChannelStats
	
	rows, err := r.db.WithContext(ctx).Raw(`
		SELECT 
			channel,
			COUNT(*) as total_count,
			SUM(CASE WHEN status IN (?, ?) THEN 1 ELSE 0 END) as success_count,
			SUM(CASE WHEN status = ? THEN 1 ELSE 0 END) as failed_count,
			MAX(sent_at) as last_sent_at
		FROM notifications 
		GROUP BY channel
	`, domain.NotificationStatusSent, domain.NotificationStatusDelivered, domain.NotificationStatusFailed).Rows()
	
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	
	for rows.Next() {
		var stat repository.ChannelStats
		var lastSentAt *string
		
		err := rows.Scan(&stat.Channel, &stat.TotalCount, &stat.SuccessCount, &stat.FailedCount, &lastSentAt)
		if err != nil {
			continue
		}
		
		if stat.TotalCount > 0 {
			stat.SuccessRate = float64(stat.SuccessCount) / float64(stat.TotalCount)
		}
		
		if lastSentAt != nil {
			stat.LastSentAt = *lastSentAt
		}
		
		stats = append(stats, stat)
	}
	
	return stats, nil
}

// DeleteOldNotifications 删除旧通知
func (r *GormNotificationRepository) DeleteOldNotifications(ctx context.Context, beforeTime int64) (int64, error) {
	result := r.db.WithContext(ctx).
		Where("created_at < FROM_UNIXTIME(?) AND status IN (?)", 
			beforeTime, 
			[]domain.NotificationStatus{
				domain.NotificationStatusSent, 
				domain.NotificationStatusDelivered,
				domain.NotificationStatusFailed,
			}).
		Delete(&domain.Notification{})
	
	return result.RowsAffected, result.Error
}

// DeleteCancelledNotifications 删除已取消的通知
func (r *GormNotificationRepository) DeleteCancelledNotifications(ctx context.Context, beforeTime int64) (int64, error) {
	result := r.db.WithContext(ctx).
		Where("created_at < FROM_UNIXTIME(?) AND status = ?", 
			beforeTime, domain.NotificationStatusCancelled).
		Delete(&domain.Notification{})
	
	return result.RowsAffected, result.Error
}
