package service

import (
	"fmt"
	"sync/atomic"

	"github.com/noah-loop/backend/api-gateway/internal/domain/entity"
	"github.com/noah-loop/backend/shared/pkg/domain"
)

// LoadBalancerStrategy 负载均衡策略
type LoadBalancerStrategy string

const (
	StrategyRoundRobin LoadBalancerStrategy = "round_robin"
	StrategyWeighted   LoadBalancerStrategy = "weighted"
	StrategyLeastConn  LoadBalancerStrategy = "least_conn"
)

// LoadBalancer 负载均衡器领域服务
type LoadBalancer struct {
	strategy LoadBalancerStrategy
	counter  uint64
}

// NewLoadBalancer 创建负载均衡器
func NewLoadBalancer(strategy LoadBalancerStrategy) *LoadBalancer {
	return &LoadBalancer{
		strategy: strategy,
		counter:  0,
	}
}

// SelectService 选择服务
func (lb *LoadBalancer) SelectService(services []*entity.Service) (*entity.Service, error) {
	if len(services) == 0 {
		return nil, domain.NewDomainError("NO_SERVICES_AVAILABLE", "No services available for load balancing")
	}
	
	// 过滤健康的服务
	healthyServices := make([]*entity.Service, 0)
	for _, service := range services {
		if service.IsHealthy() {
			healthyServices = append(healthyServices, service)
		}
	}
	
	if len(healthyServices) == 0 {
		return nil, domain.NewDomainError("NO_HEALTHY_SERVICES", "No healthy services available")
	}
	
	switch lb.strategy {
	case StrategyRoundRobin:
		return lb.roundRobin(healthyServices), nil
	case StrategyWeighted:
		return lb.weighted(healthyServices), nil
	case StrategyLeastConn:
		return lb.leastConnection(healthyServices), nil
	default:
		return lb.roundRobin(healthyServices), nil
	}
}

// roundRobin 轮询算法
func (lb *LoadBalancer) roundRobin(services []*entity.Service) *entity.Service {
	index := atomic.AddUint64(&lb.counter, 1)
	return services[(index-1)%uint64(len(services))]
}

// weighted 加权轮询算法
func (lb *LoadBalancer) weighted(services []*entity.Service) *entity.Service {
	// 构建加权服务列表
	var weightedServices []*entity.Service
	for _, service := range services {
		weight := service.GetWeight()
		for i := 0; i < weight; i++ {
			weightedServices = append(weightedServices, service)
		}
	}
	
	if len(weightedServices) == 0 {
		return services[0] // fallback
	}
	
	index := atomic.AddUint64(&lb.counter, 1)
	return weightedServices[(index-1)%uint64(len(weightedServices))]
}

// leastConnection 最少连接算法（简化版）
func (lb *LoadBalancer) leastConnection(services []*entity.Service) *entity.Service {
	// 这里简化处理，实际需要追踪连接数
	// 目前返回第一个健康的服务
	return services[0]
}

// GetStrategy 获取策略
func (lb *LoadBalancer) GetStrategy() LoadBalancerStrategy {
	return lb.strategy
}

// SetStrategy 设置策略
func (lb *LoadBalancer) SetStrategy(strategy LoadBalancerStrategy) {
	lb.strategy = strategy
	lb.counter = 0 // 重置计数器
}
