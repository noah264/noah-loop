package repository

import (
	"context"

	"github.com/noah-loop/backend/modules/rag/internal/domain"
)

// ChunkRepository 分块仓储接口
type ChunkRepository interface {
	// 基本CRUD操作
	Save(ctx context.Context, chunk *domain.Chunk) error
	FindByID(ctx context.Context, id string) (*domain.Chunk, error)
	Update(ctx context.Context, chunk *domain.Chunk) error
	Delete(ctx context.Context, id string) error

	// 查询操作
	FindByDocumentID(ctx context.Context, documentID string) ([]*domain.Chunk, error)
	FindByDocumentIDWithPagination(ctx context.Context, documentID string, offset, limit int) ([]*domain.Chunk, int64, error)
	FindByType(ctx context.Context, chunkType domain.ChunkType) ([]*domain.Chunk, error)

	// 向量相关操作
	FindWithoutEmbedding(ctx context.Context, limit int) ([]*domain.Chunk, error)
	UpdateEmbedding(ctx context.Context, chunkID string, embedding []float32) error
	UpdateEmbeddingBatch(ctx context.Context, chunkEmbeddings map[string][]float32) error

	// 批量操作
	SaveBatch(ctx context.Context, chunks []*domain.Chunk) error
	UpdateBatch(ctx context.Context, chunks []*domain.Chunk) error
	DeleteBatch(ctx context.Context, ids []string) error
	DeleteByDocumentID(ctx context.Context, documentID string) error

	// 统计操作
	CountByDocumentID(ctx context.Context, documentID string) (int64, error)
	CountByType(ctx context.Context, chunkType domain.ChunkType) (int64, error)
	CountWithEmbedding(ctx context.Context) (int64, error)
	CountWithoutEmbedding(ctx context.Context) (int64, error)
}
