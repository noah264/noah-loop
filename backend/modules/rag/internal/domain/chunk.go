package domain

import (
	"time"

	"github.com/noah-loop/backend/shared/pkg/domain"
)

// ChunkType 分块类型
type ChunkType string

const (
	ChunkTypeText      ChunkType = "text"      // 文本分块
	ChunkTypeParagraph ChunkType = "paragraph" // 段落分块
	ChunkTypeSentence  ChunkType = "sentence"  // 句子分块
	ChunkTypeSection   ChunkType = "section"   // 章节分块
	ChunkTypeTable     ChunkType = "table"     // 表格分块
	ChunkTypeCode      ChunkType = "code"      // 代码分块
)

// Chunk 文档分块实体
type Chunk struct {
	domain.Entity
	DocumentID   string             `gorm:"not null;index" json:"document_id"`
	Content      string             `gorm:"type:text;not null" json:"content"`
	Type         ChunkType          `gorm:"not null" json:"type"`
	Position     int                `gorm:"not null" json:"position"`     // 在文档中的位置
	StartIndex   int                `json:"start_index"`                  // 在原文档中的开始索引
	EndIndex     int                `json:"end_index"`                    // 在原文档中的结束索引
	TokenCount   int                `json:"token_count"`                  // 令牌数量
	Embedding    []float32          `gorm:"type:jsonb" json:"embedding"`  // 向量嵌入
	Metadata     ChunkMetadata      `gorm:"embedded" json:"metadata"`
	Similarities []ChunkSimilarity  `json:"similarities,omitempty"`      // 相似度缓存
	CreatedAt    time.Time          `json:"created_at"`
	UpdatedAt    time.Time          `json:"updated_at"`
	EmbeddedAt   *time.Time         `json:"embedded_at,omitempty"`
}

// ChunkMetadata 分块元数据
type ChunkMetadata struct {
	Title     string            `json:"title,omitempty"`
	Section   string            `json:"section,omitempty"`
	Keywords  []string          `gorm:"serializer:json" json:"keywords,omitempty"`
	Entities  []string          `gorm:"serializer:json" json:"entities,omitempty"` // 实体识别
	Sentiment string            `json:"sentiment,omitempty"`                        // 情感分析
	Custom    map[string]string `gorm:"serializer:json" json:"custom,omitempty"`
}

// ChunkSimilarity 分块相似度
type ChunkSimilarity struct {
	ChunkID    string  `json:"chunk_id"`
	Similarity float32 `json:"similarity"`
	Type       string  `json:"type"` // semantic, lexical, etc.
}

// SetEmbedding 设置向量嵌入
func (c *Chunk) SetEmbedding(embedding []float32) error {
	if len(embedding) == 0 {
		return NewDomainError("INVALID_EMBEDDING", "embedding cannot be empty")
	}
	
	c.Embedding = embedding
	now := time.Now()
	c.EmbeddedAt = &now
	c.UpdatedAt = now
	
	return nil
}

// HasEmbedding 检查是否有向量嵌入
func (c *Chunk) HasEmbedding() bool {
	return len(c.Embedding) > 0
}

// UpdateMetadata 更新元数据
func (c *Chunk) UpdateMetadata(metadata ChunkMetadata) {
	c.Metadata = metadata
	c.UpdatedAt = time.Now()
}

// AddKeyword 添加关键词
func (c *Chunk) AddKeyword(keyword string) {
	for _, existing := range c.Metadata.Keywords {
		if existing == keyword {
			return // 关键词已存在
		}
	}
	c.Metadata.Keywords = append(c.Metadata.Keywords, keyword)
	c.UpdatedAt = time.Now()
}

// AddEntity 添加实体
func (c *Chunk) AddEntity(entity string) {
	for _, existing := range c.Metadata.Entities {
		if existing == entity {
			return // 实体已存在
		}
	}
	c.Metadata.Entities = append(c.Metadata.Entities, entity)
	c.UpdatedAt = time.Now()
}

// GetContentPreview 获取内容预览
func (c *Chunk) GetContentPreview(length int) string {
	if length <= 0 || length >= len(c.Content) {
		return c.Content
	}
	return c.Content[:length] + "..."
}

// CalculateTokenCount 计算令牌数量
func (c *Chunk) CalculateTokenCount() int {
	// TODO: 实现更准确的令牌计算
	// 简单实现：按字符数除以4估算
	c.TokenCount = len(c.Content) / 4
	c.UpdatedAt = time.Now()
	return c.TokenCount
}

// NewChunk 创建新的文档分块
func NewChunk(documentID, content string, chunkType ChunkType, position int) (*Chunk, error) {
	if documentID == "" {
		return nil, NewDomainError("INVALID_DOCUMENT_ID", "document ID cannot be empty")
	}
	
	if content == "" {
		return nil, NewDomainError("INVALID_CONTENT", "chunk content cannot be empty")
	}
	
	if position < 0 {
		return nil, NewDomainError("INVALID_POSITION", "chunk position cannot be negative")
	}
	
	chunk := &Chunk{
		Entity:     domain.NewEntity(),
		DocumentID: documentID,
		Content:    content,
		Type:       chunkType,
		Position:   position,
		StartIndex: 0, // 需要在分块时计算
		EndIndex:   len(content),
		Metadata: ChunkMetadata{
			Custom: make(map[string]string),
		},
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	
	chunk.CalculateTokenCount()
	
	return chunk, nil
}
