package embedding

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/noah-loop/backend/modules/rag/internal/application/service"
	"github.com/noah-loop/backend/shared/pkg/infrastructure"
)

// OpenAIEmbeddingService OpenAI嵌入服务实现
type OpenAIEmbeddingService struct {
	config     *service.EmbeddingConfig
	httpClient *http.Client
	logger     infrastructure.Logger
	metrics    *service.EmbeddingMetrics
}

// NewOpenAIEmbeddingService 创建OpenAI嵌入服务
func NewOpenAIEmbeddingService(config *service.EmbeddingConfig, logger infrastructure.Logger) service.EmbeddingService {
	if config == nil {
		config = service.DefaultEmbeddingConfig()
	}
	
	httpClient := &http.Client{
		Timeout: time.Duration(config.Timeout) * time.Second,
	}
	
	return &OpenAIEmbeddingService{
		config:     config,
		httpClient: httpClient,
		logger:     logger,
		metrics: &service.EmbeddingMetrics{
			TotalRequests:  0,
			TotalTokens:    0,
			AverageLatency: 0,
			SuccessRate:    1.0,
			ErrorCount:     0,
		},
	}
}

// GenerateEmbedding 生成单个文本的嵌入向量
func (s *OpenAIEmbeddingService) GenerateEmbedding(ctx context.Context, text string) ([]float32, error) {
	if text == "" {
		return nil, fmt.Errorf("text cannot be empty")
	}
	
	start := time.Now()
	s.metrics.TotalRequests++
	
	// 构建请求
	reqBody := map[string]interface{}{
		"input": text,
		"model": s.config.Model,
	}
	
	embedding, tokenCount, err := s.makeRequest(ctx, reqBody)
	
	duration := time.Since(start)
	s.updateMetrics(duration, tokenCount, err == nil)
	
	if err != nil {
		return nil, err
	}
	
	if len(embedding) == 0 {
		return nil, fmt.Errorf("received empty embedding")
	}
	
	return embedding[0], nil
}

// GenerateEmbeddings 批量生成嵌入向量
func (s *OpenAIEmbeddingService) GenerateEmbeddings(ctx context.Context, texts []string) ([][]float32, error) {
	if len(texts) == 0 {
		return nil, fmt.Errorf("texts cannot be empty")
	}
	
	// 分批处理
	batchSize := s.config.BatchSize
	if batchSize <= 0 {
		batchSize = 100
	}
	
	var allEmbeddings [][]float32
	
	for i := 0; i < len(texts); i += batchSize {
		end := i + batchSize
		if end > len(texts) {
			end = len(texts)
		}
		
		batch := texts[i:end]
		batchEmbeddings, err := s.generateBatchEmbeddings(ctx, batch)
		if err != nil {
			return nil, fmt.Errorf("failed to generate batch embeddings: %w", err)
		}
		
		allEmbeddings = append(allEmbeddings, batchEmbeddings...)
	}
	
	return allEmbeddings, nil
}

// generateBatchEmbeddings 生成批量嵌入
func (s *OpenAIEmbeddingService) generateBatchEmbeddings(ctx context.Context, texts []string) ([][]float32, error) {
	start := time.Now()
	s.metrics.TotalRequests++
	
	// 构建请求
	reqBody := map[string]interface{}{
		"input": texts,
		"model": s.config.Model,
	}
	
	embeddings, tokenCount, err := s.makeRequest(ctx, reqBody)
	
	duration := time.Since(start)
	s.updateMetrics(duration, int64(tokenCount), err == nil)
	
	if err != nil {
		return nil, err
	}
	
	if len(embeddings) != len(texts) {
		return nil, fmt.Errorf("embedding count mismatch: expected %d, got %d", len(texts), len(embeddings))
	}
	
	return embeddings, nil
}

// makeRequest 发起HTTP请求
func (s *OpenAIEmbeddingService) makeRequest(ctx context.Context, reqBody map[string]interface{}) ([][]float32, int, error) {
	// 序列化请求体
	jsonBody, err := json.Marshal(reqBody)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to marshal request: %w", err)
	}
	
	// 创建HTTP请求
	apiURL := s.config.APIBase
	if apiURL == "" {
		apiURL = "https://api.openai.com"
	}
	apiURL += "/v1/embeddings"
	
	req, err := http.NewRequestWithContext(ctx, "POST", apiURL, bytes.NewBuffer(jsonBody))
	if err != nil {
		return nil, 0, fmt.Errorf("failed to create request: %w", err)
	}
	
	// 设置请求头
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+s.config.APIKey)
	
	// 发送请求
	resp, err := s.httpClient.Do(req)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()
	
	// 读取响应
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to read response: %w", err)
	}
	
	// 检查HTTP状态码
	if resp.StatusCode != http.StatusOK {
		return nil, 0, fmt.Errorf("API request failed with status %d: %s", resp.StatusCode, string(respBody))
	}
	
	// 解析响应
	var apiResp OpenAIEmbeddingResponse
	err = json.Unmarshal(respBody, &apiResp)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to unmarshal response: %w", err)
	}
	
	// 提取嵌入向量
	embeddings := make([][]float32, len(apiResp.Data))
	for i, item := range apiResp.Data {
		embedding := make([]float32, len(item.Embedding))
		for j, val := range item.Embedding {
			embedding[j] = float32(val)
		}
		embeddings[i] = embedding
	}
	
	return embeddings, apiResp.Usage.TotalTokens, nil
}

// GetDimension 获取向量维度
func (s *OpenAIEmbeddingService) GetDimension() int {
	return s.config.Dimension
}

// GetModel 获取模型名称
func (s *OpenAIEmbeddingService) GetModel() string {
	return s.config.Model
}

// ValidateEmbedding 验证嵌入向量
func (s *OpenAIEmbeddingService) ValidateEmbedding(embedding []float32) error {
	if len(embedding) == 0 {
		return fmt.Errorf("embedding cannot be empty")
	}
	
	if len(embedding) != s.config.Dimension {
		return fmt.Errorf("embedding dimension mismatch: expected %d, got %d", s.config.Dimension, len(embedding))
	}
	
	return nil
}

// updateMetrics 更新指标
func (s *OpenAIEmbeddingService) updateMetrics(duration time.Duration, tokenCount int64, success bool) {
	s.metrics.TotalTokens += tokenCount
	
	// 更新平均延迟
	if s.metrics.TotalRequests == 1 {
		s.metrics.AverageLatency = float64(duration.Milliseconds())
	} else {
		s.metrics.AverageLatency = (s.metrics.AverageLatency*float64(s.metrics.TotalRequests-1) + float64(duration.Milliseconds())) / float64(s.metrics.TotalRequests)
	}
	
	// 更新成功率
	if !success {
		s.metrics.ErrorCount++
	}
	s.metrics.SuccessRate = float64(s.metrics.TotalRequests-s.metrics.ErrorCount) / float64(s.metrics.TotalRequests)
	
	s.metrics.LastRequestAt = time.Now().Format(time.RFC3339)
}

// OpenAI API 响应结构
type OpenAIEmbeddingResponse struct {
	Object string `json:"object"`
	Data   []struct {
		Object    string    `json:"object"`
		Index     int       `json:"index"`
		Embedding []float64 `json:"embedding"`
	} `json:"data"`
	Model string `json:"model"`
	Usage struct {
		PromptTokens int `json:"prompt_tokens"`
		TotalTokens  int `json:"total_tokens"`
	} `json:"usage"`
}
