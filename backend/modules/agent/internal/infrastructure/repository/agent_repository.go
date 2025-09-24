package repository

import (
	"context"
	"errors"
	
	"github.com/google/uuid"
	"github.com/noah-loop/backend/modules/agent/internal/domain"
	"github.com/noah-loop/backend/shared/pkg/infrastructure"
	"gorm.io/gorm"
)

// GormAgentRepository GORM智能体仓储实现
type GormAgentRepository struct {
	db *infrastructure.Database
}

// NewGormAgentRepository 创建GORM智能体仓储
func NewGormAgentRepository(db *infrastructure.Database) domain.AgentRepository {
	return &GormAgentRepository{db: db}
}

// Save 保存智能体
func (r *GormAgentRepository) Save(ctx context.Context, entity *domain.Agent) error {
	return r.db.DB.WithContext(ctx).Save(entity).Error
}

// FindByID 根据ID查找智能体
func (r *GormAgentRepository) FindByID(ctx context.Context, id uuid.UUID) (*domain.Agent, error) {
	var agent domain.Agent
	err := r.db.DB.WithContext(ctx).
		Preload("Memory").
		Preload("Tools").
		First(&agent, "id = ?", id).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("agent not found")
		}
		return nil, err
	}
	return &agent, nil
}

// FindAll 查找所有智能体
func (r *GormAgentRepository) FindAll(ctx context.Context, offset, limit int) ([]*domain.Agent, error) {
	var agents []*domain.Agent
	err := r.db.DB.WithContext(ctx).
		Preload("Memory").
		Preload("Tools").
		Offset(offset).
		Limit(limit).
		Order("created_at DESC").
		Find(&agents).Error
	return agents, err
}

// Delete 删除智能体
func (r *GormAgentRepository) Delete(ctx context.Context, id uuid.UUID) error {
	return r.db.DB.WithContext(ctx).Delete(&domain.Agent{}, "id = ?", id).Error
}

// Count 计算智能体数量
func (r *GormAgentRepository) Count(ctx context.Context) (int64, error) {
	var count int64
	err := r.db.DB.WithContext(ctx).Model(&domain.Agent{}).Count(&count).Error
	return count, err
}

// FindByOwnerID 根据所有者ID查找智能体
func (r *GormAgentRepository) FindByOwnerID(ctx context.Context, ownerID uuid.UUID) ([]*domain.Agent, error) {
	var agents []*domain.Agent
	err := r.db.DB.WithContext(ctx).
		Preload("Memory").
		Preload("Tools").
		Where("owner_id = ?", ownerID).
		Order("created_at DESC").
		Find(&agents).Error
	return agents, err
}

// FindByType 根据类型查找智能体
func (r *GormAgentRepository) FindByType(ctx context.Context, agentType domain.AgentType) ([]*domain.Agent, error) {
	var agents []*domain.Agent
	err := r.db.DB.WithContext(ctx).
		Preload("Memory").
		Preload("Tools").
		Where("type = ?", agentType).
		Find(&agents).Error
	return agents, err
}

// FindActiveAgents 查找活跃的智能体
func (r *GormAgentRepository) FindActiveAgents(ctx context.Context) ([]*domain.Agent, error) {
	var agents []*domain.Agent
	err := r.db.DB.WithContext(ctx).
		Preload("Memory").
		Preload("Tools").
		Where("is_active = ?", true).
		Find(&agents).Error
	return agents, err
}

// FindByStatus 根据状态查找智能体
func (r *GormAgentRepository) FindByStatus(ctx context.Context, status domain.AgentStatus) ([]*domain.Agent, error) {
	var agents []*domain.Agent
	err := r.db.DB.WithContext(ctx).
		Preload("Memory").
		Preload("Tools").
		Where("status = ?", status).
		Find(&agents).Error
	return agents, err
}
