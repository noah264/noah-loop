package service

import (
	"context"
	"time"
	
	"github.com/noah-loop/backend/modules/mcp/internal/domain"
	"go.uber.org/zap"
)

// updateSessionMetrics 更新会话指标（异步执行）
func (s *MCPService) updateSessionMetrics() {
	if s.metrics == nil {
		return
	}
	
	ctx := context.Background()
	
	// 统计各种状态的会话数量
	sessionStats := make(map[string]int)
	
	allStatuses := []domain.SessionStatus{
		domain.SessionStatusActive,
		domain.SessionStatusIdle,
		domain.SessionStatusArchived,
		domain.SessionStatusExpired,
	}
	
	for _, status := range allStatuses {
		sessions, err := s.sessionRepo.FindByStatus(ctx, status)
		if err != nil {
			s.logger.Warn("Failed to get sessions by status for metrics", 
				zap.String("status", string(status)), 
				zap.Error(err))
			continue
		}
		sessionStats[string(status)] = len(sessions)
	}
	
	// 更新Prometheus指标
	for status, count := range sessionStats {
		s.metrics.SetActiveSessions(status, count)
	}
}

// RecordCompressionMetrics 记录压缩指标
func (s *MCPService) RecordCompressionMetrics(contextType string, originalSize, compressedSize int, level domain.CompressionLevel) {
	if s.metrics == nil {
		return
	}
	
	compressionRatio := float64(compressedSize) / float64(originalSize)
	
	s.logger.Info("Context compressed",
		zap.String("context_type", contextType),
		zap.Int("original_size", originalSize),
		zap.Int("compressed_size", compressedSize),
		zap.Float64("compression_ratio", compressionRatio),
		zap.String("compression_level", string(level)),
	)
}

// RecordContextUsage 记录上下文使用指标
func (s *MCPService) RecordContextUsage(sessionID, contextID string, accessCount int) {
	if s.metrics == nil {
		return
	}
	
	s.logger.Info("Context accessed",
		zap.String("session_id", sessionID),
		zap.String("context_id", contextID),
		zap.Int("access_count", accessCount),
	)
}

// StartMetricsCollection 启动指标收集
func (s *MCPService) StartMetricsCollection() {
	if s.metrics == nil {
		return
	}
	
	ticker := time.NewTicker(1 * time.Minute) // 每分钟更新一次
	go func() {
		defer ticker.Stop()
		for range ticker.C {
			s.updateSessionMetrics()
		}
	}()
}
