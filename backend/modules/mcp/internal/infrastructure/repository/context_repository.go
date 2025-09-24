package repository

import (
	"context"
	"errors"
	"time"
	
	"github.com/google/uuid"
	"github.com/noah-loop/backend/modules/mcp/internal/domain"
	"github.com/noah-loop/backend/shared/pkg/infrastructure"
	"gorm.io/gorm"
)

// GormContextRepository GORM上下文仓储实现
type GormContextRepository struct {
	db *infrastructure.Database
}

// NewGormContextRepository 创建GORM上下文仓储
func NewGormContextRepository(db *infrastructure.Database) domain.ContextRepository {
	return &GormContextRepository{db: db}
}

// Save 保存上下文
func (r *GormContextRepository) Save(ctx context.Context, entity *domain.Context) error {
	return r.db.DB.WithContext(ctx).Save(entity).Error
}

// FindByID 根据ID查找上下文
func (r *GormContextRepository) FindByID(ctx context.Context, id uuid.UUID) (*domain.Context, error) {
	var context domain.Context
	err := r.db.DB.WithContext(ctx).
		Preload("Session").
		First(&context, "id = ?", id).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("context not found")
		}
		return nil, err
	}
	return &context, nil
}

// FindAll 查找所有上下文
func (r *GormContextRepository) FindAll(ctx context.Context, offset, limit int) ([]*domain.Context, error) {
	var contexts []*domain.Context
	err := r.db.DB.WithContext(ctx).
		Preload("Session").
		Offset(offset).
		Limit(limit).
		Order("created_at DESC").
		Find(&contexts).Error
	return contexts, err
}

// Delete 删除上下文
func (r *GormContextRepository) Delete(ctx context.Context, id uuid.UUID) error {
	return r.db.DB.WithContext(ctx).Delete(&domain.Context{}, "id = ?", id).Error
}

// Count 计算上下文数量
func (r *GormContextRepository) Count(ctx context.Context) (int64, error) {
	var count int64
	err := r.db.DB.WithContext(ctx).Model(&domain.Context{}).Count(&count).Error
	return count, err
}

// FindBySessionID 根据会话ID查找上下文
func (r *GormContextRepository) FindBySessionID(ctx context.Context, sessionID uuid.UUID) ([]*domain.Context, error) {
	var contexts []*domain.Context
	err := r.db.DB.WithContext(ctx).
		Where("session_id = ?", sessionID).
		Order("created_at ASC").
		Find(&contexts).Error
	return contexts, err
}

// FindByType 根据类型查找上下文
func (r *GormContextRepository) FindByType(ctx context.Context, contextType domain.ContextType) ([]*domain.Context, error) {
	var contexts []*domain.Context
	err := r.db.DB.WithContext(ctx).
		Where("type = ?", contextType).
		Order("created_at DESC").
		Find(&contexts).Error
	return contexts, err
}

// FindByPriority 根据优先级查找上下文
func (r *GormContextRepository) FindByPriority(ctx context.Context, minPriority int) ([]*domain.Context, error) {
	var contexts []*domain.Context
	err := r.db.DB.WithContext(ctx).
		Where("priority >= ?", minPriority).
		Order("priority DESC, created_at DESC").
		Find(&contexts).Error
	return contexts, err
}

// FindExpiredContexts 查找过期上下文
func (r *GormContextRepository) FindExpiredContexts(ctx context.Context, before time.Time) ([]*domain.Context, error) {
	var contexts []*domain.Context
	err := r.db.DB.WithContext(ctx).
		Where("last_accessed < ? AND access_count = 0", before).
		Find(&contexts).Error
	return contexts, err
}

// GetSessionContextSize 获取会话上下文总大小
func (r *GormContextRepository) GetSessionContextSize(ctx context.Context, sessionID uuid.UUID) (int, error) {
	var totalSize struct {
		Total int
	}
	
	err := r.db.DB.WithContext(ctx).
		Model(&domain.Context{}).
		Select("COALESCE(SUM(token_count), 0) as total").
		Where("session_id = ?", sessionID).
		Scan(&totalSize).Error
	
	return totalSize.Total, err
}
