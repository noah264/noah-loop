package repository

import (
	"context"
	"errors"
	
	"github.com/google/uuid"
	"github.com/noah-loop/backend/modules/llm/internal/domain"
	"github.com/noah-loop/backend/shared/pkg/infrastructure"
	"gorm.io/gorm"
)

// GormModelRepository GORM模型仓储实现
type GormModelRepository struct {
	db *infrastructure.Database
}

// NewGormModelRepository 创建GORM模型仓储
func NewGormModelRepository(db *infrastructure.Database) domain.ModelRepository {
	return &GormModelRepository{db: db}
}

// Save 保存模型
func (r *GormModelRepository) Save(ctx context.Context, entity *domain.Model) error {
	return r.db.DB.WithContext(ctx).Save(entity).Error
}

// FindByID 根据ID查找模型
func (r *GormModelRepository) FindByID(ctx context.Context, id uuid.UUID) (*domain.Model, error) {
	var model domain.Model
	err := r.db.DB.WithContext(ctx).First(&model, "id = ?", id).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("model not found")
		}
		return nil, err
	}
	return &model, nil
}

// FindAll 查找所有模型
func (r *GormModelRepository) FindAll(ctx context.Context, offset, limit int) ([]*domain.Model, error) {
	var models []*domain.Model
	err := r.db.DB.WithContext(ctx).
		Offset(offset).
		Limit(limit).
		Find(&models).Error
	return models, err
}

// Delete 删除模型
func (r *GormModelRepository) Delete(ctx context.Context, id uuid.UUID) error {
	return r.db.DB.WithContext(ctx).Delete(&domain.Model{}, "id = ?", id).Error
}

// Count 计算模型数量
func (r *GormModelRepository) Count(ctx context.Context) (int64, error) {
	var count int64
	err := r.db.DB.WithContext(ctx).Model(&domain.Model{}).Count(&count).Error
	return count, err
}

// FindByProvider 根据提供商查找模型
func (r *GormModelRepository) FindByProvider(ctx context.Context, provider domain.ModelProvider) ([]*domain.Model, error) {
	var models []*domain.Model
	err := r.db.DB.WithContext(ctx).
		Where("provider = ?", provider).
		Find(&models).Error
	return models, err
}

// FindByType 根据类型查找模型
func (r *GormModelRepository) FindByType(ctx context.Context, modelType domain.ModelType) ([]*domain.Model, error) {
	var models []*domain.Model
	err := r.db.DB.WithContext(ctx).
		Where("type = ?", modelType).
		Find(&models).Error
	return models, err
}

// FindActiveModels 查找激活的模型
func (r *GormModelRepository) FindActiveModels(ctx context.Context) ([]*domain.Model, error) {
	var models []*domain.Model
	err := r.db.DB.WithContext(ctx).
		Where("is_active = ?", true).
		Find(&models).Error
	return models, err
}

// FindByNameAndProvider 根据名称和提供商查找模型
func (r *GormModelRepository) FindByNameAndProvider(ctx context.Context, name string, provider domain.ModelProvider) (*domain.Model, error) {
	var model domain.Model
	err := r.db.DB.WithContext(ctx).
		Where("name = ? AND provider = ?", name, provider).
		First(&model).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("model not found")
		}
		return nil, err
	}
	return &model, nil
}
