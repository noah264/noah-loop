package repository

import (
	"context"

	"github.com/noah-loop/backend/modules/rag/internal/domain"
)

// DocumentRepository 文档仓储接口
type DocumentRepository interface {
	// 基本CRUD操作
	Save(ctx context.Context, document *domain.Document) error
	FindByID(ctx context.Context, id string) (*domain.Document, error)
	FindByHash(ctx context.Context, hash string) (*domain.Document, error)
	Update(ctx context.Context, document *domain.Document) error
	Delete(ctx context.Context, id string) error

	// 查询操作
	FindByKnowledgeBaseID(ctx context.Context, knowledgeBaseID string) ([]*domain.Document, error)
	FindByStatus(ctx context.Context, status domain.DocumentStatus) ([]*domain.Document, error)
	FindByType(ctx context.Context, docType domain.DocumentType) ([]*domain.Document, error)
	FindByTags(ctx context.Context, tagNames []string) ([]*domain.Document, error)

	// 分页查询
	FindWithPagination(ctx context.Context, offset, limit int) ([]*domain.Document, int64, error)
	FindByKnowledgeBaseIDWithPagination(ctx context.Context, knowledgeBaseID string, offset, limit int) ([]*domain.Document, int64, error)

	// 搜索操作
	SearchByContent(ctx context.Context, query string, knowledgeBaseID string, limit int) ([]*domain.Document, error)
	SearchByTitle(ctx context.Context, query string, limit int) ([]*domain.Document, error)

	// 批量操作
	SaveBatch(ctx context.Context, documents []*domain.Document) error
	UpdateBatch(ctx context.Context, documents []*domain.Document) error
	DeleteBatch(ctx context.Context, ids []string) error

	// 统计操作
	CountByKnowledgeBaseID(ctx context.Context, knowledgeBaseID string) (int64, error)
	CountByStatus(ctx context.Context, status domain.DocumentStatus) (int64, error)
	GetStatsByKnowledgeBaseID(ctx context.Context, knowledgeBaseID string) (*DocumentStats, error)

	// 索引相关
	FindPendingIndexing(ctx context.Context, limit int) ([]*domain.Document, error)
	MarkAsIndexing(ctx context.Context, documentID string) error
	MarkAsIndexed(ctx context.Context, documentID string, chunks []*domain.Chunk) error
	MarkAsIndexingFailed(ctx context.Context, documentID string, reason string) error
}

// DocumentStats 文档统计信息
type DocumentStats struct {
	TotalCount    int64                               `json:"total_count"`
	StatusCounts  map[domain.DocumentStatus]int64    `json:"status_counts"`
	TypeCounts    map[domain.DocumentType]int64      `json:"type_counts"`
	TotalSize     int64                               `json:"total_size"`
	AverageSize   float64                             `json:"average_size"`
	IndexedCount  int64                               `json:"indexed_count"`
	ChunkCount    int64                               `json:"chunk_count"`
}
