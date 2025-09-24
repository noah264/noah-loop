package repository

import (
	"context"
	"errors"
	
	"github.com/google/uuid"
	"github.com/noah-loop/backend/modules/agent/internal/domain"
	"github.com/noah-loop/backend/shared/pkg/infrastructure"
	"gorm.io/gorm"
)

// GormToolRepository GORM工具仓储实现
type GormToolRepository struct {
	db *infrastructure.Database
}

// NewGormToolRepository 创建GORM工具仓储
func NewGormToolRepository(db *infrastructure.Database) domain.ToolRepository {
	return &GormToolRepository{db: db}
}

// Save 保存工具
func (r *GormToolRepository) Save(ctx context.Context, entity *domain.Tool) error {
	return r.db.DB.WithContext(ctx).Save(entity).Error
}

// FindByID 根据ID查找工具
func (r *GormToolRepository) FindByID(ctx context.Context, id uuid.UUID) (*domain.Tool, error) {
	var tool domain.Tool
	err := r.db.DB.WithContext(ctx).
		Preload("Agents").
		First(&tool, "id = ?", id).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("tool not found")
		}
		return nil, err
	}
	return &tool, nil
}

// FindAll 查找所有工具
func (r *GormToolRepository) FindAll(ctx context.Context, offset, limit int) ([]*domain.Tool, error) {
	var tools []*domain.Tool
	err := r.db.DB.WithContext(ctx).
		Preload("Agents").
		Offset(offset).
		Limit(limit).
		Order("created_at DESC").
		Find(&tools).Error
	return tools, err
}

// Delete 删除工具
func (r *GormToolRepository) Delete(ctx context.Context, id uuid.UUID) error {
	return r.db.DB.WithContext(ctx).Delete(&domain.Tool{}, "id = ?", id).Error
}

// Count 计算工具数量
func (r *GormToolRepository) Count(ctx context.Context) (int64, error) {
	var count int64
	err := r.db.DB.WithContext(ctx).Model(&domain.Tool{}).Count(&count).Error
	return count, err
}

// FindByOwnerID 根据所有者ID查找工具
func (r *GormToolRepository) FindByOwnerID(ctx context.Context, ownerID uuid.UUID) ([]*domain.Tool, error) {
	var tools []*domain.Tool
	err := r.db.DB.WithContext(ctx).
		Where("owner_id = ?", ownerID).
		Find(&tools).Error
	return tools, err
}

// FindByType 根据类型查找工具
func (r *GormToolRepository) FindByType(ctx context.Context, toolType domain.ToolType) ([]*domain.Tool, error) {
	var tools []*domain.Tool
	err := r.db.DB.WithContext(ctx).
		Where("type = ?", toolType).
		Find(&tools).Error
	return tools, err
}

// FindPublicTools 查找公共工具
func (r *GormToolRepository) FindPublicTools(ctx context.Context) ([]*domain.Tool, error) {
	var tools []*domain.Tool
	err := r.db.DB.WithContext(ctx).
		Where("is_public = ? AND is_enabled = ?", true, true).
		Find(&tools).Error
	return tools, err
}

// FindEnabledTools 查找启用的工具
func (r *GormToolRepository) FindEnabledTools(ctx context.Context) ([]*domain.Tool, error) {
	var tools []*domain.Tool
	err := r.db.DB.WithContext(ctx).
		Where("is_enabled = ?", true).
		Find(&tools).Error
	return tools, err
}

// FindByAgentID 根据智能体ID查找工具
func (r *GormToolRepository) FindByAgentID(ctx context.Context, agentID uuid.UUID) ([]*domain.Tool, error) {
	var tools []*domain.Tool
	err := r.db.DB.WithContext(ctx).
		Joins("JOIN agent_tools ON tools.id = agent_tools.tool_id").
		Where("agent_tools.agent_id = ? AND tools.is_enabled = ?", agentID, true).
		Find(&tools).Error
	return tools, err
}

// GormToolExecutionRepository GORM工具执行仓储实现
type GormToolExecutionRepository struct {
	db *infrastructure.Database
}

// NewGormToolExecutionRepository 创建GORM工具执行仓储
func NewGormToolExecutionRepository(db *infrastructure.Database) domain.ToolExecutionRepository {
	return &GormToolExecutionRepository{db: db}
}

// Save 保存工具执行记录
func (r *GormToolExecutionRepository) Save(ctx context.Context, entity *domain.ToolExecution) error {
	return r.db.DB.WithContext(ctx).Save(entity).Error
}

// FindByID 根据ID查找工具执行记录
func (r *GormToolExecutionRepository) FindByID(ctx context.Context, id uuid.UUID) (*domain.ToolExecution, error) {
	var execution domain.ToolExecution
	err := r.db.DB.WithContext(ctx).
		Preload("Tool").
		Preload("Agent").
		First(&execution, "id = ?", id).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("tool execution not found")
		}
		return nil, err
	}
	return &execution, nil
}

// FindAll 查找所有工具执行记录
func (r *GormToolExecutionRepository) FindAll(ctx context.Context, offset, limit int) ([]*domain.ToolExecution, error) {
	var executions []*domain.ToolExecution
	err := r.db.DB.WithContext(ctx).
		Preload("Tool").
		Preload("Agent").
		Offset(offset).
		Limit(limit).
		Order("created_at DESC").
		Find(&executions).Error
	return executions, err
}

// Delete 删除工具执行记录
func (r *GormToolExecutionRepository) Delete(ctx context.Context, id uuid.UUID) error {
	return r.db.DB.WithContext(ctx).Delete(&domain.ToolExecution{}, "id = ?", id).Error
}

// Count 计算工具执行记录数量
func (r *GormToolExecutionRepository) Count(ctx context.Context) (int64, error) {
	var count int64
	err := r.db.DB.WithContext(ctx).Model(&domain.ToolExecution{}).Count(&count).Error
	return count, err
}

// FindByToolID 根据工具ID查找执行记录
func (r *GormToolExecutionRepository) FindByToolID(ctx context.Context, toolID uuid.UUID, offset, limit int) ([]*domain.ToolExecution, error) {
	var executions []*domain.ToolExecution
	err := r.db.DB.WithContext(ctx).
		Preload("Tool").
		Preload("Agent").
		Where("tool_id = ?", toolID).
		Offset(offset).
		Limit(limit).
		Order("created_at DESC").
		Find(&executions).Error
	return executions, err
}

// FindByAgentID 根据智能体ID查找执行记录
func (r *GormToolExecutionRepository) FindByAgentID(ctx context.Context, agentID uuid.UUID, offset, limit int) ([]*domain.ToolExecution, error) {
	var executions []*domain.ToolExecution
	err := r.db.DB.WithContext(ctx).
		Preload("Tool").
		Preload("Agent").
		Where("agent_id = ?", agentID).
		Offset(offset).
		Limit(limit).
		Order("created_at DESC").
		Find(&executions).Error
	return executions, err
}

// FindByStatus 根据状态查找执行记录
func (r *GormToolExecutionRepository) FindByStatus(ctx context.Context, status domain.ExecutionStatus) ([]*domain.ToolExecution, error) {
	var executions []*domain.ToolExecution
	err := r.db.DB.WithContext(ctx).
		Preload("Tool").
		Preload("Agent").
		Where("status = ?", status).
		Order("created_at DESC").
		Find(&executions).Error
	return executions, err
}
