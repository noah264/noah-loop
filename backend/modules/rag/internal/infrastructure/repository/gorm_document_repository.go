package repository

import (
	"context"
	"fmt"

	"github.com/noah-loop/backend/modules/rag/internal/domain"
	"github.com/noah-loop/backend/modules/rag/internal/domain/repository"
	"gorm.io/gorm"
)

// GormDocumentRepository GORM文档仓储实现
type GormDocumentRepository struct {
	db *gorm.DB
}

// NewGormDocumentRepository 创建GORM文档仓储
func NewGormDocumentRepository(db *gorm.DB) repository.DocumentRepository {
	return &GormDocumentRepository{
		db: db,
	}
}

// Save 保存文档
func (r *GormDocumentRepository) Save(ctx context.Context, document *domain.Document) error {
	return r.db.WithContext(ctx).Create(document).Error
}

// FindByID 根据ID查找文档
func (r *GormDocumentRepository) FindByID(ctx context.Context, id string) (*domain.Document, error) {
	var document domain.Document
	err := r.db.WithContext(ctx).
		Preload("Tags").
		Preload("Chunks").
		First(&document, "id = ?", id).Error
	
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}
	
	return &document, nil
}

// FindByHash 根据哈希查找文档
func (r *GormDocumentRepository) FindByHash(ctx context.Context, hash string) (*domain.Document, error) {
	var document domain.Document
	err := r.db.WithContext(ctx).
		Preload("Tags").
		First(&document, "hash = ?", hash).Error
	
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}
	
	return &document, nil
}

// Update 更新文档
func (r *GormDocumentRepository) Update(ctx context.Context, document *domain.Document) error {
	return r.db.WithContext(ctx).Save(document).Error
}

// Delete 删除文档
func (r *GormDocumentRepository) Delete(ctx context.Context, id string) error {
	return r.db.WithContext(ctx).Delete(&domain.Document{}, "id = ?", id).Error
}

// FindByKnowledgeBaseID 根据知识库ID查找文档
func (r *GormDocumentRepository) FindByKnowledgeBaseID(ctx context.Context, knowledgeBaseID string) ([]*domain.Document, error) {
	var documents []*domain.Document
	err := r.db.WithContext(ctx).
		Preload("Tags").
		Find(&documents, "knowledge_base_id = ?", knowledgeBaseID).Error
	
	return documents, err
}

// FindByStatus 根据状态查找文档
func (r *GormDocumentRepository) FindByStatus(ctx context.Context, status domain.DocumentStatus) ([]*domain.Document, error) {
	var documents []*domain.Document
	err := r.db.WithContext(ctx).
		Preload("Tags").
		Find(&documents, "status = ?", status).Error
	
	return documents, err
}

// FindByType 根据类型查找文档
func (r *GormDocumentRepository) FindByType(ctx context.Context, docType domain.DocumentType) ([]*domain.Document, error) {
	var documents []*domain.Document
	err := r.db.WithContext(ctx).
		Preload("Tags").
		Find(&documents, "type = ?", docType).Error
	
	return documents, err
}

// FindByTags 根据标签查找文档
func (r *GormDocumentRepository) FindByTags(ctx context.Context, tagNames []string) ([]*domain.Document, error) {
	var documents []*domain.Document
	err := r.db.WithContext(ctx).
		Preload("Tags").
		Joins("JOIN document_tags ON documents.id = document_tags.document_id").
		Joins("JOIN tags ON document_tags.tag_id = tags.id").
		Where("tags.name IN ?", tagNames).
		Group("documents.id").
		Find(&documents).Error
	
	return documents, err
}

// FindWithPagination 分页查找文档
func (r *GormDocumentRepository) FindWithPagination(ctx context.Context, offset, limit int) ([]*domain.Document, int64, error) {
	var documents []*domain.Document
	var total int64
	
	// 获取总数
	err := r.db.WithContext(ctx).Model(&domain.Document{}).Count(&total).Error
	if err != nil {
		return nil, 0, err
	}
	
	// 获取分页数据
	err = r.db.WithContext(ctx).
		Preload("Tags").
		Offset(offset).
		Limit(limit).
		Order("created_at DESC").
		Find(&documents).Error
	
	return documents, total, err
}

// FindByKnowledgeBaseIDWithPagination 根据知识库ID分页查找文档
func (r *GormDocumentRepository) FindByKnowledgeBaseIDWithPagination(ctx context.Context, knowledgeBaseID string, offset, limit int) ([]*domain.Document, int64, error) {
	var documents []*domain.Document
	var total int64
	
	query := r.db.WithContext(ctx).Model(&domain.Document{}).Where("knowledge_base_id = ?", knowledgeBaseID)
	
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
		Find(&documents).Error
	
	return documents, total, err
}

// SearchByContent 根据内容搜索文档
func (r *GormDocumentRepository) SearchByContent(ctx context.Context, query string, knowledgeBaseID string, limit int) ([]*domain.Document, error) {
	var documents []*domain.Document
	
	dbQuery := r.db.WithContext(ctx).
		Preload("Tags").
		Where("content ILIKE ?", "%"+query+"%")
	
	if knowledgeBaseID != "" {
		dbQuery = dbQuery.Where("knowledge_base_id = ?", knowledgeBaseID)
	}
	
	err := dbQuery.
		Limit(limit).
		Order("created_at DESC").
		Find(&documents).Error
	
	return documents, err
}

// SearchByTitle 根据标题搜索文档
func (r *GormDocumentRepository) SearchByTitle(ctx context.Context, query string, limit int) ([]*domain.Document, error) {
	var documents []*domain.Document
	
	err := r.db.WithContext(ctx).
		Preload("Tags").
		Where("title ILIKE ?", "%"+query+"%").
		Limit(limit).
		Order("created_at DESC").
		Find(&documents).Error
	
	return documents, err
}

// SaveBatch 批量保存文档
func (r *GormDocumentRepository) SaveBatch(ctx context.Context, documents []*domain.Document) error {
	return r.db.WithContext(ctx).CreateInBatches(documents, 100).Error
}

// UpdateBatch 批量更新文档
func (r *GormDocumentRepository) UpdateBatch(ctx context.Context, documents []*domain.Document) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		for _, doc := range documents {
			if err := tx.Save(doc).Error; err != nil {
				return err
			}
		}
		return nil
	})
}

// DeleteBatch 批量删除文档
func (r *GormDocumentRepository) DeleteBatch(ctx context.Context, ids []string) error {
	return r.db.WithContext(ctx).Delete(&domain.Document{}, "id IN ?", ids).Error
}

// CountByKnowledgeBaseID 根据知识库ID统计文档数量
func (r *GormDocumentRepository) CountByKnowledgeBaseID(ctx context.Context, knowledgeBaseID string) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).
		Model(&domain.Document{}).
		Where("knowledge_base_id = ?", knowledgeBaseID).
		Count(&count).Error
	
	return count, err
}

// CountByStatus 根据状态统计文档数量
func (r *GormDocumentRepository) CountByStatus(ctx context.Context, status domain.DocumentStatus) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).
		Model(&domain.Document{}).
		Where("status = ?", status).
		Count(&count).Error
	
	return count, err
}

// GetStatsByKnowledgeBaseID 获取知识库文档统计信息
func (r *GormDocumentRepository) GetStatsByKnowledgeBaseID(ctx context.Context, knowledgeBaseID string) (*repository.DocumentStats, error) {
	stats := &repository.DocumentStats{
		StatusCounts: make(map[domain.DocumentStatus]int64),
		TypeCounts:   make(map[domain.DocumentType]int64),
	}
	
	// 获取总数
	err := r.db.WithContext(ctx).
		Model(&domain.Document{}).
		Where("knowledge_base_id = ?", knowledgeBaseID).
		Count(&stats.TotalCount).Error
	if err != nil {
		return nil, err
	}
	
	// 获取状态统计
	var statusStats []struct {
		Status domain.DocumentStatus
		Count  int64
	}
	err = r.db.WithContext(ctx).
		Model(&domain.Document{}).
		Select("status, count(*) as count").
		Where("knowledge_base_id = ?", knowledgeBaseID).
		Group("status").
		Scan(&statusStats).Error
	if err != nil {
		return nil, err
	}
	
	for _, stat := range statusStats {
		stats.StatusCounts[stat.Status] = stat.Count
		if stat.Status == domain.DocumentStatusIndexed {
			stats.IndexedCount = stat.Count
		}
	}
	
	// 获取类型统计
	var typeStats []struct {
		Type  domain.DocumentType
		Count int64
	}
	err = r.db.WithContext(ctx).
		Model(&domain.Document{}).
		Select("type, count(*) as count").
		Where("knowledge_base_id = ?", knowledgeBaseID).
		Group("type").
		Scan(&typeStats).Error
	if err != nil {
		return nil, err
	}
	
	for _, stat := range typeStats {
		stats.TypeCounts[stat.Type] = stat.Count
	}
	
	// 获取大小统计
	var sizeStats struct {
		TotalSize   int64
		AverageSize float64
	}
	err = r.db.WithContext(ctx).
		Model(&domain.Document{}).
		Select("sum(size) as total_size, avg(size) as average_size").
		Where("knowledge_base_id = ?", knowledgeBaseID).
		Scan(&sizeStats).Error
	if err != nil {
		return nil, err
	}
	
	stats.TotalSize = sizeStats.TotalSize
	stats.AverageSize = sizeStats.AverageSize
	
	// 获取分块统计
	err = r.db.WithContext(ctx).
		Model(&domain.Chunk{}).
		Joins("JOIN documents ON chunks.document_id = documents.id").
		Where("documents.knowledge_base_id = ?", knowledgeBaseID).
		Count(&stats.ChunkCount).Error
	if err != nil {
		return nil, err
	}
	
	return stats, nil
}

// FindPendingIndexing 查找待索引的文档
func (r *GormDocumentRepository) FindPendingIndexing(ctx context.Context, limit int) ([]*domain.Document, error) {
	var documents []*domain.Document
	err := r.db.WithContext(ctx).
		Where("status = ?", domain.DocumentStatusPending).
		Limit(limit).
		Order("created_at ASC").
		Find(&documents).Error
	
	return documents, err
}

// MarkAsIndexing 标记为索引中
func (r *GormDocumentRepository) MarkAsIndexing(ctx context.Context, documentID string) error {
	return r.db.WithContext(ctx).
		Model(&domain.Document{}).
		Where("id = ?", documentID).
		Update("status", domain.DocumentStatusIndexing).Error
}

// MarkAsIndexed 标记为已索引
func (r *GormDocumentRepository) MarkAsIndexed(ctx context.Context, documentID string, chunks []*domain.Chunk) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// 更新文档状态
		now := gorm.Expr("NOW()")
		err := tx.Model(&domain.Document{}).
			Where("id = ?", documentID).
			Updates(map[string]interface{}{
				"status":     domain.DocumentStatusIndexed,
				"indexed_at": now,
				"updated_at": now,
			}).Error
		if err != nil {
			return fmt.Errorf("failed to update document status: %w", err)
		}
		
		// 保存分块
		if len(chunks) > 0 {
			err = tx.CreateInBatches(chunks, 100).Error
			if err != nil {
				return fmt.Errorf("failed to save chunks: %w", err)
			}
		}
		
		return nil
	})
}

// MarkAsIndexingFailed 标记为索引失败
func (r *GormDocumentRepository) MarkAsIndexingFailed(ctx context.Context, documentID string, reason string) error {
	updates := map[string]interface{}{
		"status":     domain.DocumentStatusFailed,
		"updated_at": gorm.Expr("NOW()"),
	}
	
	// 可以将失败原因存储在元数据中
	if reason != "" {
		// 这里可以扩展为存储错误信息的字段
		updates["metadata"] = gorm.Expr("jsonb_set(metadata, '{error}', ?)", fmt.Sprintf(`"%s"`, reason))
	}
	
	return r.db.WithContext(ctx).
		Model(&domain.Document{}).
		Where("id = ?", documentID).
		Updates(updates).Error
}
