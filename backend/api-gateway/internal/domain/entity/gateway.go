package entity

import (
	"sync"
	"time"

	"github.com/noah-loop/backend/api-gateway/internal/domain/valueobject"
	"github.com/noah-loop/backend/shared/pkg/domain"
)

// Gateway API网关聚合根
type Gateway struct {
	domain.AggregateRoot
	name        string
	version     string
	services    map[string]*Service
	routes      []*valueobject.Route
	status      GatewayStatus
	createdAt   time.Time
	updatedAt   time.Time
	mutex       sync.RWMutex
}

// GatewayStatus 网关状态
type GatewayStatus string

const (
	GatewayStatusStarting GatewayStatus = "starting"
	GatewayStatusRunning  GatewayStatus = "running" 
	GatewayStatusStopping GatewayStatus = "stopping"
	GatewayStatusStopped  GatewayStatus = "stopped"
)

// NewGateway 创建网关实例
func NewGateway(name, version string) *Gateway {
	gateway := &Gateway{
		AggregateRoot: domain.NewAggregateRoot(),
		name:          name,
		version:       version,
		services:      make(map[string]*Service),
		routes:        make([]*valueobject.Route, 0),
		status:        GatewayStatusStarting,
		createdAt:     time.Now(),
		updatedAt:     time.Now(),
	}
	
	// 发布网关创建事件
	gateway.PublishEvent(domain.NewDomainEvent("gateway.created", gateway))
	
	return gateway
}

// AddService 添加服务
func (g *Gateway) AddService(service *Service) error {
	g.mutex.Lock()
	defer g.mutex.Unlock()
	
	if _, exists := g.services[service.GetName()]; exists {
		return domain.NewDomainError("SERVICE_ALREADY_EXISTS", "Service already exists: "+service.GetName())
	}
	
	g.services[service.GetName()] = service
	g.updatedAt = time.Now()
	
	// 发布服务添加事件
	g.PublishEvent(domain.NewDomainEvent("gateway.service_added", map[string]interface{}{
		"gateway_id":   g.GetID(),
		"service_name": service.GetName(),
	}))
	
	return nil
}

// RemoveService 移除服务
func (g *Gateway) RemoveService(serviceName string) error {
	g.mutex.Lock()
	defer g.mutex.Unlock()
	
	if _, exists := g.services[serviceName]; !exists {
		return domain.NewDomainError("SERVICE_NOT_FOUND", "Service not found: "+serviceName)
	}
	
	delete(g.services, serviceName)
	g.updatedAt = time.Now()
	
	// 发布服务移除事件
	g.PublishEvent(domain.NewDomainEvent("gateway.service_removed", map[string]interface{}{
		"gateway_id":   g.GetID(),
		"service_name": serviceName,
	}))
	
	return nil
}

// GetService 获取服务
func (g *Gateway) GetService(serviceName string) (*Service, error) {
	g.mutex.RLock()
	defer g.mutex.RUnlock()
	
	service, exists := g.services[serviceName]
	if !exists {
		return nil, domain.NewDomainError("SERVICE_NOT_FOUND", "Service not found: "+serviceName)
	}
	
	return service, nil
}

// GetAllServices 获取所有服务
func (g *Gateway) GetAllServices() map[string]*Service {
	g.mutex.RLock()
	defer g.mutex.RUnlock()
	
	// 返回副本以避免并发修改
	services := make(map[string]*Service)
	for name, service := range g.services {
		services[name] = service
	}
	
	return services
}

// GetHealthyServices 获取健康的服务
func (g *Gateway) GetHealthyServices() []*Service {
	g.mutex.RLock()
	defer g.mutex.RUnlock()
	
	var healthyServices []*Service
	for _, service := range g.services {
		if service.IsHealthy() {
			healthyServices = append(healthyServices, service)
		}
	}
	
	return healthyServices
}

// AddRoute 添加路由
func (g *Gateway) AddRoute(route *valueobject.Route) {
	g.mutex.Lock()
	defer g.mutex.Unlock()
	
	g.routes = append(g.routes, route)
	g.updatedAt = time.Now()
	
	// 发布路由添加事件
	g.PublishEvent(domain.NewDomainEvent("gateway.route_added", map[string]interface{}{
		"gateway_id": g.GetID(),
		"route":      route,
	}))
}

// GetRoutes 获取所有路由
func (g *Gateway) GetRoutes() []*valueobject.Route {
	g.mutex.RLock()
	defer g.mutex.RUnlock()
	
	// 返回副本
	routes := make([]*valueobject.Route, len(g.routes))
	copy(routes, g.routes)
	
	return routes
}

// FindRoute 查找路由
func (g *Gateway) FindRoute(path string) (*valueobject.Route, error) {
	g.mutex.RLock()
	defer g.mutex.RUnlock()
	
	for _, route := range g.routes {
		if route.Matches(path) {
			return route, nil
		}
	}
	
	return nil, domain.NewDomainError("ROUTE_NOT_FOUND", "Route not found for path: "+path)
}

// Start 启动网关
func (g *Gateway) Start() {
	g.mutex.Lock()
	defer g.mutex.Unlock()
	
	g.status = GatewayStatusRunning
	g.updatedAt = time.Now()
	
	// 发布网关启动事件
	g.PublishEvent(domain.NewDomainEvent("gateway.started", g))
}

// Stop 停止网关
func (g *Gateway) Stop() {
	g.mutex.Lock()
	defer g.mutex.Unlock()
	
	g.status = GatewayStatusStopping
	g.updatedAt = time.Now()
	
	// 发布网关停止事件
	g.PublishEvent(domain.NewDomainEvent("gateway.stopping", g))
}

// MarkStopped 标记为已停止
func (g *Gateway) MarkStopped() {
	g.mutex.Lock()
	defer g.mutex.Unlock()
	
	g.status = GatewayStatusStopped
	g.updatedAt = time.Now()
	
	// 发布网关已停止事件
	g.PublishEvent(domain.NewDomainEvent("gateway.stopped", g))
}

// IsRunning 检查是否运行中
func (g *Gateway) IsRunning() bool {
	g.mutex.RLock()
	defer g.mutex.RUnlock()
	
	return g.status == GatewayStatusRunning
}

// GetName 获取网关名称
func (g *Gateway) GetName() string {
	return g.name
}

// GetVersion 获取网关版本
func (g *Gateway) GetVersion() string {
	return g.version
}

// GetStatus 获取网关状态
func (g *Gateway) GetStatus() GatewayStatus {
	g.mutex.RLock()
	defer g.mutex.RUnlock()
	
	return g.status
}

// GetCreatedAt 获取创建时间
func (g *Gateway) GetCreatedAt() time.Time {
	return g.createdAt
}

// GetUpdatedAt 获取更新时间
func (g *Gateway) GetUpdatedAt() time.Time {
	g.mutex.RLock()
	defer g.mutex.RUnlock()
	
	return g.updatedAt
}

// GetServiceCount 获取服务数量
func (g *Gateway) GetServiceCount() int {
	g.mutex.RLock()
	defer g.mutex.RUnlock()
	
	return len(g.services)
}

// GetHealthyServiceCount 获取健康服务数量
func (g *Gateway) GetHealthyServiceCount() int {
	g.mutex.RLock()
	defer g.mutex.RUnlock()
	
	count := 0
	for _, service := range g.services {
		if service.IsHealthy() {
			count++
		}
	}
	
	return count
}
