package service

import (
	"context"
	"time"
	
	"github.com/noah-loop/backend/modules/orchestrator/internal/domain"
	"go.uber.org/zap"
)

// RecordStepExecution 记录步骤执行指标
func (s *OrchestratorService) RecordStepExecution(workflowID, stepID string, stepType domain.StepType, duration time.Duration, success bool) {
	if s.metrics == nil {
		return
	}
	
	status := "success"
	if !success {
		status = "failed"
	}
	
	s.logger.Info("Step execution recorded",
		zap.String("workflow_id", workflowID),
		zap.String("step_id", stepID),
		zap.String("step_type", string(stepType)),
		zap.Duration("duration", duration),
		zap.String("status", status),
	)
}

// UpdateWorkflowMetrics 更新工作流指标
func (s *OrchestratorService) UpdateWorkflowMetrics() {
	if s.metrics == nil {
		return
	}
	
	ctx := context.Background()
	
	// 获取运行中的执行数量
	runningExecutions, err := s.executionRepo.FindByStatus(ctx, domain.ExecutionStatusRunning)
	if err != nil {
		s.logger.Warn("Failed to get running executions for metrics", zap.Error(err))
		return
	}
	
	s.logger.Info("Workflow metrics updated",
		zap.Int("running_executions", len(runningExecutions)),
	)
}

// StartMetricsCollection 启动指标收集
func (s *OrchestratorService) StartMetricsCollection() {
	if s.metrics == nil {
		return
	}
	
	ticker := time.NewTicker(30 * time.Second) // 每30秒更新一次
	go func() {
		defer ticker.Stop()
		for range ticker.C {
			s.UpdateWorkflowMetrics()
		}
	}()
}
