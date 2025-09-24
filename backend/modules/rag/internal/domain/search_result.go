package domain

import (
	"time"
)

// SearchResultType 搜索结果类型
type SearchResultType string

const (
	SearchResultTypeDocument SearchResultType = "document" // 文档结果
	SearchResultTypeChunk    SearchResultType = "chunk"    // 分块结果
	SearchResultTypeMixed    SearchResultType = "mixed"    // 混合结果
)

// SearchResult 搜索结果值对象
type SearchResult struct {
	ID          string            `json:"id"`
	Type        SearchResultType  `json:"type"`
	Score       float32           `json:"score"`        // 相似度分数
	Content     string            `json:"content"`      // 内容
	Title       string            `json:"title"`        // 标题
	Source      string            `json:"source"`       // 来源
	Metadata    map[string]string `json:"metadata"`     // 元数据
	Highlight   string            `json:"highlight"`    // 高亮片段
	ChunkInfo   *ChunkInfo        `json:"chunk_info,omitempty"` // 分块信息
	DocumentInfo *DocumentInfo    `json:"document_info,omitempty"` // 文档信息
	SearchedAt  time.Time         `json:"searched_at"`  // 搜索时间
}

// ChunkInfo 分块信息
type ChunkInfo struct {
	Position    int    `json:"position"`     // 在文档中的位置
	StartIndex  int    `json:"start_index"`  // 开始索引
	EndIndex    int    `json:"end_index"`    // 结束索引
	TokenCount  int    `json:"token_count"`  // 令牌数量
	ChunkType   string `json:"chunk_type"`   // 分块类型
}

// DocumentInfo 文档信息
type DocumentInfo struct {
	DocumentID   string    `json:"document_id"`
	DocumentType string    `json:"document_type"`
	Language     string    `json:"language"`
	Size         int64     `json:"size"`
	CreatedAt    time.Time `json:"created_at"`
	IndexedAt    time.Time `json:"indexed_at"`
}

// SearchQuery 搜索查询值对象
type SearchQuery struct {
	Query         string            `json:"query"`           // 查询文本
	KnowledgeBaseID string          `json:"knowledge_base_id"` // 知识库ID
	TopK          int               `json:"top_k"`           // 返回结果数量
	ScoreThreshold float32          `json:"score_threshold"` // 分数阈值
	Filters       SearchFilters     `json:"filters"`         // 过滤条件
	SearchType    SearchType        `json:"search_type"`     // 搜索类型
	Rerank        bool              `json:"rerank"`          // 是否重排序
	IncludeMetadata bool            `json:"include_metadata"` // 是否包含元数据
}

// SearchFilters 搜索过滤条件
type SearchFilters struct {
	DocumentTypes []string          `json:"document_types,omitempty"` // 文档类型过滤
	Tags          []string          `json:"tags,omitempty"`           // 标签过滤
	DateRange     *DateRange        `json:"date_range,omitempty"`     // 日期范围
	Sources       []string          `json:"sources,omitempty"`        // 来源过滤
	Languages     []string          `json:"languages,omitempty"`      // 语言过滤
	Custom        map[string]string `json:"custom,omitempty"`         // 自定义过滤
}

// DateRange 日期范围
type DateRange struct {
	Start time.Time `json:"start"`
	End   time.Time `json:"end"`
}

// SearchType 搜索类型
type SearchType string

const (
	SearchTypeSemantic SearchType = "semantic" // 语义搜索
	SearchTypeLexical  SearchType = "lexical"  // 词汇搜索
	SearchTypeHybrid   SearchType = "hybrid"   // 混合搜索
)

// SearchResults 搜索结果集合
type SearchResults struct {
	Results    []SearchResult `json:"results"`
	Total      int            `json:"total"`
	Query      SearchQuery    `json:"query"`
	Duration   time.Duration  `json:"duration"`
	SearchedAt time.Time      `json:"searched_at"`
}

// NewSearchQuery 创建搜索查询
func NewSearchQuery(query, knowledgeBaseID string) *SearchQuery {
	return &SearchQuery{
		Query:           query,
		KnowledgeBaseID: knowledgeBaseID,
		TopK:            10,
		ScoreThreshold:  0.0,
		SearchType:      SearchTypeSemantic,
		Rerank:          false,
		IncludeMetadata: true,
		Filters:         SearchFilters{},
	}
}

// WithTopK 设置返回结果数量
func (sq *SearchQuery) WithTopK(topK int) *SearchQuery {
	sq.TopK = topK
	return sq
}

// WithScoreThreshold 设置分数阈值
func (sq *SearchQuery) WithScoreThreshold(threshold float32) *SearchQuery {
	sq.ScoreThreshold = threshold
	return sq
}

// WithSearchType 设置搜索类型
func (sq *SearchQuery) WithSearchType(searchType SearchType) *SearchQuery {
	sq.SearchType = searchType
	return sq
}

// WithFilters 设置过滤条件
func (sq *SearchQuery) WithFilters(filters SearchFilters) *SearchQuery {
	sq.Filters = filters
	return sq
}

// AddDocumentTypeFilter 添加文档类型过滤
func (sq *SearchQuery) AddDocumentTypeFilter(docType string) *SearchQuery {
	sq.Filters.DocumentTypes = append(sq.Filters.DocumentTypes, docType)
	return sq
}

// AddTagFilter 添加标签过滤
func (sq *SearchQuery) AddTagFilter(tag string) *SearchQuery {
	sq.Filters.Tags = append(sq.Filters.Tags, tag)
	return sq
}

// SetDateRange 设置日期范围
func (sq *SearchQuery) SetDateRange(start, end time.Time) *SearchQuery {
	sq.Filters.DateRange = &DateRange{
		Start: start,
		End:   end,
	}
	return sq
}

// NewSearchResult 创建搜索结果
func NewSearchResult(id, content, title, source string, score float32, resultType SearchResultType) *SearchResult {
	return &SearchResult{
		ID:         id,
		Type:       resultType,
		Score:      score,
		Content:    content,
		Title:      title,
		Source:     source,
		Metadata:   make(map[string]string),
		SearchedAt: time.Now(),
	}
}

// SetChunkInfo 设置分块信息
func (sr *SearchResult) SetChunkInfo(info *ChunkInfo) {
	sr.ChunkInfo = info
	sr.Type = SearchResultTypeChunk
}

// SetDocumentInfo 设置文档信息
func (sr *SearchResult) SetDocumentInfo(info *DocumentInfo) {
	sr.DocumentInfo = info
	if sr.Type == "" {
		sr.Type = SearchResultTypeDocument
	}
}

// AddMetadata 添加元数据
func (sr *SearchResult) AddMetadata(key, value string) {
	if sr.Metadata == nil {
		sr.Metadata = make(map[string]string)
	}
	sr.Metadata[key] = value
}

// SetHighlight 设置高亮片段
func (sr *SearchResult) SetHighlight(highlight string) {
	sr.Highlight = highlight
}

// IsRelevant 检查结果是否相关
func (sr *SearchResult) IsRelevant(threshold float32) bool {
	return sr.Score >= threshold
}

// GetPreview 获取内容预览
func (sr *SearchResult) GetPreview(length int) string {
	if length <= 0 || length >= len(sr.Content) {
		return sr.Content
	}
	return sr.Content[:length] + "..."
}

// NewSearchResults 创建搜索结果集合
func NewSearchResults(query SearchQuery) *SearchResults {
	return &SearchResults{
		Results:    make([]SearchResult, 0),
		Total:      0,
		Query:      query,
		SearchedAt: time.Now(),
	}
}

// AddResult 添加搜索结果
func (srs *SearchResults) AddResult(result SearchResult) {
	srs.Results = append(srs.Results, result)
	srs.Total = len(srs.Results)
}

// FilterByScore 按分数过滤结果
func (srs *SearchResults) FilterByScore(threshold float32) {
	filtered := make([]SearchResult, 0)
	for _, result := range srs.Results {
		if result.Score >= threshold {
			filtered = append(filtered, result)
		}
	}
	srs.Results = filtered
	srs.Total = len(srs.Results)
}

// SortByScore 按分数排序（降序）
func (srs *SearchResults) SortByScore() {
	// TODO: 实现排序逻辑
}

// HasResults 检查是否有结果
func (srs *SearchResults) HasResults() bool {
	return len(srs.Results) > 0
}

// GetTopResults 获取前N个结果
func (srs *SearchResults) GetTopResults(n int) []SearchResult {
	if n >= len(srs.Results) {
		return srs.Results
	}
	return srs.Results[:n]
}
