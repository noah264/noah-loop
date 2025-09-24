package domain

import (
	"time"

	"github.com/noah-loop/backend/shared/pkg/domain"
)

// DocumentStatus 文档状态
type DocumentStatus string

const (
	DocumentStatusPending   DocumentStatus = "pending"   // 待处理
	DocumentStatusIndexing  DocumentStatus = "indexing"  // 索引中
	DocumentStatusIndexed   DocumentStatus = "indexed"   // 已索引
	DocumentStatusFailed    DocumentStatus = "failed"    // 失败
	DocumentStatusDeleted   DocumentStatus = "deleted"   // 已删除
)

// DocumentType 文档类型
type DocumentType string

const (
	DocumentTypeText     DocumentType = "text"     // 纯文本
	DocumentTypePDF      DocumentType = "pdf"      // PDF文档
	DocumentTypeMarkdown DocumentType = "markdown" // Markdown
	DocumentTypeHTML     DocumentType = "html"     // HTML
	DocumentTypeWord     DocumentType = "word"     // Word文档
)

// Document 文档聚合根
type Document struct {
	domain.Entity
	Title       string         `gorm:"not null" json:"title"`
	Content     string         `gorm:"type:text" json:"content"`
	Type        DocumentType   `gorm:"not null" json:"type"`
	Status      DocumentStatus `gorm:"not null;default:'pending'" json:"status"`
	Source      string         `json:"source"`       // 文档来源
	Hash        string         `gorm:"unique" json:"hash"` // 内容哈希
	Size        int64          `json:"size"`         // 文档大小
	Language    string         `json:"language"`     // 文档语言
	Tags        []Tag          `gorm:"many2many:document_tags;" json:"tags"`
	Chunks      []Chunk        `json:"chunks"`       // 文档分块
	Metadata    DocumentMetadata `gorm:"embedded" json:"metadata"`
	KnowledgeBaseID string `gorm:"index" json:"knowledge_base_id"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
	IndexedAt   *time.Time     `json:"indexed_at,omitempty"`
}

// DocumentMetadata 文档元数据
type DocumentMetadata struct {
	Author      string            `json:"author,omitempty"`
	Keywords    []string          `gorm:"serializer:json" json:"keywords,omitempty"`
	Description string            `json:"description,omitempty"`
	Category    string            `json:"category,omitempty"`
	Version     string            `json:"version,omitempty"`
	Custom      map[string]string `gorm:"serializer:json" json:"custom,omitempty"`
}

// UpdateStatus 更新文档状态
func (d *Document) UpdateStatus(status DocumentStatus) error {
	if !d.isValidStatusTransition(d.Status, status) {
		return NewDomainError("INVALID_STATUS_TRANSITION", "invalid status transition")
	}
	
	d.Status = status
	d.UpdatedAt = time.Now()
	
	if status == DocumentStatusIndexed {
		now := time.Now()
		d.IndexedAt = &now
	}
	
	return nil
}

// MarkAsIndexed 标记为已索引
func (d *Document) MarkAsIndexed(chunks []Chunk) error {
	d.Status = DocumentStatusIndexed
	d.Chunks = chunks
	now := time.Now()
	d.IndexedAt = &now
	d.UpdatedAt = now
	
	return nil
}

// AddTag 添加标签
func (d *Document) AddTag(tag Tag) {
	for _, existingTag := range d.Tags {
		if existingTag.Name == tag.Name {
			return // 标签已存在
		}
	}
	d.Tags = append(d.Tags, tag)
	d.UpdatedAt = time.Now()
}

// RemoveTag 移除标签
func (d *Document) RemoveTag(tagName string) {
	for i, tag := range d.Tags {
		if tag.Name == tagName {
			d.Tags = append(d.Tags[:i], d.Tags[i+1:]...)
			d.UpdatedAt = time.Now()
			break
		}
	}
}

// IsIndexed 检查是否已索引
func (d *Document) IsIndexed() bool {
	return d.Status == DocumentStatusIndexed
}

// CanBeDeleted 检查是否可以删除
func (d *Document) CanBeDeleted() bool {
	return d.Status != DocumentStatusDeleted
}

// GetChunkCount 获取分块数量
func (d *Document) GetChunkCount() int {
	return len(d.Chunks)
}

// isValidStatusTransition 检查状态转换是否有效
func (d *Document) isValidStatusTransition(from, to DocumentStatus) bool {
	validTransitions := map[DocumentStatus][]DocumentStatus{
		DocumentStatusPending: {DocumentStatusIndexing, DocumentStatusFailed, DocumentStatusDeleted},
		DocumentStatusIndexing: {DocumentStatusIndexed, DocumentStatusFailed, DocumentStatusDeleted},
		DocumentStatusIndexed: {DocumentStatusIndexing, DocumentStatusDeleted}, // 可以重新索引
		DocumentStatusFailed: {DocumentStatusIndexing, DocumentStatusDeleted},  // 可以重试
		DocumentStatusDeleted: {}, // 删除状态不能转换到其他状态
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

// NewDocument 创建新文档
func NewDocument(title, content string, docType DocumentType, source string) (*Document, error) {
	if title == "" {
		return nil, NewDomainError("INVALID_TITLE", "document title cannot be empty")
	}
	
	if content == "" {
		return nil, NewDomainError("INVALID_CONTENT", "document content cannot be empty")
	}
	
	hash := calculateContentHash(content)
	
	doc := &Document{
		Entity:   domain.NewEntity(),
		Title:    title,
		Content:  content,
		Type:     docType,
		Status:   DocumentStatusPending,
		Source:   source,
		Hash:     hash,
		Size:     int64(len(content)),
		Language: detectLanguage(content),
		Tags:     make([]Tag, 0),
		Chunks:   make([]Chunk, 0),
		Metadata: DocumentMetadata{
			Custom: make(map[string]string),
		},
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	
	return doc, nil
}

// calculateContentHash 计算内容哈希
func calculateContentHash(content string) string {
	// TODO: 实现内容哈希算法 (SHA-256)
	return "hash_placeholder"
}

// detectLanguage 检测语言
func detectLanguage(content string) string {
	// TODO: 实现语言检测
	return "zh-CN" // 默认中文
}
