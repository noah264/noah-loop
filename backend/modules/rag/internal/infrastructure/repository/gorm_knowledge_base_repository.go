package repository

import (
	"context"

	"github.com/noah-loop/backend/modules/rag/internal/domain"
	"github.com/noah-loop/backend/modules/rag/internal/domain/repository"
	"gorm.io/gorm"
)

// GormKnowledgeBaseRepository GORM知识库仓储实现
type GormKnowledgeBaseRepository struct {
	db *gorm.DB
}

// NewGormKnowledgeBaseRepository 创建GORM知识库仓储
func NewGormKnowledgeBaseRepository(db *gorm.DB) repository.KnowledgeBaseRepository {
	return &GormKnowledgeBaseRepository{
		db: db,
	}
}

// Save 保存知识库
func (r *GormKnowledgeBaseRepository) Save(ctx context.Context, knowledgeBase *domain.KnowledgeBase) error {
	return r.db.WithContext(ctx).Create(knowledgeBase).Error
}

// FindByID 根据ID查找知识库
func (r *GormKnowledgeBaseRepository) FindByID(ctx context.Context, id string) (*domain.KnowledgeBase, error) {
	var kb domain.KnowledgeBase
	err := r.db.WithContext(ctx).
		Preload("Documents").
		Preload("Tags").
		First(&kb, "id = ?", id).Error
	
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}
	
	return &kb, nil
}

// FindByName 根据名称查找知识库
func (r *GormKnowledgeBaseRepository) FindByName(ctx context.Context, name, ownerID string) (*domain.KnowledgeBase, error) {
	var kb domain.KnowledgeBase
	err := r.db.WithContext(ctx).
		Where("name = ? AND owner_id = ?", name, ownerID).
		First(&kb).Error
	
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}
	
	return &kb, nil
}

// Update 更新知识库
func (r *GormKnowledgeBaseRepository) Update(ctx context.Context, knowledgeBase *domain.KnowledgeBase) error {
	return r.db.WithContext(ctx).Save(knowledgeBase).Error
}

// Delete 删除知识库
func (r *GormKnowledgeBaseRepository) Delete(ctx context.Context, id string) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// 先删除相关的文档
		if err := tx.Where("knowledge_base_id = ?", id).Delete(&domain.Document{}).Error; err != nil {
			return err
		}
		
		// 删除知识库
		return tx.Delete(&domain.KnowledgeBase{}, "id = ?", id).Error
	})
}

// FindByOwnerID 根据所有者ID查找知识库
func (r *GormKnowledgeBaseRepository) FindByOwnerID(ctx context.Context, ownerID string) ([]*domain.KnowledgeBase, error) {
	var kbs []*domain.KnowledgeBase
	err := r.db.WithContext(ctx).
		Preload("Tags").
		Where("owner_id = ?", ownerID).
		Order("created_at DESC").
		Find(&kbs).Error
	
	return kbs, err
}

// FindByStatus 根据状态查找知识库
func (r *GormKnowledgeBaseRepository) FindByStatus(ctx context.Context, status domain.KnowledgeBaseStatus) ([]*domain.KnowledgeBase, error) {
	var kbs []*domain.KnowledgeBase
	err := r.db.WithContext(ctx).
		Preload("Tags").
		Where("status = ?", status).
		Order("created_at DESC").
		Find(&kbs).Error
	
	return kbs, err
}

// FindByTags 根据标签查找知识库
func (r *GormKnowledgeBaseRepository) FindByTags(ctx context.Context, tagNames []string) ([]*domain.KnowledgeBase, error) {
	var kbs []*domain.KnowledgeBase
	err := r.db.WithContext(ctx).
		Preload("Tags").
		Joins("JOIN knowledge_base_tags ON knowledge_bases.id = knowledge_base_tags.knowledge_base_id").
		Joins("JOIN tags ON knowledge_base_tags.tag_id = tags.id").
		Where("tags.name IN ?", tagNames).
		Group("knowledge_bases.id").
		Find(&kbs).Error
	
	return kbs, err
}

// FindWithPagination 分页查找知识库
func (r *GormKnowledgeBaseRepository) FindWithPagination(ctx context.Context, offset, limit int) ([]*domain.KnowledgeBase, int64, error) {
	var kbs []*domain.KnowledgeBase
	var total int64
	
	// 获取总数
	err := r.db.WithContext(ctx).Model(&domain.KnowledgeBase{}).Count(&total).Error
	if err != nil {
		return nil, 0, err
	}
	
	// 获取分页数据
	err = r.db.WithContext(ctx).
		Preload("Tags").
		Offset(offset).
		Limit(limit).
		Order("created_at DESC").
		Find(&kbs).Error
	
	return kbs, total, err
}

// FindByOwnerIDWithPagination 根据所有者ID分页查找知识库
func (r *GormKnowledgeBaseRepository) FindByOwnerIDWithPagination(ctx context.Context, ownerID string, offset, limit int) ([]*domain.KnowledgeBase, int64, error) {
	var kbs []*domain.KnowledgeBase
	var total int64
	
	query := r.db.WithContext(ctx).Model(&domain.KnowledgeBase{}).Where("owner_id = ?", ownerID)
	
	// 获取总数
	err := query.Count(&total).Error
	if err != nil {
		return nil, 0, err
	}
	
	// 获取分页数据
	err = query.
		Preload("Tags").
		Offset(offset).
		Limit(limit).
		Order("created_at DESC").
		Find(&kbs).Error
	
	return kbs, total, err
}

// SearchByName 根据名称搜索知识库
func (r *GormKnowledgeBaseRepository) SearchByName(ctx context.Context, query string, ownerID string, limit int) ([]*domain.KnowledgeBase, error) {
	var kbs []*domain.KnowledgeBase
	
	dbQuery := r.db.WithContext(ctx).
		Preload("Tags").
		Where("name ILIKE ?", "%"+query+"%")
	
	if ownerID != "" {
		dbQuery = dbQuery.Where("owner_id = ?", ownerID)
	}
	
	err := dbQuery.
		Limit(limit).
		Order("created_at DESC").
		Find(&kbs).Error
	
	return kbs, err
}

// SearchByDescription 根据描述搜索知识库
func (r *GormKnowledgeBaseRepository) SearchByDescription(ctx context.Context, query string, limit int) ([]*domain.KnowledgeBase, error) {
	var kbs []*domain.KnowledgeBase
	
	err := r.db.WithContext(ctx).
		Preload("Tags").
		Where("description ILIKE ?", "%"+query+"%").
		Limit(limit).
		Order("created_at DESC").
		Find(&kbs).Error
	
	return kbs, err
}

// CountByOwnerID 根据所有者ID统计知识库数量
func (r *GormKnowledgeBaseRepository) CountByOwnerID(ctx context.Context, ownerID string) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).
		Model(&domain.KnowledgeBase{}).
		Where("owner_id = ?", ownerID).
		Count(&count).Error
	
	return count, err
}

// CountByStatus 根据状态统计知识库数量
func (r *GormKnowledgeBaseRepository) CountByStatus(ctx context.Context, status domain.KnowledgeBaseStatus) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).
		Model(&domain.KnowledgeBase{}).
		Where("status = ?", status).
		Count(&count).Error
	
	return count, err
}

// UpdateStatistics 更新知识库统计信息
func (r *GormKnowledgeBaseRepository) UpdateStatistics(ctx context.Context, knowledgeBaseID string, stats domain.KnowledgeBaseStats) error {
	return r.db.WithContext(ctx).
		Model(&domain.KnowledgeBase{}).
		Where("id = ?", knowledgeBaseID).
		Updates(map[string]interface{}{
			"statistics": stats,
			"updated_at": gorm.Expr("NOW()"),
		}).Error
}

// RecordQuery 记录查询统计
func (r *GormKnowledgeBaseRepository) RecordQuery(ctx context.Context, knowledgeBaseID string, score float32) error {
	// 更新查询统计
	return r.db.WithContext(ctx).Exec(`
		UPDATE knowledge_bases 
		SET 
			statistics = jsonb_set(
				jsonb_set(
					jsonb_set(statistics, '{query_count}', 
						(COALESCE((statistics->>'query_count')::bigint, 0) + 1)::text::jsonb),
					'{last_query_at}', 
					to_jsonb(NOW()::text)),
				'{average_score}',
				CASE 
					WHEN COALESCE((statistics->>'query_count')::bigint, 0) = 0 
					THEN to_jsonb(?::text)
					ELSE to_jsonb(((COALESCE((statistics->>'average_score')::float, 0) * COALESCE((statistics->>'query_count')::bigint, 0) + ?) / (COALESCE((statistics->>'query_count')::bigint, 0) + 1))::text)
				END
			),
			updated_at = NOW()
		WHERE id = ?
	`, score, score, knowledgeBaseID).Error
}

// GetQueryHistory 获取查询历史
func (r *GormKnowledgeBaseRepository) GetQueryHistory(ctx context.Context, knowledgeBaseID string, limit int) ([]repository.QueryRecord, error) {
	// 这里应该有一个单独的查询历史表
	// 为了简化，这里返回空记录
	return []repository.QueryRecord{}, nil
}

// CheckAccess 检查访问权限
func (r *GormKnowledgeBaseRepository) CheckAccess(ctx context.Context, knowledgeBaseID, userID string) (bool, error) {
	var count int64
	
	// 检查是否是所有者
	err := r.db.WithContext(ctx).
		Model(&domain.KnowledgeBase{}).
		Where("id = ? AND owner_id = ?", knowledgeBaseID, userID).
		Count(&count).Error
	
	if err != nil {
		return false, err
	}
	
	if count > 0 {
		return true, nil
	}
	
	// TODO: 检查共享权限表
	// 这里需要实现权限表的查询逻辑
	
	return false, nil
}

// GrantAccess 授予访问权限
func (r *GormKnowledgeBaseRepository) GrantAccess(ctx context.Context, knowledgeBaseID, userID string, permission repository.Permission) error {
	// TODO: 实现权限授予逻辑
	// 这里需要创建权限表并实现相关逻辑
	return nil
}

// RevokeAccess 撤销访问权限
func (r *GormKnowledgeBaseRepository) RevokeAccess(ctx context.Context, knowledgeBaseID, userID string) error {
	// TODO: 实现权限撤销逻辑
	return nil
}

// ListAccessUsers 列出有访问权限的用户
func (r *GormKnowledgeBaseRepository) ListAccessUsers(ctx context.Context, knowledgeBaseID string) ([]repository.UserPermission, error) {
	// TODO: 实现权限用户列表逻辑
	return []repository.UserPermission{}, nil
}
