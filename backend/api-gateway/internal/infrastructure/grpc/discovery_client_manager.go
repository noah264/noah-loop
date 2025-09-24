package grpc

import (
	"context"
	"fmt"
	"sync"
	"time"

	agentv1 "github.com/noah-loop/backend/api-gateway/proto/agent/v1"
	llmv1 "github.com/noah-loop/backend/api-gateway/proto/llm/v1"
	"github.com/noah-loop/backend/api-gateway/internal/infrastructure/etcd"
	"github.com/noah-loop/backend/shared/pkg/infrastructure"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

// DiscoveryClientManager 基于服务发现的gRPC客户端管理器
type DiscoveryClientManager struct {
	discovery   *etcd.ServiceDiscovery
	logger      infrastructure.Logger
	connections map[string]map[string]*grpc.ClientConn // service -> instance -> connection
	clients     map[string]interface{}                 // service -> client
	mutex       sync.RWMutex
	
	// 服务监听上下文
	watchContexts map[string]context.CancelFunc
	
	// 负载均衡器
	loadBalancer *LoadBalancer
}

// LoadBalancer 简单的负载均衡器
type LoadBalancer struct {
	counters map[string]uint64
	mutex    sync.Mutex
}

// NewLoadBalancer 创建负载均衡器
func NewLoadBalancer() *LoadBalancer {
	return &LoadBalancer{
		counters: make(map[string]uint64),
	}
}

// Next 获取下一个实例（轮询）
func (lb *LoadBalancer) Next(serviceName string, instances []*etcd.ServiceInfo) *etcd.ServiceInfo {
	if len(instances) == 0 {
		return nil
	}
	
	lb.mutex.Lock()
	defer lb.mutex.Unlock()
	
	counter := lb.counters[serviceName]
	lb.counters[serviceName] = counter + 1
	
	return instances[counter%uint64(len(instances))]
}

// NewDiscoveryClientManager 创建基于服务发现的客户端管理器
func NewDiscoveryClientManager(discovery *etcd.ServiceDiscovery, logger infrastructure.Logger) *DiscoveryClientManager {
	return &DiscoveryClientManager{
		discovery:     discovery,
		logger:        logger,
		connections:   make(map[string]map[string]*grpc.ClientConn),
		clients:       make(map[string]interface{}),
		watchContexts: make(map[string]context.CancelFunc),
		loadBalancer:  NewLoadBalancer(),
	}
}

// Initialize 初始化客户端管理器
func (dcm *DiscoveryClientManager) Initialize(ctx context.Context) error {
	// 初始化各服务的服务发现
	services := []string{"agent", "llm", "mcp", "orchestrator"}
	
	for _, serviceName := range services {
		if err := dcm.startServiceDiscovery(ctx, serviceName); err != nil {
			dcm.logger.Error("Failed to start service discovery",
				zap.String("service", serviceName),
				zap.Error(err))
			return err
		}
	}
	
	return nil
}

// startServiceDiscovery 启动服务发现
func (dcm *DiscoveryClientManager) startServiceDiscovery(ctx context.Context, serviceName string) error {
	// 创建监听上下文
	watchCtx, cancel := context.WithCancel(ctx)
	dcm.watchContexts[serviceName] = cancel
	
	// 开始监听服务变化
	serviceCh, err := dcm.discovery.WatchService(watchCtx, serviceName)
	if err != nil {
		cancel()
		return fmt.Errorf("failed to watch service %s: %w", serviceName, err)
	}
	
	// 初始化连接映射
	dcm.mutex.Lock()
	dcm.connections[serviceName] = make(map[string]*grpc.ClientConn)
	dcm.mutex.Unlock()
	
	// 启动协程处理服务变化
	go func() {
		defer cancel()
		
		for {
			select {
			case <-watchCtx.Done():
				dcm.logger.Info("Service discovery stopped", zap.String("service", serviceName))
				return
			case services, ok := <-serviceCh:
				if !ok {
					dcm.logger.Error("Service channel closed", zap.String("service", serviceName))
					return
				}
				
				dcm.updateServiceConnections(serviceName, services)
			}
		}
	}()
	
	dcm.logger.Info("Service discovery started", zap.String("service", serviceName))
	return nil
}

// updateServiceConnections 更新服务连接
func (dcm *DiscoveryClientManager) updateServiceConnections(serviceName string, services []*etcd.ServiceInfo) {
	dcm.mutex.Lock()
	defer dcm.mutex.Unlock()
	
	// 获取当前连接
	currentConns := dcm.connections[serviceName]
	newConns := make(map[string]*grpc.ClientConn)
	
	// 为新发现的服务实例创建连接
	for _, service := range services {
		if service.Health != "healthy" {
			continue // 跳过不健康的实例
		}
		
		instanceKey := fmt.Sprintf("%s:%d", service.Address, service.GRPCPort)
		
		// 如果连接已存在，复用
		if conn, exists := currentConns[instanceKey]; exists {
			newConns[instanceKey] = conn
			delete(currentConns, instanceKey) // 从当前连接中移除，避免关闭
			continue
		}
		
		// 创建新连接
		address := fmt.Sprintf("%s:%d", service.Address, service.GRPCPort)
		conn, err := dcm.createConnection(address)
		if err != nil {
			dcm.logger.Error("Failed to create connection",
				zap.String("service", serviceName),
				zap.String("address", address),
				zap.Error(err))
			continue
		}
		
		newConns[instanceKey] = conn
		dcm.logger.Info("Created new gRPC connection",
			zap.String("service", serviceName),
			zap.String("address", address))
	}
	
	// 关闭不再需要的连接
	for instanceKey, conn := range currentConns {
		conn.Close()
		dcm.logger.Info("Closed gRPC connection",
			zap.String("service", serviceName),
			zap.String("instance", instanceKey))
	}
	
	// 更新连接映射
	dcm.connections[serviceName] = newConns
	
	// 更新客户端
	dcm.updateClients(serviceName, services)
}

// updateClients 更新客户端
func (dcm *DiscoveryClientManager) updateClients(serviceName string, services []*etcd.ServiceInfo) {
	// 获取健康的服务实例
	healthyServices := make([]*etcd.ServiceInfo, 0)
	for _, service := range services {
		if service.Health == "healthy" {
			healthyServices = append(healthyServices, service)
		}
	}
	
	if len(healthyServices) == 0 {
		dcm.logger.Warn("No healthy instances found", zap.String("service", serviceName))
		dcm.clients[serviceName] = nil
		return
	}
	
	// 使用负载均衡选择实例
	selectedService := dcm.loadBalancer.Next(serviceName, healthyServices)
	if selectedService == nil {
		dcm.logger.Warn("Load balancer returned no instance", zap.String("service", serviceName))
		dcm.clients[serviceName] = nil
		return
	}
	
	// 获取连接
	instanceKey := fmt.Sprintf("%s:%d", selectedService.Address, selectedService.GRPCPort)
	conn, exists := dcm.connections[serviceName][instanceKey]
	if !exists {
		dcm.logger.Error("Connection not found for selected instance",
			zap.String("service", serviceName),
			zap.String("instance", instanceKey))
		dcm.clients[serviceName] = nil
		return
	}
	
	// 创建客户端
	switch serviceName {
	case "agent":
		dcm.clients[serviceName] = agentv1.NewAgentServiceClient(conn)
	case "llm":
		dcm.clients[serviceName] = llmv1.NewLLMServiceClient(conn)
	// case "mcp":
	//     dcm.clients[serviceName] = mcpv1.NewMCPServiceClient(conn)
	// case "orchestrator":
	//     dcm.clients[serviceName] = orchestratorv1.NewOrchestratorServiceClient(conn)
	}
	
	dcm.logger.Debug("Updated client for service",
		zap.String("service", serviceName),
		zap.String("selected_instance", instanceKey))
}

// createConnection 创建gRPC连接
func (dcm *DiscoveryClientManager) createConnection(address string) (*grpc.ClientConn, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	
	opts := []grpc.DialOption{
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithTimeout(10 * time.Second),
		grpc.WithBlock(),
		grpc.WithDefaultCallOptions(
			grpc.MaxCallRecvMsgSize(4*1024*1024), // 4MB
			grpc.MaxCallSendMsgSize(4*1024*1024), // 4MB
		),
	}
	
	return grpc.DialContext(ctx, address, opts...)
}

// GetAgentClient 获取Agent客户端
func (dcm *DiscoveryClientManager) GetAgentClient() agentv1.AgentServiceClient {
	dcm.mutex.RLock()
	defer dcm.mutex.RUnlock()
	
	if client, ok := dcm.clients["agent"].(agentv1.AgentServiceClient); ok {
		return client
	}
	return nil
}

// GetLLMClient 获取LLM客户端
func (dcm *DiscoveryClientManager) GetLLMClient() llmv1.LLMServiceClient {
	dcm.mutex.RLock()
	defer dcm.mutex.RUnlock()
	
	if client, ok := dcm.clients["llm"].(llmv1.LLMServiceClient); ok {
		return client
	}
	return nil
}

// GetHealthyServices 获取健康服务列表
func (dcm *DiscoveryClientManager) GetHealthyServices(ctx context.Context, serviceName string) ([]*etcd.ServiceInfo, error) {
	return dcm.discovery.GetHealthyServices(ctx, serviceName)
}

// Close 关闭客户端管理器
func (dcm *DiscoveryClientManager) Close() error {
	dcm.mutex.Lock()
	defer dcm.mutex.Unlock()
	
	// 取消所有监听
	for serviceName, cancel := range dcm.watchContexts {
		cancel()
		dcm.logger.Info("Cancelled service discovery", zap.String("service", serviceName))
	}
	
	// 关闭所有连接
	for serviceName, conns := range dcm.connections {
		for instanceKey, conn := range conns {
			conn.Close()
			dcm.logger.Info("Closed connection",
				zap.String("service", serviceName),
				zap.String("instance", instanceKey))
		}
	}
	
	dcm.logger.Info("Discovery client manager closed")
	return nil
}
