package providers

import (
	"context"
	"errors"
	"fmt"
	
	"github.com/noah-loop/backend/modules/llm/internal/application/service"
	"github.com/noah-loop/backend/modules/llm/internal/domain"
	"github.com/noah-loop/backend/shared/pkg/infrastructure"
	openai "github.com/sashabaranov/go-openai"
	"go.uber.org/zap"
)

// OpenAIProvider OpenAI提供商实现
type OpenAIProvider struct {
	client *openai.Client
	logger infrastructure.Logger
}

// NewOpenAIProvider 创建OpenAI提供商
func NewOpenAIProvider(apiKey string, logger infrastructure.Logger) *OpenAIProvider {
	config := openai.DefaultConfig(apiKey)
	client := openai.NewClientWithConfig(config)
	
	return &OpenAIProvider{
		client: client,
		logger: logger,
	}
}

// Process 处理请求
func (p *OpenAIProvider) Process(ctx context.Context, request *service.ProviderRequest) (*service.ProviderResponse, error) {
	switch request.Model.Type {
	case domain.ModelTypeChat:
		return p.processChat(ctx, request)
	case domain.ModelTypeCompletion:
		return p.processCompletion(ctx, request)
	case domain.ModelTypeEmbedding:
		return p.processEmbedding(ctx, request)
	default:
		return nil, fmt.Errorf("unsupported model type: %s", request.Model.Type)
	}
}

// processChat 处理聊天请求
func (p *OpenAIProvider) processChat(ctx context.Context, request *service.ProviderRequest) (*service.ProviderResponse, error) {
	// 解析消息
	messagesData, ok := request.Input["messages"]
	if !ok {
		return nil, errors.New("messages field is required")
	}
	
	messagesSlice, ok := messagesData.([]interface{})
	if !ok {
		return nil, errors.New("messages must be an array")
	}
	
	var messages []openai.ChatCompletionMessage
	for _, msgData := range messagesSlice {
		msgMap, ok := msgData.(map[string]interface{})
		if !ok {
			return nil, errors.New("invalid message format")
		}
		
		role, ok := msgMap["role"].(string)
		if !ok {
			return nil, errors.New("message role is required")
		}
		
		content, ok := msgMap["content"].(string)
		if !ok {
			return nil, errors.New("message content is required")
		}
		
		messages = append(messages, openai.ChatCompletionMessage{
			Role:    role,
			Content: content,
		})
	}
	
	// 构建请求
	req := openai.ChatCompletionRequest{
		Model:    request.Model.Name,
		Messages: messages,
	}
	
	// 应用配置
	if maxTokens, ok := request.Config["max_tokens"].(float64); ok {
		req.MaxTokens = int(maxTokens)
	}
	if temperature, ok := request.Config["temperature"].(float64); ok {
		req.Temperature = float32(temperature)
	}
	if topP, ok := request.Config["top_p"].(float64); ok {
		req.TopP = float32(topP)
	}
	
	// 调用API
	resp, err := p.client.CreateChatCompletion(ctx, req)
	if err != nil {
		p.logger.Error("OpenAI chat completion failed", zap.Error(err))
		return nil, err
	}
	
	// 构建响应
	output := map[string]interface{}{
		"choices": make([]map[string]interface{}, len(resp.Choices)),
	}
	
	for i, choice := range resp.Choices {
		output["choices"].([]map[string]interface{})[i] = map[string]interface{}{
			"message": map[string]interface{}{
				"role":    choice.Message.Role,
				"content": choice.Message.Content,
			},
			"finish_reason": choice.FinishReason,
		}
	}
	
	return &service.ProviderResponse{
		Output:     output,
		TokensUsed: resp.Usage.TotalTokens,
		Metadata: map[string]interface{}{
			"model":              resp.Model,
			"prompt_tokens":      resp.Usage.PromptTokens,
			"completion_tokens":  resp.Usage.CompletionTokens,
		},
	}, nil
}

// processCompletion 处理文本补全请求
func (p *OpenAIProvider) processCompletion(ctx context.Context, request *service.ProviderRequest) (*service.ProviderResponse, error) {
	prompt, ok := request.Input["prompt"].(string)
	if !ok {
		return nil, errors.New("prompt field is required")
	}
	
	req := openai.CompletionRequest{
		Model:  request.Model.Name,
		Prompt: prompt,
	}
	
	// 应用配置
	if maxTokens, ok := request.Config["max_tokens"].(float64); ok {
		req.MaxTokens = int(maxTokens)
	}
	if temperature, ok := request.Config["temperature"].(float64); ok {
		req.Temperature = float32(temperature)
	}
	
	resp, err := p.client.CreateCompletion(ctx, req)
	if err != nil {
		p.logger.Error("OpenAI completion failed", zap.Error(err))
		return nil, err
	}
	
	output := map[string]interface{}{
		"choices": make([]map[string]interface{}, len(resp.Choices)),
	}
	
	for i, choice := range resp.Choices {
		output["choices"].([]map[string]interface{})[i] = map[string]interface{}{
			"text":          choice.Text,
			"finish_reason": choice.FinishReason,
		}
	}
	
	return &service.ProviderResponse{
		Output:     output,
		TokensUsed: resp.Usage.TotalTokens,
		Metadata: map[string]interface{}{
			"model":              resp.Model,
			"prompt_tokens":      resp.Usage.PromptTokens,
			"completion_tokens":  resp.Usage.CompletionTokens,
		},
	}, nil
}

// processEmbedding 处理嵌入请求
func (p *OpenAIProvider) processEmbedding(ctx context.Context, request *service.ProviderRequest) (*service.ProviderResponse, error) {
	text, ok := request.Input["text"].(string)
	if !ok {
		return nil, errors.New("text field is required")
	}
	
	req := openai.EmbeddingRequest{
		Model: openai.EmbeddingModel(request.Model.Name),
		Input: text,
	}
	
	resp, err := p.client.CreateEmbeddings(ctx, req)
	if err != nil {
		p.logger.Error("OpenAI embedding failed", zap.Error(err))
		return nil, err
	}
	
	output := map[string]interface{}{
		"embeddings": make([]map[string]interface{}, len(resp.Data)),
	}
	
	for i, embedding := range resp.Data {
		output["embeddings"].([]map[string]interface{})[i] = map[string]interface{}{
			"embedding": embedding.Embedding,
		}
	}
	
	return &service.ProviderResponse{
		Output:     output,
		TokensUsed: resp.Usage.TotalTokens,
		Metadata: map[string]interface{}{
			"model":         resp.Model,
			"prompt_tokens": resp.Usage.PromptTokens,
		},
	}, nil
}

// Health 健康检查
func (p *OpenAIProvider) Health(ctx context.Context) error {
	// 简单的健康检查，可以尝试调用模型列表API
	_, err := p.client.ListModels(ctx)
	return err
}
