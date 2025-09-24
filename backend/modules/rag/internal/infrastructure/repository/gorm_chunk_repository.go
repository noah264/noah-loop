package repository

import (
	"context"

	"github.com/noah-loop/backend/modules/rag/internal/domain"
	"github.com/noah-loop/backend/modules/rag/internal/domain/repository"
	"gorm.io/gorm"
)

// GormChunkRepository GORM分块仓储实现
type GormChunkRepository struct {
	db *gorm.DB
}

// NewGormChunkRepository 创建GORM分块仓储
func NewGormChunkRepository(db *gorm.DB) repository.ChunkRepository {
	return &GormChunkRepository{
		db: db,
	}
}

// Save 保存分块
func (r *GormChunkRepository) Save(ctx context.Context, chunk *domain.Chunk) error {
	return r.db.WithContext(ctx).Create(chunk).Error
}

// FindByID 根据ID查找分块
func (r *GormChunkRepository) FindByID(ctx context.Context, id string) (*domain.Chunk, error) {
	var chunk domain.Chunk
	err := r.db.WithContext(ctx).First(&chunk, "id = ?", id).Error
	
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}
	
	return &chunk, nil
}

// Update 更新分块
func (r *GormChunkRepository) Update(ctx context.Context, chunk *domain.Chunk) error {
	return r.db.WithContext(ctx).Save(chunk).Error
}

// Delete 删除分块
func (r *GormChunkRepository) Delete(ctx context.Context, id string) error {
	return r.db.WithContext(ctx).Delete(&domain.Chunk{}, "id = ?", id).Error
}

// FindByDocumentID 根据文档ID查找分块
func (r *GormChunkRepository) FindByDocumentID(ctx context.Context, documentID string) ([]*domain.Chunk, error) {
	var chunks []*domain.Chunk
	err := r.db.WithContext(ctx).
		Where("document_id = ?", documentID).
		Order("position ASC").
		Find(&chunks).Error
	
	return chunks, err
}

// FindByDocumentIDWithPagination 根据文档ID分页查找分块
func (r *GormChunkRepository) FindByDocumentIDWithPagination(ctx context.Context, documentID string, offset, limit int) ([]*domain.Chunk, int64, error) {
	var chunks []*domain.Chunk
	var total int64
	
	query := r.db.WithContext(ctx).Model(&domain.Chunk{}).Where("document_id = ?", documentID)
	
	// 获取总数
	err := query.Count(&total).Error
	if err != nil {
		return nil, 0, err
	}
	
	// 获取分页数据
	err = query.
		Offset(offset).
		Limit(limit).
		Order("position ASC").
		Find(&chunks).Error
	
	return chunks, total, err
}

// FindByType 根据类型查找分块
func (r *GormChunkRepository) FindByType(ctx context.Context, chunkType domain.ChunkType) ([]*domain.Chunk, error) {
	var chunks []*domain.Chunk
	err := r.db.WithContext(ctx).
		Where("type = ?", chunkType).
		Order("created_at DESC").
		Find(&chunks).Error
	
	return chunks, err
}

// FindWithoutEmbedding 查找没有嵌入向量的分块
func (r *GormChunkRepository) FindWithoutEmbedding(ctx context.Context, limit int) ([]*domain.Chunk, error) {
	var chunks []*domain.Chunk
	err := r.db.WithContext(ctx).
		Where("embedding IS NULL OR jsonb_array_length(embedding) = 0").
		Limit(limit).
		Order("created_at ASC").
		Find(&chunks).Error
	
	return chunks, err
}

// UpdateEmbedding 更新嵌入向量
func (r *GormChunkRepository) UpdateEmbedding(ctx context.Context, chunkID string, embedding []float32) error {
	now := gorm.Expr("NOW()")
	return r.db.WithContext(ctx).
		Model(&domain.Chunk{}).
		Where("id = ?", chunkID).
		Updates(map[string]interface{}{
			"embedding":    embedding,
			"embedded_at":  now,
			"updated_at":   now,
		}).Error
}

// UpdateEmbeddingBatch 批量更新嵌入向量
func (r *GormChunkRepository) UpdateEmbeddingBatch(ctx context.Context, chunkEmbeddings map[string][]float32) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		for chunkID, embedding := range chunkEmbeddings {
			now := gorm.Expr("NOW()")
			err := tx.Model(&domain.Chunk{}).
				Where("id = ?", chunkID).
				Updates(map[string]interface{}{
					"embedding":   embedding,
					"embedded_at": now,
					"updated_at":  now,
				}).Error
			if err != nil {
				return err
			}
		}
		return nil
	})
}

// SaveBatch 批量保存分块
func (r *GormChunkRepository) SaveBatch(ctx context.Context, chunks []*domain.Chunk) error {
	return r.db.WithContext(ctx).CreateInBatches(chunks, 100).Error
}

// UpdateBatch 批量更新分块
func (r *GormChunkRepository) UpdateBatch(ctx context.Context, chunks []*domain.Chunk) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		for _, chunk := range chunks {
			if err := tx.Save(chunk).Error; err != nil {
				return err
			}
		}
		return nil
	})
}

// DeleteBatch 批量删除分块
func (r *GormChunkRepository) DeleteBatch(ctx context.Context, ids []string) error {
	return r.db.WithContext(ctx).Delete(&domain.Chunk{}, "id IN ?", ids).Error
}

// DeleteByDocumentID 根据文档ID删除分块
func (r *GormChunkRepository) DeleteByDocumentID(ctx context.Context, documentID string) error {
	return r.db.WithContext(ctx).Delete(&domain.Chunk{}, "document_id = ?", documentID).Error
}

// CountByDocumentID 根据文档ID统计分块数量
func (r *GormChunkRepository) CountByDocumentID(ctx context.Context, documentID string) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).
		Model(&domain.Chunk{}).
		Where("document_id = ?", documentID).
		Count(&count).Error
	
	return count, err
}

// CountByType 根据类型统计分块数量
func (r *GormChunkRepository) CountByType(ctx context.Context, chunkType domain.ChunkType) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).
		Model(&domain.Chunk{}).
		Where("type = ?", chunkType).
		Count(&count).Error
	
	return count, err
}

// CountWithEmbedding 统计有嵌入向量的分块数量
func (r *GormChunkRepository) CountWithEmbedding(ctx context.Context) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).
		Model(&domain.Chunk{}).
		Where("embedding IS NOT NULL AND jsonb_array_length(embedding) > 0").
		Count(&count).Error
	
	return count, err
}

// CountWithoutEmbedding 统计没有嵌入向量的分块数量
func (r *GormChunkRepository) CountWithoutEmbedding(ctx context.Context) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).
		Model(&domain.Chunk{}).
		Where("embedding IS NULL OR jsonb_array_length(embedding) = 0").
		Count(&count).Error
	
	return count, err
}
