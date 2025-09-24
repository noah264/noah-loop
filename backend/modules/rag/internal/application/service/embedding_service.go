package service

import (
	"context"
	"fmt"
)

// EmbeddingService 嵌入向量服务接口
type EmbeddingService interface {
	// GenerateEmbedding 生成单个文本的嵌入向量
	GenerateEmbedding(ctx context.Context, text string) ([]float32, error)
	
	// GenerateEmbeddings 批量生成嵌入向量
	GenerateEmbeddings(ctx context.Context, texts []string) ([][]float32, error)
	
	// GetDimension 获取向量维度
	GetDimension() int
	
	// GetModel 获取模型名称
	GetModel() string
	
	// ValidateEmbedding 验证嵌入向量
	ValidateEmbedding(embedding []float32) error
}

// EmbeddingProvider 嵌入向量提供商
type EmbeddingProvider string

const (
	EmbeddingProviderOpenAI    EmbeddingProvider = "openai"
	EmbeddingProviderHuggingFace EmbeddingProvider = "huggingface"
	EmbeddingProviderLocal     EmbeddingProvider = "local"
	EmbeddingProviderAzure     EmbeddingProvider = "azure"
)

// EmbeddingConfig 嵌入配置
type EmbeddingConfig struct {
	Provider    EmbeddingProvider `json:"provider"`
	Model       string           `json:"model"`
	APIKey      string           `json:"api_key,omitempty"`
	APIBase     string           `json:"api_base,omitempty"`
	Dimension   int              `json:"dimension"`
	MaxTokens   int              `json:"max_tokens"`
	BatchSize   int              `json:"batch_size"`
	Timeout     int              `json:"timeout"` // 秒
}

// DefaultEmbeddingConfig 默认配置
func DefaultEmbeddingConfig() *EmbeddingConfig {
	return &EmbeddingConfig{
		Provider:   EmbeddingProviderOpenAI,
		Model:      "text-embedding-ada-002",
		Dimension:  1536,
		MaxTokens:  8191,
		BatchSize:  100,
		Timeout:    30,
	}
}

// Validate 验证配置
func (c *EmbeddingConfig) Validate() error {
	if c.Provider == "" {
		return fmt.Errorf("embedding provider is required")
	}
	
	if c.Model == "" {
		return fmt.Errorf("embedding model is required")
	}
	
	if c.Dimension <= 0 {
		return fmt.Errorf("embedding dimension must be positive")
	}
	
	if c.MaxTokens <= 0 {
		return fmt.Errorf("max tokens must be positive")
	}
	
	if c.BatchSize <= 0 {
		return fmt.Errorf("batch size must be positive")
	}
	
	return nil
}

// EmbeddingMetrics 嵌入指标
type EmbeddingMetrics struct {
	TotalRequests     int64   `json:"total_requests"`
	TotalTokens       int64   `json:"total_tokens"`
	AverageLatency    float64 `json:"average_latency"` // 毫秒
	SuccessRate       float64 `json:"success_rate"`
	ErrorCount        int64   `json:"error_count"`
	LastRequestAt     string  `json:"last_request_at"`
}

// EmbeddingResult 嵌入结果
type EmbeddingResult struct {
	Text       string    `json:"text"`
	Embedding  []float32 `json:"embedding"`
	TokenCount int       `json:"token_count"`
	Model      string    `json:"model"`
	Duration   int64     `json:"duration"` // 毫秒
}

// BatchEmbeddingResult 批量嵌入结果
type BatchEmbeddingResult struct {
	Results      []EmbeddingResult `json:"results"`
	TotalTokens  int               `json:"total_tokens"`
	TotalDuration int64            `json:"total_duration"` // 毫秒
	Model        string            `json:"model"`
	SuccessCount int               `json:"success_count"`
	ErrorCount   int               `json:"error_count"`
	Errors       []string          `json:"errors,omitempty"`
}

// TextPreprocessor 文本预处理器接口
type TextPreprocessor interface {
	// Preprocess 预处理文本
	Preprocess(text string) string
	
	// TokenCount 计算令牌数量
	TokenCount(text string) int
	
	// TruncateText 截断文本
	TruncateText(text string, maxTokens int) string
	
	// CleanText 清理文本
	CleanText(text string) string
}

// BasicTextPreprocessor 基础文本预处理器
type BasicTextPreprocessor struct {
	MaxTokens int
}

// NewBasicTextPreprocessor 创建基础文本预处理器
func NewBasicTextPreprocessor(maxTokens int) *BasicTextPreprocessor {
	return &BasicTextPreprocessor{
		MaxTokens: maxTokens,
	}
}

// Preprocess 预处理文本
func (p *BasicTextPreprocessor) Preprocess(text string) string {
	// 清理文本
	cleaned := p.CleanText(text)
	
	// 截断文本
	if p.TokenCount(cleaned) > p.MaxTokens {
		cleaned = p.TruncateText(cleaned, p.MaxTokens)
	}
	
	return cleaned
}

// TokenCount 计算令牌数量（简单实现）
func (p *BasicTextPreprocessor) TokenCount(text string) int {
	// 简单实现：按字符数除以4估算
	// 实际应用中应该使用更准确的tokenizer
	return len(text) / 4
}

// TruncateText 截断文本
func (p *BasicTextPreprocessor) TruncateText(text string, maxTokens int) string {
	maxChars := maxTokens * 4 // 简单估算
	if len(text) <= maxChars {
		return text
	}
	return text[:maxChars]
}

// CleanText 清理文本
func (p *BasicTextPreprocessor) CleanText(text string) string {
	// TODO: 实现文本清理逻辑
	// 1. 移除多余的空白字符
	// 2. 标准化换行符
	// 3. 移除特殊字符
	// 4. 处理编码问题
	return text
}

// EmbeddingCache 嵌入缓存接口
type EmbeddingCache interface {
	// Get 获取缓存的嵌入
	Get(ctx context.Context, text string) ([]float32, bool)
	
	// Set 设置嵌入缓存
	Set(ctx context.Context, text string, embedding []float32) error
	
	// Delete 删除缓存
	Delete(ctx context.Context, text string) error
	
	// Clear 清空缓存
	Clear(ctx context.Context) error
	
	// Stats 获取缓存统计
	Stats(ctx context.Context) (*CacheStats, error)
}

// CacheStats 缓存统计
type CacheStats struct {
	HitCount  int64   `json:"hit_count"`
	MissCount int64   `json:"miss_count"`
	HitRate   float64 `json:"hit_rate"`
	Size      int64   `json:"size"`
	Memory    int64   `json:"memory"` // 字节
}
