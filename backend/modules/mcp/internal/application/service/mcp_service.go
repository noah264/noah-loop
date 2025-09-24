package service

import (
	"context"
	"fmt"
	"time"
	
	"github.com/google/uuid"
	"github.com/noah-loop/backend/modules/mcp/internal/domain"
	"github.com/noah-loop/backend/shared/pkg/application"
	"github.com/noah-loop/backend/shared/pkg/infrastructure"
	"go.uber.org/zap"
)

// MCPService MCP应用服务
type MCPService struct {
	sessionRepo domain.SessionRepository
	contextRepo domain.ContextRepository
	eventBus    application.EventBus
	logger      infrastructure.Logger
	metrics     *infrastructure.MetricsRegistry
	compressor  ContextCompressor
}

// NewMCPService 创建MCP服务
func NewMCPService(
	sessionRepo domain.SessionRepository,
	contextRepo domain.ContextRepository,
	eventBus application.EventBus,
	logger infrastructure.Logger,
	metrics *infrastructure.MetricsRegistry,
) *MCPService {
	return &MCPService{
		sessionRepo: sessionRepo,
		contextRepo: contextRepo,
		eventBus:    eventBus,
		logger:      logger,
		metrics:     metrics,
		compressor:  NewSimpleCompressor(),
	}
}

// SetCompressor 设置压缩器
func (s *MCPService) SetCompressor(compressor ContextCompressor) {
	s.compressor = compressor
}

// CreateSession 创建会话
func (s *MCPService) CreateSession(ctx context.Context, cmd *CreateSessionCommand) (*application.Result, error) {
	if err := cmd.Validate(); err != nil {
		return &application.Result{Success: false, Error: err.Error()}, err
	}
	
	// 创建会话
	session := domain.NewSession(cmd.UserID, cmd.AgentID, cmd.Title)
	session.Description = cmd.Description
	session.Metadata = cmd.Metadata
	if cmd.MaxContextSize > 0 {
		session.MaxContextSize = cmd.MaxContextSize
	}
	if cmd.ExpiresIn > 0 {
		expiresAt := time.Now().Add(cmd.ExpiresIn)
		session.ExpiresAt = &expiresAt
	}
	
	// 保存会话
	if err := s.sessionRepo.Save(ctx, session); err != nil {
		s.logger.Error("Failed to save session", zap.Error(err))
		return &application.Result{Success: false, Error: "failed to save session"}, err
	}
	
	// 发布事件
	for _, event := range session.GetDomainEvents() {
		if err := s.eventBus.Publish(ctx, event); err != nil {
			s.logger.Warn("Failed to publish event", zap.Error(err))
		}
	}
	session.ClearDomainEvents()
	
	// 更新会话指标
	if s.metrics != nil {
		go s.updateSessionMetrics()
	}
	
	return &application.Result{Success: true, Data: session}, nil
}

// AddContext 添加上下文
func (s *MCPService) AddContext(ctx context.Context, cmd *AddContextCommand) (*application.Result, error) {
	if err := cmd.Validate(); err != nil {
		return &application.Result{Success: false, Error: err.Error()}, err
	}
	
	// 获取会话
	session, err := s.sessionRepo.FindByID(ctx, cmd.SessionID)
	if err != nil {
		return &application.Result{Success: false, Error: "session not found"}, err
	}
	
	// 创建上下文
	context := domain.NewContext(cmd.SessionID, cmd.Type, cmd.Title, cmd.Content)
	context.Metadata = cmd.Metadata
	context.Priority = cmd.Priority
	
	// 检查是否需要压缩
	if context.TokenCount > 1000 && s.compressor != nil {
		if err := s.compressContext(context, cmd.CompressionLevel); err != nil {
			s.logger.Warn("Failed to compress context", zap.Error(err))
		}
	}
	
	// 添加到会话
	if err := session.AddContext(context); err != nil {
		return &application.Result{Success: false, Error: err.Error()}, err
	}
	
	// 保存上下文和会话
	if err := s.contextRepo.Save(ctx, context); err != nil {
		s.logger.Error("Failed to save context", zap.Error(err))
		return &application.Result{Success: false, Error: "failed to save context"}, err
	}
	
	if err := s.sessionRepo.Save(ctx, session); err != nil {
		s.logger.Error("Failed to update session", zap.Error(err))
		return &application.Result{Success: false, Error: "failed to update session"}, err
	}
	
	// 发布事件
	for _, event := range context.GetDomainEvents() {
		if err := s.eventBus.Publish(ctx, event); err != nil {
			s.logger.Warn("Failed to publish event", zap.Error(err))
		}
	}
	context.ClearDomainEvents()
	
	for _, event := range session.GetDomainEvents() {
		if err := s.eventBus.Publish(ctx, event); err != nil {
			s.logger.Warn("Failed to publish event", zap.Error(err))
		}
	}
	session.ClearDomainEvents()
	
	// 记录上下文指标
	if s.metrics != nil {
		s.logger.Info("Context added", 
			zap.String("context_type", string(context.Type)),
			zap.Int("token_count", context.TokenCount),
			zap.Bool("is_compressed", context.IsCompressed),
		)
	}
	
	return &application.Result{Success: true, Data: context}, nil
}

// compressContext 压缩上下文
func (s *MCPService) compressContext(context *domain.Context, level domain.CompressionLevel) error {
	if s.compressor == nil {
		return fmt.Errorf("no compressor available")
	}
	
	compressedContent, err := s.compressor.Compress(context.Content, level)
	if err != nil {
		return err
	}
	
	return context.Compress(level, compressedContent)
}

// GetContext 获取上下文
func (s *MCPService) GetContext(ctx context.Context, query *GetContextQuery) (*application.Result, error) {
	if err := query.Validate(); err != nil {
		return &application.Result{Success: false, Error: err.Error()}, err
	}
	
	// 获取上下文
	context, err := s.contextRepo.FindByID(ctx, query.ContextID)
	if err != nil {
		return &application.Result{Success: false, Error: "context not found"}, err
	}
	
	// 记录访问
	context.Access()
	if err := s.contextRepo.Save(ctx, context); err != nil {
		s.logger.Warn("Failed to update context access", zap.Error(err))
	}
	
	// 如果上下文被压缩且需要解压，进行解压缩
	if context.IsCompressed && query.Decompress && s.compressor != nil {
		originalContent, err := s.compressor.Decompress(context.Content, context.CompressionLevel)
		if err != nil {
			s.logger.Warn("Failed to decompress context", zap.Error(err))
		} else {
			// 创建临时的解压缩版本用于返回（不修改存储的版本）
			decompressedContext := *context
			decompressedContext.Content = originalContent
			decompressedContext.IsCompressed = false
			return &application.Result{Success: true, Data: &decompressedContext}, nil
		}
	}
	
	return &application.Result{Success: true, Data: context}, nil
}

// GetSessionContexts 获取会话上下文
func (s *MCPService) GetSessionContexts(ctx context.Context, query *GetSessionContextsQuery) (*application.Result, error) {
	if err := query.Validate(); err != nil {
		return &application.Result{Success: false, Error: err.Error()}, err
	}
	
	// 获取会话
	session, err := s.sessionRepo.FindByID(ctx, query.SessionID)
	if err != nil {
		return &application.Result{Success: false, Error: "session not found"}, err
	}
	
	// 更新会话活动时间
	session.UpdateActivity()
	s.sessionRepo.Save(ctx, session)
	
	// 获取上下文
	contexts, err := s.contextRepo.FindBySessionID(ctx, query.SessionID)
	if err != nil {
		return &application.Result{Success: false, Error: "failed to get contexts"}, err
	}
	
	// 过滤和排序
	var filteredContexts []*domain.Context
	for _, context := range contexts {
		// 按类型过滤
		if query.Type != nil && context.Type != *query.Type {
			continue
		}
		
		// 按优先级过滤
		if query.MinPriority > 0 && context.Priority < query.MinPriority {
			continue
		}
		
		filteredContexts = append(filteredContexts, context)
	}
	
	// 应用分页
	offset := (query.Page - 1) * query.PageSize
	if offset >= len(filteredContexts) {
		filteredContexts = []*domain.Context{}
	} else {
		end := offset + query.PageSize
		if end > len(filteredContexts) {
			end = len(filteredContexts)
		}
		filteredContexts = filteredContexts[offset:end]
	}
	
	return &application.Result{Success: true, Data: map[string]interface{}{
		"contexts":  filteredContexts,
		"total":     len(contexts),
		"session":   session,
		"page":      query.Page,
		"page_size": query.PageSize,
	}}, nil
}

// CleanupExpiredSessions 清理过期会话
func (s *MCPService) CleanupExpiredSessions(ctx context.Context) error {
	expiredSessions, err := s.sessionRepo.FindExpiredSessions(ctx)
	if err != nil {
		return err
	}
	
	for _, session := range expiredSessions {
		session.Expire()
		if err := s.sessionRepo.Save(ctx, session); err != nil {
			s.logger.Error("Failed to expire session", zap.String("session_id", session.ID.String()), zap.Error(err))
		} else {
			s.logger.Info("Session expired", zap.String("session_id", session.ID.String()))
		}
	}
	
	return nil
}

// ManageIdleSessions 管理空闲会话
func (s *MCPService) ManageIdleSessions(ctx context.Context, idleThreshold time.Duration) error {
	idleSessions, err := s.sessionRepo.FindIdleSessions(ctx, idleThreshold)
	if err != nil {
		return err
	}
	
	for _, session := range idleSessions {
		session.SetIdle()
		if err := s.sessionRepo.Save(ctx, session); err != nil {
			s.logger.Error("Failed to set session idle", zap.String("session_id", session.ID.String()), zap.Error(err))
		} else {
			s.logger.Info("Session set to idle", zap.String("session_id", session.ID.String()))
		}
	}
	
	return nil
}

// ContextCompressor 上下文压缩器接口
type ContextCompressor interface {
	Compress(content string, level domain.CompressionLevel) (string, error)
	Decompress(compressedContent string, level domain.CompressionLevel) (string, error)
}

// SimpleCompressor 简单压缩器实现
type SimpleCompressor struct{}

// NewSimpleCompressor 创建简单压缩器
func NewSimpleCompressor() ContextCompressor {
	return &SimpleCompressor{}
}

// Compress 压缩内容
func (c *SimpleCompressor) Compress(content string, level domain.CompressionLevel) (string, error) {
	// 这里实现简单的压缩逻辑
	// 实际应用中可能使用gzip、lz4等压缩算法
	switch level {
	case domain.CompressionLight:
		// 轻度压缩：移除多余空格
		return compressSpaces(content), nil
	case domain.CompressionMedium:
		// 中度压缩：摘要化
		return summarizeContent(content), nil
	case domain.CompressionHeavy:
		// 重度压缩：关键词提取
		return extractKeywords(content), nil
	default:
		return content, nil
	}
}

// Decompress 解压缩内容
func (c *SimpleCompressor) Decompress(compressedContent string, level domain.CompressionLevel) (string, error) {
	// 简单实现，实际应用中需要保存原始内容或使用可逆压缩
	return compressedContent, nil
}

// 辅助函数
func compressSpaces(content string) string {
	// 简化实现：移除多余空格
	return content // 实际实现会进行空格压缩
}

func summarizeContent(content string) string {
	// 简化实现：内容摘要
	if len(content) > 500 {
		return content[:500] + "..."
	}
	return content
}

func extractKeywords(content string) string {
	// 简化实现：关键词提取
	if len(content) > 200 {
		return content[:200] + "..."
	}
	return content
}
