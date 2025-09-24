package domain

import (
	"time"

	"github.com/noah-loop/backend/shared/pkg/domain"
)

// KnowledgeBaseStatus 知识库状态
type KnowledgeBaseStatus string

const (
	KnowledgeBaseStatusActive   KnowledgeBaseStatus = "active"   // 活跃
	KnowledgeBaseStatusInactive KnowledgeBaseStatus = "inactive" // 非活跃
	KnowledgeBaseStatusIndexing KnowledgeBaseStatus = "indexing" // 索引中
	KnowledgeBaseStatusDeleted  KnowledgeBaseStatus = "deleted"  // 已删除
)

// KnowledgeBase 知识库聚合根
type KnowledgeBase struct {
	domain.Entity
	Name         string                 `gorm:"not null" json:"name"`
	Description  string                 `json:"description"`
	Status       KnowledgeBaseStatus    `gorm:"not null;default:'active'" json:"status"`
	OwnerID      string                 `gorm:"not null;index" json:"owner_id"`
	Documents    []Document             `json:"documents"`
	Settings     KnowledgeBaseSettings  `gorm:"embedded" json:"settings"`
	Statistics   KnowledgeBaseStats     `gorm:"embedded" json:"statistics"`
	Tags         []Tag                  `gorm:"many2many:knowledge_base_tags;" json:"tags"`
	CreatedAt    time.Time              `json:"created_at"`
	UpdatedAt    time.Time              `json:"updated_at"`
	LastIndexedAt *time.Time            `json:"last_indexed_at,omitempty"`
}

// KnowledgeBaseSettings 知识库设置
type KnowledgeBaseSettings struct {
	ChunkSize       int     `json:"chunk_size" gorm:"default:1000"`        // 分块大小
	ChunkOverlap    int     `json:"chunk_overlap" gorm:"default:200"`      // 分块重叠
	EmbeddingModel  string  `json:"embedding_model" gorm:"default:'text-embedding-ada-002'"` // 嵌入模型
	Language        string  `json:"language" gorm:"default:'zh-CN'"`       // 主要语言
	AutoUpdate      bool    `json:"auto_update" gorm:"default:true"`       // 自动更新索引
	MaxDocuments    int     `json:"max_documents" gorm:"default:10000"`    // 最大文档数
	SimilarityThreshold float32 `json:"similarity_threshold" gorm:"default:0.7"` // 相似度阈值
	EnableMetadata  bool    `json:"enable_metadata" gorm:"default:true"`   // 启用元数据
	EnableVersioning bool   `json:"enable_versioning" gorm:"default:false"` // 启用版本控制
}

// KnowledgeBaseStats 知识库统计信息
type KnowledgeBaseStats struct {
	DocumentCount   int     `json:"document_count"`
	ChunkCount      int     `json:"chunk_count"`
	TotalSize       int64   `json:"total_size"`        // 总大小（字节）
	IndexedCount    int     `json:"indexed_count"`     // 已索引文档数
	AverageSize     float64 `json:"average_size"`      // 平均文档大小
	LastQueryAt     *time.Time `json:"last_query_at"`  // 最后查询时间
	QueryCount      int64   `json:"query_count"`       // 查询总数
	AverageScore    float32 `json:"average_score"`     // 平均检索分数
}

// AddDocument 添加文档到知识库
func (kb *KnowledgeBase) AddDocument(document *Document) error {
	if kb.Status == KnowledgeBaseStatusDeleted {
		return NewDomainError("KNOWLEDGE_BASE_DELETED", "cannot add document to deleted knowledge base")
	}
	
	if kb.Statistics.DocumentCount >= kb.Settings.MaxDocuments {
		return NewDomainError("MAX_DOCUMENTS_REACHED", "maximum number of documents reached")
	}
	
	document.KnowledgeBaseID = kb.ID
	kb.Documents = append(kb.Documents, *document)
	kb.updateStatistics()
	kb.UpdatedAt = time.Now()
	
	return nil
}

// RemoveDocument 从知识库移除文档
func (kb *KnowledgeBase) RemoveDocument(documentID string) error {
	if kb.Status == KnowledgeBaseStatusDeleted {
		return NewDomainError("KNOWLEDGE_BASE_DELETED", "cannot remove document from deleted knowledge base")
	}
	
	for i, doc := range kb.Documents {
		if doc.ID == documentID {
			kb.Documents = append(kb.Documents[:i], kb.Documents[i+1:]...)
			kb.updateStatistics()
			kb.UpdatedAt = time.Now()
			return nil
		}
	}
	
	return NewDomainError("DOCUMENT_NOT_FOUND", "document not found in knowledge base")
}

// UpdateSettings 更新知识库设置
func (kb *KnowledgeBase) UpdateSettings(settings KnowledgeBaseSettings) error {
	if kb.Status == KnowledgeBaseStatusDeleted {
		return NewDomainError("KNOWLEDGE_BASE_DELETED", "cannot update settings of deleted knowledge base")
	}
	
	// 验证设置
	if settings.ChunkSize <= 0 {
		return NewDomainError("INVALID_CHUNK_SIZE", "chunk size must be positive")
	}
	
	if settings.ChunkOverlap < 0 || settings.ChunkOverlap >= settings.ChunkSize {
		return NewDomainError("INVALID_CHUNK_OVERLAP", "chunk overlap must be non-negative and less than chunk size")
	}
	
	if settings.SimilarityThreshold < 0 || settings.SimilarityThreshold > 1 {
		return NewDomainError("INVALID_SIMILARITY_THRESHOLD", "similarity threshold must be between 0 and 1")
	}
	
	kb.Settings = settings
	kb.UpdatedAt = time.Now()
	
	return nil
}

// UpdateStatus 更新知识库状态
func (kb *KnowledgeBase) UpdateStatus(status KnowledgeBaseStatus) error {
	if !kb.isValidStatusTransition(kb.Status, status) {
		return NewDomainError("INVALID_STATUS_TRANSITION", "invalid status transition")
	}
	
	kb.Status = status
	kb.UpdatedAt = time.Now()
	
	return nil
}

// AddTag 添加标签
func (kb *KnowledgeBase) AddTag(tag Tag) {
	for _, existingTag := range kb.Tags {
		if existingTag.Name == tag.Name {
			return // 标签已存在
		}
	}
	kb.Tags = append(kb.Tags, tag)
	kb.UpdatedAt = time.Now()
}

// RecordQuery 记录查询统计
func (kb *KnowledgeBase) RecordQuery(score float32) {
	now := time.Now()
	kb.Statistics.LastQueryAt = &now
	kb.Statistics.QueryCount++
	
	// 计算平均分数
	if kb.Statistics.QueryCount == 1 {
		kb.Statistics.AverageScore = score
	} else {
		kb.Statistics.AverageScore = (kb.Statistics.AverageScore*float32(kb.Statistics.QueryCount-1) + score) / float32(kb.Statistics.QueryCount)
	}
	
	kb.UpdatedAt = now
}

// CanBeQueried 检查是否可以查询
func (kb *KnowledgeBase) CanBeQueried() bool {
	return kb.Status == KnowledgeBaseStatusActive && kb.Statistics.IndexedCount > 0
}

// IsIndexing 检查是否正在索引
func (kb *KnowledgeBase) IsIndexing() bool {
	return kb.Status == KnowledgeBaseStatusIndexing
}

// MarkAsIndexed 标记为已索引
func (kb *KnowledgeBase) MarkAsIndexed() {
	if kb.Status == KnowledgeBaseStatusIndexing {
		kb.Status = KnowledgeBaseStatusActive
	}
	now := time.Now()
	kb.LastIndexedAt = &now
	kb.UpdatedAt = now
}

// updateStatistics 更新统计信息
func (kb *KnowledgeBase) updateStatistics() {
	kb.Statistics.DocumentCount = len(kb.Documents)
	kb.Statistics.IndexedCount = 0
	kb.Statistics.TotalSize = 0
	kb.Statistics.ChunkCount = 0
	
	for _, doc := range kb.Documents {
		kb.Statistics.TotalSize += doc.Size
		kb.Statistics.ChunkCount += len(doc.Chunks)
		if doc.IsIndexed() {
			kb.Statistics.IndexedCount++
		}
	}
	
	if kb.Statistics.DocumentCount > 0 {
		kb.Statistics.AverageSize = float64(kb.Statistics.TotalSize) / float64(kb.Statistics.DocumentCount)
	}
}

// isValidStatusTransition 检查状态转换是否有效
func (kb *KnowledgeBase) isValidStatusTransition(from, to KnowledgeBaseStatus) bool {
	validTransitions := map[KnowledgeBaseStatus][]KnowledgeBaseStatus{
		KnowledgeBaseStatusActive:   {KnowledgeBaseStatusInactive, KnowledgeBaseStatusIndexing, KnowledgeBaseStatusDeleted},
		KnowledgeBaseStatusInactive: {KnowledgeBaseStatusActive, KnowledgeBaseStatusDeleted},
		KnowledgeBaseStatusIndexing: {KnowledgeBaseStatusActive, KnowledgeBaseStatusInactive, KnowledgeBaseStatusDeleted},
		KnowledgeBaseStatusDeleted:  {}, // 删除状态不能转换
	}
	
	allowedStates, exists := validTransitions[from]
	if !exists {
		return false
	}
	
	for _, allowed := range allowedStates {
		if allowed == to {
			return true
		}
	}
	
	return false
}

// NewKnowledgeBase 创建新的知识库
func NewKnowledgeBase(name, description, ownerID string) (*KnowledgeBase, error) {
	if name == "" {
		return nil, NewDomainError("INVALID_NAME", "knowledge base name cannot be empty")
	}
	
	if ownerID == "" {
		return nil, NewDomainError("INVALID_OWNER_ID", "owner ID cannot be empty")
	}
	
	kb := &KnowledgeBase{
		Entity:      domain.NewEntity(),
		Name:        name,
		Description: description,
		Status:      KnowledgeBaseStatusActive,
		OwnerID:     ownerID,
		Documents:   make([]Document, 0),
		Settings:    KnowledgeBaseSettings{
			ChunkSize:           1000,
			ChunkOverlap:        200,
			EmbeddingModel:      "text-embedding-ada-002",
			Language:            "zh-CN",
			AutoUpdate:          true,
			MaxDocuments:        10000,
			SimilarityThreshold: 0.7,
			EnableMetadata:      true,
			EnableVersioning:    false,
		},
		Statistics: KnowledgeBaseStats{},
		Tags:       make([]Tag, 0),
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
	}
	
	return kb, nil
}
