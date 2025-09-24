package service

import "github.com/noah-loop/backend/modules/rag/internal/domain"

// CreateKnowledgeBaseCommand 创建知识库命令
type CreateKnowledgeBaseCommand struct {
	Name        string                            `json:"name" binding:"required"`
	Description string                            `json:"description"`
	OwnerID     string                            `json:"owner_id" binding:"required"`
	Settings    *domain.KnowledgeBaseSettings     `json:"settings,omitempty"`
	Tags        []string                          `json:"tags,omitempty"`
}

// UpdateKnowledgeBaseCommand 更新知识库命令
type UpdateKnowledgeBaseCommand struct {
	ID          string                            `json:"id" binding:"required"`
	Name        string                            `json:"name,omitempty"`
	Description string                            `json:"description,omitempty"`
	Status      domain.KnowledgeBaseStatus        `json:"status,omitempty"`
	Settings    *domain.KnowledgeBaseSettings     `json:"settings,omitempty"`
	Tags        []string                          `json:"tags,omitempty"`
}

// DeleteKnowledgeBaseCommand 删除知识库命令
type DeleteKnowledgeBaseCommand struct {
	ID string `json:"id" binding:"required"`
}

// AddDocumentCommand 添加文档命令
type AddDocumentCommand struct {
	Title           string                    `json:"title" binding:"required"`
	Content         string                    `json:"content" binding:"required"`
	Type            domain.DocumentType       `json:"type" binding:"required"`
	Source          string                    `json:"source"`
	Language        string                    `json:"language"`
	KnowledgeBaseID string                    `json:"knowledge_base_id" binding:"required"`
	Metadata        *domain.DocumentMetadata  `json:"metadata,omitempty"`
	Tags            []string                  `json:"tags,omitempty"`
}

// UpdateDocumentCommand 更新文档命令
type UpdateDocumentCommand struct {
	ID          string                    `json:"id" binding:"required"`
	Title       string                    `json:"title,omitempty"`
	Content     string                    `json:"content,omitempty"`
	Status      domain.DocumentStatus     `json:"status,omitempty"`
	Metadata    *domain.DocumentMetadata  `json:"metadata,omitempty"`
	Tags        []string                  `json:"tags,omitempty"`
}

// DeleteDocumentCommand 删除文档命令
type DeleteDocumentCommand struct {
	ID string `json:"id" binding:"required"`
}

// ProcessDocumentCommand 处理文档命令
type ProcessDocumentCommand struct {
	DocumentID string `json:"document_id" binding:"required"`
	ForceReprocess bool `json:"force_reprocess"`
}

// SearchCommand 搜索命令
type SearchCommand struct {
	Query           string                `json:"query" binding:"required"`
	KnowledgeBaseID string                `json:"knowledge_base_id" binding:"required"`
	TopK            int                   `json:"top_k"`
	ScoreThreshold  float32               `json:"score_threshold"`
	SearchType      domain.SearchType     `json:"search_type"`
	Filters         *domain.SearchFilters `json:"filters,omitempty"`
	Rerank          bool                  `json:"rerank"`
	IncludeMetadata bool                  `json:"include_metadata"`
}

// ToSearchQuery 转换为搜索查询
func (cmd *SearchCommand) ToSearchQuery() *domain.SearchQuery {
	query := domain.NewSearchQuery(cmd.Query, cmd.KnowledgeBaseID)
	
	if cmd.TopK > 0 {
		query.WithTopK(cmd.TopK)
	}
	
	if cmd.ScoreThreshold > 0 {
		query.WithScoreThreshold(cmd.ScoreThreshold)
	}
	
	if cmd.SearchType != "" {
		query.WithSearchType(cmd.SearchType)
	}
	
	if cmd.Filters != nil {
		query.WithFilters(*cmd.Filters)
	}
	
	query.Rerank = cmd.Rerank
	query.IncludeMetadata = cmd.IncludeMetadata
	
	return query
}

// BatchAddDocumentsCommand 批量添加文档命令
type BatchAddDocumentsCommand struct {
	KnowledgeBaseID string               `json:"knowledge_base_id" binding:"required"`
	Documents       []AddDocumentCommand `json:"documents" binding:"required"`
}

// BatchDeleteDocumentsCommand 批量删除文档命令
type BatchDeleteDocumentsCommand struct {
	DocumentIDs []string `json:"document_ids" binding:"required"`
}

// ReindexKnowledgeBaseCommand 重新索引知识库命令
type ReindexKnowledgeBaseCommand struct {
	KnowledgeBaseID string `json:"knowledge_base_id" binding:"required"`
	ForceReindex    bool   `json:"force_reindex"`
}

// GetKnowledgeBaseStatsCommand 获取知识库统计命令
type GetKnowledgeBaseStatsCommand struct {
	KnowledgeBaseID string `json:"knowledge_base_id" binding:"required"`
}

// ListKnowledgeBasesCommand 列出知识库命令
type ListKnowledgeBasesCommand struct {
	OwnerID string `json:"owner_id"`
	Status  string `json:"status,omitempty"`
	Offset  int    `json:"offset"`
	Limit   int    `json:"limit"`
}

// ListDocumentsCommand 列出文档命令
type ListDocumentsCommand struct {
	KnowledgeBaseID string `json:"knowledge_base_id" binding:"required"`
	Status          string `json:"status,omitempty"`
	Type            string `json:"type,omitempty"`
	Offset          int    `json:"offset"`
	Limit           int    `json:"limit"`
}

// GetDocumentCommand 获取文档命令
type GetDocumentCommand struct {
	ID              string `json:"id" binding:"required"`
	IncludeContent  bool   `json:"include_content"`
	IncludeChunks   bool   `json:"include_chunks"`
}

// GetKnowledgeBaseCommand 获取知识库命令
type GetKnowledgeBaseCommand struct {
	ID               string `json:"id" binding:"required"`
	IncludeDocuments bool   `json:"include_documents"`
	IncludeStats     bool   `json:"include_stats"`
}
