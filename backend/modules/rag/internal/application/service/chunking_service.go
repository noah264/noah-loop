package service

import (
	"context"
	"fmt"
	"strings"

	"github.com/noah-loop/backend/modules/rag/internal/domain"
)

// ChunkingService 分块服务接口
type ChunkingService interface {
	// ChunkDocument 对文档进行分块
	ChunkDocument(ctx context.Context, document *domain.Document) ([]*domain.Chunk, error)
	
	// ChunkText 对文本进行分块
	ChunkText(ctx context.Context, text string, chunkType domain.ChunkType) ([]*domain.Chunk, error)
	
	// GetOptimalChunkSize 获取最佳分块大小
	GetOptimalChunkSize(text string, maxTokens int) int
	
	// ValidateChunk 验证分块
	ValidateChunk(chunk *domain.Chunk) error
}

// ChunkingStrategy 分块策略
type ChunkingStrategy string

const (
	ChunkingStrategyFixedSize    ChunkingStrategy = "fixed_size"    // 固定大小
	ChunkingStrategySemantic     ChunkingStrategy = "semantic"      // 语义分块
	ChunkingStrategyStructural   ChunkingStrategy = "structural"    // 结构化分块
	ChunkingStrategyHybrid       ChunkingStrategy = "hybrid"        // 混合分块
)

// ChunkingConfig 分块配置
type ChunkingConfig struct {
	Strategy      ChunkingStrategy `json:"strategy"`
	ChunkSize     int             `json:"chunk_size"`     // 分块大小（字符数）
	ChunkOverlap  int             `json:"chunk_overlap"`  // 重叠大小
	MinChunkSize  int             `json:"min_chunk_size"` // 最小分块大小
	MaxChunkSize  int             `json:"max_chunk_size"` // 最大分块大小
	Separators    []string        `json:"separators"`     // 分隔符
	KeepSeparator bool            `json:"keep_separator"` // 保留分隔符
}

// DefaultChunkingConfig 默认分块配置
func DefaultChunkingConfig() *ChunkingConfig {
	return &ChunkingConfig{
		Strategy:      ChunkingStrategyFixedSize,
		ChunkSize:     1000,
		ChunkOverlap:  200,
		MinChunkSize:  100,
		MaxChunkSize:  2000,
		Separators:    []string{"\n\n", "\n", "。", "！", "？", ".", "!", "?"},
		KeepSeparator: false,
	}
}

// Validate 验证配置
func (c *ChunkingConfig) Validate() error {
	if c.ChunkSize <= 0 {
		return fmt.Errorf("chunk size must be positive")
	}
	
	if c.ChunkOverlap < 0 {
		return fmt.Errorf("chunk overlap cannot be negative")
	}
	
	if c.ChunkOverlap >= c.ChunkSize {
		return fmt.Errorf("chunk overlap must be less than chunk size")
	}
	
	if c.MinChunkSize <= 0 {
		return fmt.Errorf("min chunk size must be positive")
	}
	
	if c.MaxChunkSize <= c.MinChunkSize {
		return fmt.Errorf("max chunk size must be greater than min chunk size")
	}
	
	return nil
}

// DefaultChunkingService 默认分块服务实现
type DefaultChunkingService struct {
	config *ChunkingConfig
}

// NewDefaultChunkingService 创建默认分块服务
func NewDefaultChunkingService(config *ChunkingConfig) *DefaultChunkingService {
	if config == nil {
		config = DefaultChunkingConfig()
	}
	
	return &DefaultChunkingService{
		config: config,
	}
}

// ChunkDocument 对文档进行分块
func (s *DefaultChunkingService) ChunkDocument(ctx context.Context, document *domain.Document) ([]*domain.Chunk, error) {
	if document == nil {
		return nil, fmt.Errorf("document cannot be nil")
	}
	
	if document.Content == "" {
		return nil, fmt.Errorf("document content cannot be empty")
	}
	
	// 根据文档类型选择分块策略
	chunkType := s.getChunkTypeForDocument(document.Type)
	
	// 预处理文档内容
	content := s.preprocessContent(document.Content, document.Type)
	
	// 执行分块
	textChunks := s.splitText(content)
	
	// 创建分块对象
	chunks := make([]*domain.Chunk, 0, len(textChunks))
	for i, textChunk := range textChunks {
		chunk, err := domain.NewChunk(document.ID, textChunk.Content, chunkType, i)
		if err != nil {
			return nil, fmt.Errorf("failed to create chunk %d: %w", i, err)
		}
		
		// 设置分块位置信息
		chunk.StartIndex = textChunk.StartIndex
		chunk.EndIndex = textChunk.EndIndex
		
		// 设置元数据
		chunk.Metadata.Title = document.Title
		if document.Metadata.Author != "" {
			chunk.Metadata.Custom["author"] = document.Metadata.Author
		}
		if document.Source != "" {
			chunk.Metadata.Custom["source"] = document.Source
		}
		
		chunks = append(chunks, chunk)
	}
	
	return chunks, nil
}

// ChunkText 对文本进行分块
func (s *DefaultChunkingService) ChunkText(ctx context.Context, text string, chunkType domain.ChunkType) ([]*domain.Chunk, error) {
	if text == "" {
		return nil, fmt.Errorf("text cannot be empty")
	}
	
	textChunks := s.splitText(text)
	
	chunks := make([]*domain.Chunk, 0, len(textChunks))
	for i, textChunk := range textChunks {
		chunk, err := domain.NewChunk("", textChunk.Content, chunkType, i)
		if err != nil {
			return nil, fmt.Errorf("failed to create chunk %d: %w", i, err)
		}
		
		chunk.StartIndex = textChunk.StartIndex
		chunk.EndIndex = textChunk.EndIndex
		
		chunks = append(chunks, chunk)
	}
	
	return chunks, nil
}

// GetOptimalChunkSize 获取最佳分块大小
func (s *DefaultChunkingService) GetOptimalChunkSize(text string, maxTokens int) int {
	// 简单实现：基于文本长度和最大令牌数计算
	textLength := len(text)
	estimatedTokens := textLength / 4 // 简单估算
	
	if estimatedTokens <= maxTokens {
		return textLength
	}
	
	return (maxTokens * 4) // 转换回字符数
}

// ValidateChunk 验证分块
func (s *DefaultChunkingService) ValidateChunk(chunk *domain.Chunk) error {
	if chunk == nil {
		return fmt.Errorf("chunk cannot be nil")
	}
	
	if chunk.Content == "" {
		return fmt.Errorf("chunk content cannot be empty")
	}
	
	contentLength := len(chunk.Content)
	if contentLength < s.config.MinChunkSize {
		return fmt.Errorf("chunk size %d is below minimum %d", contentLength, s.config.MinChunkSize)
	}
	
	if contentLength > s.config.MaxChunkSize {
		return fmt.Errorf("chunk size %d exceeds maximum %d", contentLength, s.config.MaxChunkSize)
	}
	
	return nil
}

// TextChunk 文本分块结构
type TextChunk struct {
	Content    string
	StartIndex int
	EndIndex   int
}

// splitText 分割文本
func (s *DefaultChunkingService) splitText(text string) []TextChunk {
	switch s.config.Strategy {
	case ChunkingStrategyFixedSize:
		return s.fixedSizeSplit(text)
	case ChunkingStrategySemantic:
		return s.semanticSplit(text)
	case ChunkingStrategyStructural:
		return s.structuralSplit(text)
	default:
		return s.fixedSizeSplit(text)
	}
}

// fixedSizeSplit 固定大小分割
func (s *DefaultChunkingService) fixedSizeSplit(text string) []TextChunk {
	var chunks []TextChunk
	textLen := len(text)
	
	if textLen <= s.config.ChunkSize {
		return []TextChunk{{
			Content:    text,
			StartIndex: 0,
			EndIndex:   textLen,
		}}
	}
	
	start := 0
	for start < textLen {
		end := start + s.config.ChunkSize
		if end > textLen {
			end = textLen
		}
		
		// 尝试在分隔符处分割
		actualEnd := s.findBestSplitPoint(text, start, end)
		
		chunk := TextChunk{
			Content:    text[start:actualEnd],
			StartIndex: start,
			EndIndex:   actualEnd,
		}
		chunks = append(chunks, chunk)
		
		// 计算下一个开始位置（考虑重叠）
		start = actualEnd - s.config.ChunkOverlap
		if start < 0 {
			start = 0
		}
		
		// 如果没有进展，强制移动
		if start == chunk.StartIndex {
			start = actualEnd
		}
	}
	
	return chunks
}

// semanticSplit 语义分割（简单实现）
func (s *DefaultChunkingService) semanticSplit(text string) []TextChunk {
	// 简单实现：按段落分割
	paragraphs := strings.Split(text, "\n\n")
	var chunks []TextChunk
	currentChunk := ""
	startIndex := 0
	
	for _, paragraph := range paragraphs {
		if len(currentChunk)+len(paragraph) > s.config.ChunkSize && currentChunk != "" {
			// 创建当前分块
			chunks = append(chunks, TextChunk{
				Content:    strings.TrimSpace(currentChunk),
				StartIndex: startIndex,
				EndIndex:   startIndex + len(currentChunk),
			})
			
			// 开始新分块
			startIndex += len(currentChunk)
			currentChunk = paragraph
		} else {
			if currentChunk != "" {
				currentChunk += "\n\n"
			}
			currentChunk += paragraph
		}
	}
	
	// 添加最后一个分块
	if currentChunk != "" {
		chunks = append(chunks, TextChunk{
			Content:    strings.TrimSpace(currentChunk),
			StartIndex: startIndex,
			EndIndex:   startIndex + len(currentChunk),
		})
	}
	
	return chunks
}

// structuralSplit 结构化分割（简单实现）
func (s *DefaultChunkingService) structuralSplit(text string) []TextChunk {
	// 简单实现：按标题和段落分割
	// TODO: 实现更复杂的结构化分割逻辑
	return s.semanticSplit(text)
}

// findBestSplitPoint 找到最佳分割点
func (s *DefaultChunkingService) findBestSplitPoint(text string, start, maxEnd int) int {
	if maxEnd >= len(text) {
		return len(text)
	}
	
	// 在分隔符附近寻找最佳分割点
	searchStart := maxEnd - 100
	if searchStart < start {
		searchStart = start
	}
	
	for _, separator := range s.config.Separators {
		for i := maxEnd - 1; i >= searchStart; i-- {
			if i+len(separator) <= len(text) && text[i:i+len(separator)] == separator {
				if s.config.KeepSeparator {
					return i + len(separator)
				}
				return i
			}
		}
	}
	
	// 如果找不到分隔符，返回原始结束位置
	return maxEnd
}

// preprocessContent 预处理内容
func (s *DefaultChunkingService) preprocessContent(content string, docType domain.DocumentType) string {
	// 根据文档类型进行预处理
	switch docType {
	case domain.DocumentTypeHTML:
		return s.preprocessHTML(content)
	case domain.DocumentTypeMarkdown:
		return s.preprocessMarkdown(content)
	default:
		return s.preprocessText(content)
	}
}

// preprocessHTML 预处理HTML内容
func (s *DefaultChunkingService) preprocessHTML(content string) string {
	// TODO: 实现HTML预处理（移除标签、提取纯文本等）
	return content
}

// preprocessMarkdown 预处理Markdown内容
func (s *DefaultChunkingService) preprocessMarkdown(content string) string {
	// TODO: 实现Markdown预处理（保留结构信息等）
	return content
}

// preprocessText 预处理纯文本内容
func (s *DefaultChunkingService) preprocessText(content string) string {
	// 标准化换行符
	content = strings.ReplaceAll(content, "\r\n", "\n")
	content = strings.ReplaceAll(content, "\r", "\n")
	
	// 移除多余的空行
	lines := strings.Split(content, "\n")
	var cleanLines []string
	prevEmpty := false
	
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			if !prevEmpty {
				cleanLines = append(cleanLines, "")
				prevEmpty = true
			}
		} else {
			cleanLines = append(cleanLines, line)
			prevEmpty = false
		}
	}
	
	return strings.Join(cleanLines, "\n")
}

// getChunkTypeForDocument 根据文档类型获取分块类型
func (s *DefaultChunkingService) getChunkTypeForDocument(docType domain.DocumentType) domain.ChunkType {
	switch docType {
	case domain.DocumentTypeMarkdown:
		return domain.ChunkTypeSection
	case domain.DocumentTypeHTML:
		return domain.ChunkTypeSection
	case domain.DocumentTypePDF:
		return domain.ChunkTypeParagraph
	default:
		return domain.ChunkTypeText
	}
}
