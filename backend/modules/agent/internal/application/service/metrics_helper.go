package service

import (
	"context"
	"time"
	
	"github.com/noah-loop/backend/modules/agent/internal/domain"
	"go.uber.org/zap"
)

// updateAgentMetrics 更新智能体指标（异步执行）
func (s *AgentService) updateAgentMetrics() {
	if s.metrics == nil {
		return
	}
	
	ctx := context.Background()
	
	// 获取活跃智能体统计
	activeAgents, err := s.agentRepo.FindActiveAgents(ctx)
	if err != nil {
		s.logger.Warn("Failed to get active agents for metrics", zap.Error(err))
		return
	}
	
	// 按类型和状态分组统计
	agentStats := make(map[string]map[string]int)
	
	for _, agent := range activeAgents {
		agentType := string(agent.Type)
		status := string(agent.Status)
		
		if agentStats[agentType] == nil {
			agentStats[agentType] = make(map[string]int)
		}
		agentStats[agentType][status]++
	}
	
	// 更新Prometheus指标
	for agentType, statusMap := range agentStats {
		for status, count := range statusMap {
			s.metrics.SetActiveAgents(agentType, status, count)
		}
	}
}

// RecordChatMetrics 记录对话指标
func (s *AgentService) RecordChatMetrics(agentType string, duration time.Duration, success bool) {
	if s.metrics == nil {
		return
	}
	
	// 这里可以添加自定义的对话指标
	// 比如对话轮数、响应时间等
	status := "success"
	if !success {
		status = "failed"
	}
	
	s.logger.Info("Chat metrics recorded",
		zap.String("agent_type", agentType),
		zap.Duration("duration", duration),
		zap.String("status", status),
	)
}

// StartMetricsCollection 启动指标收集（定期执行）
func (s *AgentService) StartMetricsCollection() {
	if s.metrics == nil {
		return
	}
	
	ticker := time.NewTicker(30 * time.Second) // 每30秒更新一次
	go func() {
		defer ticker.Stop()
		for range ticker.C {
			s.updateAgentMetrics()
		}
	}()
}
