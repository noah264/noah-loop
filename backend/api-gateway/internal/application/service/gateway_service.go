package service

import (
	"context"
	"net/http"
	"strconv"
	"time"

	"github.com/noah-loop/backend/api-gateway/internal/domain/entity"
	"github.com/noah-loop/backend/api-gateway/internal/domain/repository"
	domainService "github.com/noah-loop/backend/api-gateway/internal/domain/service"
	"github.com/noah-loop/backend/api-gateway/internal/domain/valueobject"
	"github.com/noah-loop/backend/shared/pkg/infrastructure"
	"go.uber.org/zap"
)

// GatewayService 网关应用服务
type GatewayService struct {
	gateway         *entity.Gateway
	serviceRepo     repository.ServiceRepository
	loadBalancer    *domainService.LoadBalancer
	circuitBreakers map[string]*domainService.CircuitBreaker
	config          GatewayConfig
	logger          infrastructure.Logger
	metrics         *infrastructure.MetricsRegistry
	healthTicker    *time.Ticker
	stopHealthCheck chan bool
}

// GatewayConfig 网关配置接口
type GatewayConfig interface {
	GetGatewayName() string
	GetGatewayVersion() string
	GetServices() map[string]ServiceConfig
}

// ServiceConfig 服务配置
type ServiceConfig struct {
	Name string
	Host string
	Port int
	Path string
}

// NewGatewayService 创建网关应用服务
func NewGatewayService(
	config GatewayConfig,
	serviceRepo repository.ServiceRepository,
	logger infrastructure.Logger,
	metrics *infrastructure.MetricsRegistry,
) *GatewayService {
	gateway := entity.NewGateway(config.GetGatewayName(), config.GetGatewayVersion())
	loadBalancer := domainService.NewLoadBalancer(domainService.StrategyRoundRobin)
	circuitBreakers := make(map[string]*domainService.CircuitBreaker)
	
	return &GatewayService{
		gateway:         gateway,
		serviceRepo:     serviceRepo,
		loadBalancer:    loadBalancer,
		circuitBreakers: circuitBreakers,
		config:          config,
		logger:          logger,
		metrics:         metrics,
		stopHealthCheck: make(chan bool, 1),
	}
}

// Initialize 初始化网关
func (gs *GatewayService) Initialize() error {
	ctx := context.Background()
	
	// 初始化服务
	services := gs.config.GetServices()
	for _, serviceConfig := range services {
		if err := gs.registerService(ctx, serviceConfig); err != nil {
			gs.logger.Error("Failed to register service", 
				zap.String("service", serviceConfig.Name), 
				zap.Error(err))
			return err
		}
	}
	
	// 初始化路由
	gs.setupDefaultRoutes()
	
	// 启动网关
	gs.gateway.Start()
	
	gs.logger.Info("Gateway initialized successfully",
		zap.String("name", gs.gateway.GetName()),
		zap.String("version", gs.gateway.GetVersion()),
		zap.Int("services", gs.gateway.GetServiceCount()))
	
	return nil
}

// registerService 注册服务
func (gs *GatewayService) registerService(ctx context.Context, config ServiceConfig) error {
	serviceEntity := entity.NewService(entity.ServiceConfig{
		Name: config.Name,
		Host: config.Host,
		Port: config.Port,
		Path: config.Path,
	})
	
	// 保存到仓储
	if err := gs.serviceRepo.Save(ctx, serviceEntity); err != nil {
		return err
	}
	
	// 添加到网关
	if err := gs.gateway.AddService(serviceEntity); err != nil {
		return err
	}
	
	// 创建熔断器
	circuitBreaker := domainService.NewCircuitBreaker(domainService.CircuitBreakerConfig{
		ServiceName:     config.Name,
		MaxFailures:     5,
		Timeout:         60 * time.Second,
		HalfOpenMaxReqs: 3,
	})
	gs.circuitBreakers[config.Name] = circuitBreaker
	
	return nil
}

// setupDefaultRoutes 设置默认路由
func (gs *GatewayService) setupDefaultRoutes() {
	for _, service := range gs.gateway.GetAllServices() {
		route, err := valueobject.NewRoute(valueobject.RouteConfig{
			Pattern:     "/api/v1/" + service.GetName() + "/*",
			ServiceName: service.GetName(),
			Method:      "ANY",
		})
		if err != nil {
			gs.logger.Error("Failed to create route", 
				zap.String("service", service.GetName()), 
				zap.Error(err))
			continue
		}
		
		gs.gateway.AddRoute(route)
	}
}

// ProxyRequest 代理请求
func (gs *GatewayService) ProxyRequest(serviceName string, req *http.Request) (*http.Response, error) {
	// 检查熔断器
	circuitBreaker, exists := gs.circuitBreakers[serviceName]
	if exists {
		if err := circuitBreaker.CanExecute(); err != nil {
			return nil, err
		}
	}
	
	// 获取服务
	service, err := gs.gateway.GetService(serviceName)
	if err != nil {
		if circuitBreaker != nil {
			circuitBreaker.RecordFailure()
		}
		return nil, err
	}
	
	// 检查服务健康状态
	if !service.IsHealthy() {
		if circuitBreaker != nil {
			circuitBreaker.RecordFailure()
		}
		return nil, gs.createServiceUnavailableResponse()
	}
	
	// 执行请求（这里简化处理，实际需要实现HTTP代理）
	start := time.Now()
	
	// 模拟请求执行
	resp := &http.Response{
		StatusCode: http.StatusOK,
		Header:     make(http.Header),
	}
	resp.Header.Set("X-Proxy-Service", serviceName)
	
	// 记录成功
	if circuitBreaker != nil {
		circuitBreaker.RecordSuccess()
	}
	
	// 记录指标
	duration := time.Since(start)
	gs.recordProxyMetrics(serviceName, resp.StatusCode, duration)
	
	return resp, nil
}

// createServiceUnavailableResponse 创建服务不可用响应
func (gs *GatewayService) createServiceUnavailableResponse() error {
	return &ServiceUnavailableError{
		Message: "Service temporarily unavailable",
	}
}

// ServiceUnavailableError 服务不可用错误
type ServiceUnavailableError struct {
	Message string
}

func (e *ServiceUnavailableError) Error() string {
	return e.Message
}

// GetServiceStatus 获取服务状态
func (gs *GatewayService) GetServiceStatus() map[string]interface{} {
	status := make(map[string]interface{})
	
	for name, service := range gs.gateway.GetAllServices() {
		circuitBreaker := gs.circuitBreakers[name]
		
		serviceStatus := map[string]interface{}{
			"name":       service.GetName(),
			"host":       service.GetHost(),
			"port":       service.GetPort(),
			"healthy":    service.IsHealthy(),
			"last_check": service.GetLastCheck().Format(time.RFC3339),
		}
		
		if circuitBreaker != nil {
			serviceStatus["circuit_breaker"] = map[string]interface{}{
				"state":    circuitBreaker.GetStateName(),
				"failures": circuitBreaker.GetFailureCount(),
			}
		}
		
		status[name] = serviceStatus
	}
	
	return status
}

// GetGatewayInfo 获取网关信息
func (gs *GatewayService) GetGatewayInfo() map[string]interface{} {
	return map[string]interface{}{
		"name":             gs.gateway.GetName(),
		"version":          gs.gateway.GetVersion(),
		"status":           string(gs.gateway.GetStatus()),
		"services":         gs.gateway.GetServiceCount(),
		"healthy_services": gs.gateway.GetHealthyServiceCount(),
		"proxy_mode":       "reverse_proxy",
		"load_balancer":    string(gs.loadBalancer.GetStrategy()),
		"created_at":       gs.gateway.GetCreatedAt().Format(time.RFC3339),
		"updated_at":       gs.gateway.GetUpdatedAt().Format(time.RFC3339),
	}
}

// StartHealthChecker 启动健康检查
func (gs *GatewayService) StartHealthChecker() {
	gs.healthTicker = time.NewTicker(30 * time.Second)
	
	go func() {
		for {
			select {
			case <-gs.healthTicker.C:
				gs.performHealthCheck()
			case <-gs.stopHealthCheck:
				return
			}
		}
	}()
	
	gs.logger.Info("Health checker started")
}

// StopHealthChecker 停止健康检查
func (gs *GatewayService) StopHealthChecker() {
	if gs.healthTicker != nil {
		gs.healthTicker.Stop()
	}
	
	select {
	case gs.stopHealthCheck <- true:
	default:
	}
	
	gs.logger.Info("Health checker stopped")
}

// performHealthCheck 执行健康检查
func (gs *GatewayService) performHealthCheck() {
	for _, service := range gs.gateway.GetAllServices() {
		go gs.checkServiceHealth(service)
	}
}

// checkServiceHealth 检查单个服务健康状态
func (gs *GatewayService) checkServiceHealth(service *entity.Service) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	
	healthURL := service.GetHealthCheckURL()
	
	req, err := http.NewRequestWithContext(ctx, "GET", healthURL, nil)
	if err != nil {
		service.UpdateHealth(false)
		return
	}
	
	client := &http.Client{Timeout: 5 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		service.UpdateHealth(false)
		gs.logger.Warn("Service health check failed", 
			zap.String("service", service.GetName()), 
			zap.Error(err))
		return
	}
	defer resp.Body.Close()
	
	healthy := resp.StatusCode == http.StatusOK
	service.UpdateHealth(healthy)
	
	if !healthy {
		gs.logger.Warn("Service unhealthy", 
			zap.String("service", service.GetName()), 
			zap.Int("status_code", resp.StatusCode))
	}
	
	// 更新仓储中的服务状态
	ctx = context.Background()
	if err := gs.serviceRepo.Update(ctx, service); err != nil {
		gs.logger.Error("Failed to update service in repository", 
			zap.String("service", service.GetName()), 
			zap.Error(err))
	}
}

// recordProxyMetrics 记录代理指标
func (gs *GatewayService) recordProxyMetrics(serviceName string, statusCode int, duration time.Duration) {
	if gs.metrics != nil {
		gs.metrics.RecordHTTPRequest("PROXY", "/api/v1/"+serviceName, strconv.Itoa(statusCode), duration)
	}
	
	gs.logger.Debug("Proxy request completed",
		zap.String("service", serviceName),
		zap.Int("status_code", statusCode),
		zap.Duration("duration", duration))
}

// Shutdown 关闭网关
func (gs *GatewayService) Shutdown() {
	gs.StopHealthChecker()
	gs.gateway.Stop()
	gs.gateway.MarkStopped()
	
	gs.logger.Info("Gateway service shutdown completed")
}
