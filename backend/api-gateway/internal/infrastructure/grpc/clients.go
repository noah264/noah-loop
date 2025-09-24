package grpc

import (
	"context"
	"fmt"
	"time"

	agentv1 "github.com/noah-loop/backend/api-gateway/proto/agent/v1"
	llmv1 "github.com/noah-loop/backend/api-gateway/proto/llm/v1"
	"github.com/noah-loop/backend/shared/pkg/infrastructure"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

// ClientManager gRPC客户端管理器
type ClientManager struct {
	config      *infrastructure.Config
	logger      infrastructure.Logger
	connections map[string]*grpc.ClientConn
	
	// 客户端
	AgentClient  agentv1.AgentServiceClient
	LLMClient    llmv1.LLMServiceClient
	// MCPClient    mcpv1.MCPServiceClient
	// OrchestratorClient orchestratorv1.OrchestratorServiceClient
}

// NewClientManager 创建gRPC客户端管理器
func NewClientManager(config *infrastructure.Config, logger infrastructure.Logger) (*ClientManager, error) {
	manager := &ClientManager{
		config:      config,
		logger:      logger,
		connections: make(map[string]*grpc.ClientConn),
	}
	
	// 初始化连接
	if err := manager.initializeConnections(); err != nil {
		return nil, fmt.Errorf("failed to initialize gRPC connections: %w", err)
	}
	
	return manager, nil
}

// initializeConnections 初始化gRPC连接
func (cm *ClientManager) initializeConnections() error {
	// Agent服务连接
	if err := cm.connectToAgent(); err != nil {
		cm.logger.Error("Failed to connect to Agent service", zap.Error(err))
		return err
	}
	
	// LLM服务连接
	if err := cm.connectToLLM(); err != nil {
		cm.logger.Error("Failed to connect to LLM service", zap.Error(err))
		return err
	}
	
	// 其他服务连接...
	
	return nil
}

// connectToAgent 连接到Agent服务
func (cm *ClientManager) connectToAgent() error {
	address := fmt.Sprintf("localhost:%d", cm.config.Services.Agent.GRPCPort)
	
	conn, err := cm.createConnection(address, "agent")
	if err != nil {
		return err
	}
	
	cm.connections["agent"] = conn
	cm.AgentClient = agentv1.NewAgentServiceClient(conn)
	
	cm.logger.Info("Connected to Agent service via gRPC", 
		zap.String("address", address))
	
	return nil
}

// connectToLLM 连接到LLM服务
func (cm *ClientManager) connectToLLM() error {
	address := fmt.Sprintf("localhost:%d", cm.config.Services.LLM.GRPCPort)
	
	conn, err := cm.createConnection(address, "llm")
	if err != nil {
		return err
	}
	
	cm.connections["llm"] = conn
	cm.LLMClient = llmv1.NewLLMServiceClient(conn)
	
	cm.logger.Info("Connected to LLM service via gRPC", 
		zap.String("address", address))
	
	return nil
}

// createConnection 创建gRPC连接
func (cm *ClientManager) createConnection(address, serviceName string) (*grpc.ClientConn, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	
	// 连接配置
	opts := []grpc.DialOption{
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithTimeout(10 * time.Second),
		grpc.WithBlock(), // 阻塞直到连接建立
		grpc.WithDefaultCallOptions(
			grpc.MaxCallRecvMsgSize(4*1024*1024), // 4MB
			grpc.MaxCallSendMsgSize(4*1024*1024), // 4MB
		),
	}
	
	conn, err := grpc.DialContext(ctx, address, opts...)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to %s service at %s: %w", serviceName, address, err)
	}
	
	return conn, nil
}

// HealthCheck 检查所有服务连接健康状态
func (cm *ClientManager) HealthCheck(ctx context.Context) map[string]bool {
	results := make(map[string]bool)
	
	// 检查Agent服务
	if cm.AgentClient != nil {
		_, err := cm.AgentClient.GetAgent(ctx, &agentv1.GetAgentRequest{AgentId: "health-check"})
		results["agent"] = err == nil // 简化的健康检查
	}
	
	// 检查LLM服务
	if cm.LLMClient != nil {
		_, err := cm.LLMClient.GetModel(ctx, &llmv1.GetModelRequest{ModelId: "health-check"})
		results["llm"] = err == nil
	}
	
	return results
}

// Close 关闭所有连接
func (cm *ClientManager) Close() error {
	for serviceName, conn := range cm.connections {
		if err := conn.Close(); err != nil {
			cm.logger.Error("Failed to close gRPC connection", 
				zap.String("service", serviceName), 
				zap.Error(err))
		} else {
			cm.logger.Info("Closed gRPC connection", 
				zap.String("service", serviceName))
		}
	}
	
	return nil
}

// GetConnection 获取指定服务的连接
func (cm *ClientManager) GetConnection(serviceName string) (*grpc.ClientConn, bool) {
	conn, exists := cm.connections[serviceName]
	return conn, exists
}
