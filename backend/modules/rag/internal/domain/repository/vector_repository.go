package repository

import (
	"context"

	"github.com/noah-loop/backend/modules/rag/internal/domain"
)

// VectorRepository 向量仓储接口
type VectorRepository interface {
	// 向量索引管理
	CreateIndex(ctx context.Context, indexName string, dimension int, metricType MetricType) error
	DeleteIndex(ctx context.Context, indexName string) error
	ListIndexes(ctx context.Context) ([]IndexInfo, error)
	GetIndexInfo(ctx context.Context, indexName string) (*IndexInfo, error)

	// 向量存储
	Insert(ctx context.Context, indexName string, vectors []VectorRecord) error
	Update(ctx context.Context, indexName string, vectors []VectorRecord) error
	Delete(ctx context.Context, indexName string, ids []string) error

	// 向量搜索
	Search(ctx context.Context, query *VectorQuery) (*VectorSearchResult, error)
	SearchBatch(ctx context.Context, queries []*VectorQuery) ([]*VectorSearchResult, error)

	// 相似度计算
	ComputeSimilarity(ctx context.Context, vector1, vector2 []float32, metricType MetricType) (float32, error)
	ComputeSimilarityBatch(ctx context.Context, queryVector []float32, vectors [][]float32, metricType MetricType) ([]float32, error)

	// 统计信息
	GetVectorCount(ctx context.Context, indexName string) (int64, error)
	GetIndexStats(ctx context.Context, indexName string) (*IndexStats, error)

	// 健康检查
	Health(ctx context.Context) error
}

// MetricType 距离度量类型
type MetricType string

const (
	MetricTypeCosine     MetricType = "cosine"      // 余弦相似度
	MetricTypeEuclidean  MetricType = "euclidean"   // 欧氏距离
	MetricTypeDotProduct MetricType = "dot_product" // 点积
	MetricTypeHamming    MetricType = "hamming"     // 汉明距离
)

// VectorRecord 向量记录
type VectorRecord struct {
	ID       string            `json:"id"`
	Vector   []float32         `json:"vector"`
	Metadata map[string]string `json:"metadata"`
}

// VectorQuery 向量查询
type VectorQuery struct {
	IndexName      string            `json:"index_name"`
	QueryVector    []float32         `json:"query_vector"`
	TopK           int               `json:"top_k"`
	ScoreThreshold float32           `json:"score_threshold"`
	MetricType     MetricType        `json:"metric_type"`
	Filter         map[string]string `json:"filter"`          // 元数据过滤
	IncludeVector  bool              `json:"include_vector"`  // 是否返回向量
	IncludeMetadata bool             `json:"include_metadata"` // 是否返回元数据
}

// VectorSearchResult 向量搜索结果
type VectorSearchResult struct {
	Query    *VectorQuery        `json:"query"`
	Results  []VectorSearchMatch `json:"results"`
	Total    int                 `json:"total"`
	Duration int64               `json:"duration"` // 毫秒
}

// VectorSearchMatch 向量搜索匹配结果
type VectorSearchMatch struct {
	ID       string            `json:"id"`
	Score    float32           `json:"score"`
	Vector   []float32         `json:"vector,omitempty"`
	Metadata map[string]string `json:"metadata,omitempty"`
}

// IndexInfo 索引信息
type IndexInfo struct {
	Name        string     `json:"name"`
	Dimension   int        `json:"dimension"`
	MetricType  MetricType `json:"metric_type"`
	VectorCount int64      `json:"vector_count"`
	IndexSize   int64      `json:"index_size"` // 字节
	CreatedAt   string     `json:"created_at"`
	UpdatedAt   string     `json:"updated_at"`
}

// IndexStats 索引统计信息
type IndexStats struct {
	VectorCount    int64   `json:"vector_count"`
	IndexSize      int64   `json:"index_size"`      // 字节
	MemoryUsage    int64   `json:"memory_usage"`    // 字节
	QueryCount     int64   `json:"query_count"`     // 查询次数
	AverageLatency float64 `json:"average_latency"` // 平均延迟（毫秒）
	LastQueryAt    string  `json:"last_query_at"`
}

// NewVectorQuery 创建向量查询
func NewVectorQuery(indexName string, queryVector []float32, topK int) *VectorQuery {
	return &VectorQuery{
		IndexName:       indexName,
		QueryVector:     queryVector,
		TopK:            topK,
		ScoreThreshold:  0.0,
		MetricType:      MetricTypeCosine,
		Filter:          make(map[string]string),
		IncludeVector:   false,
		IncludeMetadata: true,
	}
}

// WithScoreThreshold 设置分数阈值
func (vq *VectorQuery) WithScoreThreshold(threshold float32) *VectorQuery {
	vq.ScoreThreshold = threshold
	return vq
}

// WithMetricType 设置度量类型
func (vq *VectorQuery) WithMetricType(metricType MetricType) *VectorQuery {
	vq.MetricType = metricType
	return vq
}

// WithFilter 设置元数据过滤
func (vq *VectorQuery) WithFilter(key, value string) *VectorQuery {
	if vq.Filter == nil {
		vq.Filter = make(map[string]string)
	}
	vq.Filter[key] = value
	return vq
}

// WithIncludeVector 设置是否返回向量
func (vq *VectorQuery) WithIncludeVector(include bool) *VectorQuery {
	vq.IncludeVector = include
	return vq
}

// WithIncludeMetadata 设置是否返回元数据
func (vq *VectorQuery) WithIncludeMetadata(include bool) *VectorQuery {
	vq.IncludeMetadata = include
	return vq
}
