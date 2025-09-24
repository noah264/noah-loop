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

// GormSessionRepository GORM会话仓储实现
type GormSessionRepository struct {
	db *infrastructure.Database
}

// NewGormSessionRepository 创建GORM会话仓储
func NewGormSessionRepository(db *infrastructure.Database) domain.SessionRepository {
	return &GormSessionRepository{db: db}
}

// Save 保存会话
func (r *GormSessionRepository) Save(ctx context.Context, entity *domain.Session) error {
	return r.db.DB.WithContext(ctx).Save(entity).Error
}

// FindByID 根据ID查找会话
func (r *GormSessionRepository) FindByID(ctx context.Context, id uuid.UUID) (*domain.Session, error) {
	var session domain.Session
	err := r.db.DB.WithContext(ctx).
		Preload("Contexts").
		First(&session, "id = ?", id).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("session not found")
		}
		return nil, err
	}
	return &session, nil
}

// FindAll 查找所有会话
func (r *GormSessionRepository) FindAll(ctx context.Context, offset, limit int) ([]*domain.Session, error) {
	var sessions []*domain.Session
	err := r.db.DB.WithContext(ctx).
		Preload("Contexts").
		Offset(offset).
		Limit(limit).
		Order("last_activity DESC").
		Find(&sessions).Error
	return sessions, err
}

// Delete 删除会话
func (r *GormSessionRepository) Delete(ctx context.Context, id uuid.UUID) error {
	return r.db.DB.WithContext(ctx).Delete(&domain.Session{}, "id = ?", id).Error
}

// Count 计算会话数量
func (r *GormSessionRepository) Count(ctx context.Context) (int64, error) {
	var count int64
	err := r.db.DB.WithContext(ctx).Model(&domain.Session{}).Count(&count).Error
	return count, err
}

// FindByUserID 根据用户ID查找会话
func (r *GormSessionRepository) FindByUserID(ctx context.Context, userID uuid.UUID) ([]*domain.Session, error) {
	var sessions []*domain.Session
	err := r.db.DB.WithContext(ctx).
		Preload("Contexts").
		Where("user_id = ?", userID).
		Order("last_activity DESC").
		Find(&sessions).Error
	return sessions, err
}

// FindByAgentID 根据智能体ID查找会话
func (r *GormSessionRepository) FindByAgentID(ctx context.Context, agentID uuid.UUID) ([]*domain.Session, error) {
	var sessions []*domain.Session
	err := r.db.DB.WithContext(ctx).
		Preload("Contexts").
		Where("agent_id = ?", agentID).
		Order("last_activity DESC").
		Find(&sessions).Error
	return sessions, err
}

// FindByStatus 根据状态查找会话
func (r *GormSessionRepository) FindByStatus(ctx context.Context, status domain.SessionStatus) ([]*domain.Session, error) {
	var sessions []*domain.Session
	err := r.db.DB.WithContext(ctx).
		Preload("Contexts").
		Where("status = ?", status).
		Order("last_activity DESC").
		Find(&sessions).Error
	return sessions, err
}

// FindExpiredSessions 查找过期会话
func (r *GormSessionRepository) FindExpiredSessions(ctx context.Context) ([]*domain.Session, error) {
	var sessions []*domain.Session
	err := r.db.DB.WithContext(ctx).
		Where("expires_at IS NOT NULL AND expires_at < ? AND status != ?", 
			time.Now(), domain.SessionStatusExpired).
		Find(&sessions).Error
	return sessions, err
}

// FindIdleSessions 查找空闲会话
func (r *GormSessionRepository) FindIdleSessions(ctx context.Context, idleThreshold time.Duration) ([]*domain.Session, error) {
	var sessions []*domain.Session
	cutoffTime := time.Now().Add(-idleThreshold)
	err := r.db.DB.WithContext(ctx).
		Where("status = ? AND last_activity < ?", domain.SessionStatusActive, cutoffTime).
		Find(&sessions).Error
	return sessions, err
}
