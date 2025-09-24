package service

import (
	"context"
	"fmt"
	"time"

	agentv1 "github.com/noah-loop/backend/api-gateway/proto/agent/v1"
	llmv1 "github.com/noah-loop/backend/api-gateway/proto/llm/v1"
	grpcClients "github.com/noah-loop/backend/api-gateway/internal/infrastructure/grpc"
	"github.com/noah-loop/backend/shared/pkg/infrastructure"
	"go.uber.org/zap"
)

// GRPCProxyService gRPC代理服务
type GRPCProxyService struct {
	clientManager *grpcClients.ClientManager
	logger        infrastructure.Logger
}

// NewGRPCProxyService 创建gRPC代理服务
func NewGRPCProxyService(clientManager *grpcClients.ClientManager, logger infrastructure.Logger) *GRPCProxyService {
	return &GRPCProxyService{
		clientManager: clientManager,
		logger:        logger,
	}
}

// Agent相关代理方法

// CreateAgent 创建智能体
func (s *GRPCProxyService) CreateAgent(ctx context.Context, req *agentv1.CreateAgentRequest) (*agentv1.CreateAgentResponse, error) {
	start := time.Now()
	defer func() {
		s.logger.Debug("gRPC proxy call completed",
			zap.String("method", "CreateAgent"),
			zap.Duration("duration", time.Since(start)))
	}()
	
	return s.clientManager.AgentClient.CreateAgent(ctx, req)
}

// GetAgent 获取智能体
func (s *GRPCProxyService) GetAgent(ctx context.Context, req *agentv1.GetAgentRequest) (*agentv1.GetAgentResponse, error) {
	start := time.Now()
	defer func() {
		s.logger.Debug("gRPC proxy call completed",
			zap.String("method", "GetAgent"),
			zap.Duration("duration", time.Since(start)))
	}()
	
	return s.clientManager.AgentClient.GetAgent(ctx, req)
}

// ListAgents 获取智能体列表
func (s *GRPCProxyService) ListAgents(ctx context.Context, req *agentv1.ListAgentsRequest) (*agentv1.ListAgentsResponse, error) {
	start := time.Now()
	defer func() {
		s.logger.Debug("gRPC proxy call completed",
			zap.String("method", "ListAgents"),
			zap.Duration("duration", time.Since(start)))
	}()
	
	return s.clientManager.AgentClient.ListAgents(ctx, req)
}

// UpdateAgent 更新智能体
func (s *GRPCProxyService) UpdateAgent(ctx context.Context, req *agentv1.UpdateAgentRequest) (*agentv1.UpdateAgentResponse, error) {
	return s.clientManager.AgentClient.UpdateAgent(ctx, req)
}

// DeleteAgent 删除智能体
func (s *GRPCProxyService) DeleteAgent(ctx context.Context, req *agentv1.DeleteAgentRequest) (*agentv1.DeleteAgentResponse, error) {
	return s.clientManager.AgentClient.DeleteAgent(ctx, req)
}

// Chat 与智能体对话
func (s *GRPCProxyService) Chat(ctx context.Context, req *agentv1.ChatRequest) (*agentv1.ChatResponse, error) {
	return s.clientManager.AgentClient.Chat(ctx, req)
}

// StreamChat 流式对话
func (s *GRPCProxyService) StreamChat(ctx context.Context, req *agentv1.ChatRequest) (agentv1.AgentService_StreamChatClient, error) {
	return s.clientManager.AgentClient.StreamChat(ctx, req)
}

// ExecuteTool 执行工具
func (s *GRPCProxyService) ExecuteTool(ctx context.Context, req *agentv1.ExecuteToolRequest) (*agentv1.ExecuteToolResponse, error) {
	return s.clientManager.AgentClient.ExecuteTool(ctx, req)
}

// LLM相关代理方法

// CreateModel 创建模型
func (s *GRPCProxyService) CreateModel(ctx context.Context, req *llmv1.CreateModelRequest) (*llmv1.CreateModelResponse, error) {
	start := time.Now()
	defer func() {
		s.logger.Debug("gRPC proxy call completed",
			zap.String("method", "CreateModel"),
			zap.Duration("duration", time.Since(start)))
	}()
	
	return s.clientManager.LLMClient.CreateModel(ctx, req)
}

// GetModel 获取模型
func (s *GRPCProxyService) GetModel(ctx context.Context, req *llmv1.GetModelRequest) (*llmv1.GetModelResponse, error) {
	return s.clientManager.LLMClient.GetModel(ctx, req)
}

// ListModels 获取模型列表
func (s *GRPCProxyService) ListModels(ctx context.Context, req *llmv1.ListModelsRequest) (*llmv1.ListModelsResponse, error) {
	return s.clientManager.LLMClient.ListModels(ctx, req)
}

// ProcessRequest 处理请求
func (s *GRPCProxyService) ProcessRequest(ctx context.Context, req *llmv1.ProcessRequestRequest) (*llmv1.ProcessRequestResponse, error) {
	return s.clientManager.LLMClient.ProcessRequest(ctx, req)
}

// StreamProcessRequest 流式处理请求
func (s *GRPCProxyService) StreamProcessRequest(ctx context.Context, req *llmv1.ProcessRequestRequest) (llmv1.LLMService_StreamProcessRequestClient, error) {
	return s.clientManager.LLMClient.StreamProcessRequest(ctx, req)
}

// GetUsageStats 获取使用统计
func (s *GRPCProxyService) GetUsageStats(ctx context.Context, req *llmv1.GetUsageStatsRequest) (*llmv1.GetUsageStatsResponse, error) {
	return s.clientManager.LLMClient.GetUsageStats(ctx, req)
}

// HealthCheckServices 检查所有服务健康状态
func (s *GRPCProxyService) HealthCheckServices(ctx context.Context) (map[string]bool, error) {
	results := s.clientManager.HealthCheck(ctx)
	
	// 记录健康检查结果
	for service, healthy := range results {
		if healthy {
			s.logger.Debug("Service health check passed", zap.String("service", service))
		} else {
			s.logger.Warn("Service health check failed", zap.String("service", service))
		}
	}
	
	return results, nil
}
