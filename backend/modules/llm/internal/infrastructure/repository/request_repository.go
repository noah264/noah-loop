package repository

import (
	"context"
	"errors"
	"time"
	
	"github.com/google/uuid"
	"github.com/noah-loop/backend/modules/llm/internal/domain"
	"github.com/noah-loop/backend/shared/pkg/infrastructure"
	"gorm.io/gorm"
)

// GormRequestRepository GORM请求仓储实现
type GormRequestRepository struct {
	db *infrastructure.Database
}

// NewGormRequestRepository 创建GORM请求仓储
func NewGormRequestRepository(db *infrastructure.Database) domain.RequestRepository {
	return &GormRequestRepository{db: db}
}

// Save 保存请求
func (r *GormRequestRepository) Save(ctx context.Context, entity *domain.Request) error {
	return r.db.DB.WithContext(ctx).Save(entity).Error
}

// FindByID 根据ID查找请求
func (r *GormRequestRepository) FindByID(ctx context.Context, id uuid.UUID) (*domain.Request, error) {
	var request domain.Request
	err := r.db.DB.WithContext(ctx).
		Preload("Model").
		First(&request, "id = ?", id).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("request not found")
		}
		return nil, err
	}
	return &request, nil
}

// FindAll 查找所有请求
func (r *GormRequestRepository) FindAll(ctx context.Context, offset, limit int) ([]*domain.Request, error) {
	var requests []*domain.Request
	err := r.db.DB.WithContext(ctx).
		Preload("Model").
		Offset(offset).
		Limit(limit).
		Order("created_at DESC").
		Find(&requests).Error
	return requests, err
}

// Delete 删除请求
func (r *GormRequestRepository) Delete(ctx context.Context, id uuid.UUID) error {
	return r.db.DB.WithContext(ctx).Delete(&domain.Request{}, "id = ?", id).Error
}

// Count 计算请求数量
func (r *GormRequestRepository) Count(ctx context.Context) (int64, error) {
	var count int64
	err := r.db.DB.WithContext(ctx).Model(&domain.Request{}).Count(&count).Error
	return count, err
}

// FindByModelID 根据模型ID查找请求
func (r *GormRequestRepository) FindByModelID(ctx context.Context, modelID uuid.UUID, offset, limit int) ([]*domain.Request, error) {
	var requests []*domain.Request
	err := r.db.DB.WithContext(ctx).
		Preload("Model").
		Where("model_id = ?", modelID).
		Offset(offset).
		Limit(limit).
		Order("created_at DESC").
		Find(&requests).Error
	return requests, err
}

// FindByUserID 根据用户ID查找请求
func (r *GormRequestRepository) FindByUserID(ctx context.Context, userID uuid.UUID, offset, limit int) ([]*domain.Request, error) {
	var requests []*domain.Request
	err := r.db.DB.WithContext(ctx).
		Preload("Model").
		Where("user_id = ?", userID).
		Offset(offset).
		Limit(limit).
		Order("created_at DESC").
		Find(&requests).Error
	return requests, err
}

// FindBySessionID 根据会话ID查找请求
func (r *GormRequestRepository) FindBySessionID(ctx context.Context, sessionID uuid.UUID) ([]*domain.Request, error) {
	var requests []*domain.Request
	err := r.db.DB.WithContext(ctx).
		Preload("Model").
		Where("session_id = ?", sessionID).
		Order("created_at ASC").
		Find(&requests).Error
	return requests, err
}

// FindByStatus 根据状态查找请求
func (r *GormRequestRepository) FindByStatus(ctx context.Context, status domain.RequestStatus, offset, limit int) ([]*domain.Request, error) {
	var requests []*domain.Request
	err := r.db.DB.WithContext(ctx).
		Preload("Model").
		Where("status = ?", status).
		Offset(offset).
		Limit(limit).
		Order("created_at DESC").
		Find(&requests).Error
	return requests, err
}

// GetUsageStats 获取使用统计
func (r *GormRequestRepository) GetUsageStats(ctx context.Context, userID uuid.UUID, start, end time.Time) (*domain.UsageStats, error) {
	var stats struct {
		TotalRequests int64   `json:"total_requests"`
		TotalTokens   int64   `json:"total_tokens"`
		TotalCost     float64 `json:"total_cost"`
		AvgDuration   float64 `json:"avg_duration"`
	}
	
	query := r.db.DB.WithContext(ctx).
		Model(&domain.Request{}).
		Select(`
			COUNT(*) as total_requests,
			COALESCE(SUM(tokens_used), 0) as total_tokens,
			COALESCE(SUM(cost), 0) as total_cost,
			COALESCE(AVG(EXTRACT(EPOCH FROM duration::interval)), 0) as avg_duration
		`).
		Where("user_id = ? AND created_at BETWEEN ? AND ? AND status = ?", 
			userID, start, end, domain.RequestStatusCompleted)
	
	if err := query.Scan(&stats).Error; err != nil {
		return nil, err
	}
	
	return &domain.UsageStats{
		TotalRequests: int(stats.TotalRequests),
		TotalTokens:   int(stats.TotalTokens),
		TotalCost:     stats.TotalCost,
		AvgDuration:   stats.AvgDuration,
	}, nil
}
