package entity

import (
	"sync"
	"time"

	"github.com/noah-loop/backend/shared/pkg/domain"
)

// Service 上游服务实体
type Service struct {
	domain.Entity
	name        string
	host        string
	port        int
	path        string
	weight      int
	healthy     bool
	lastCheck   time.Time
	createdAt   time.Time
	updatedAt   time.Time
	metadata    map[string]string
	mutex       sync.RWMutex
}

// ServiceConfig 服务配置
type ServiceConfig struct {
	Name     string
	Host     string
	Port     int
	Path     string
	Weight   int
	Metadata map[string]string
}

// NewService 创建服务实例
func NewService(config ServiceConfig) *Service {
	if config.Weight <= 0 {
		config.Weight = 1
	}
	
	if config.Metadata == nil {
		config.Metadata = make(map[string]string)
	}
	
	service := &Service{
		Entity:    domain.NewEntity(),
		name:      config.Name,
		host:      config.Host,
		port:      config.Port,
		path:      config.Path,
		weight:    config.Weight,
		healthy:   false, // 初始状态为不健康，需要通过健康检查
		createdAt: time.Now(),
		updatedAt: time.Now(),
		metadata:  config.Metadata,
	}
	
	return service
}

// UpdateHealth 更新健康状态
func (s *Service) UpdateHealth(healthy bool) {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	
	oldHealthy := s.healthy
	s.healthy = healthy
	s.lastCheck = time.Now()
	s.updatedAt = time.Now()
	
	// 如果健康状态发生变化，记录日志或发送事件
	if oldHealthy != healthy {
		status := "unhealthy"
		if healthy {
			status = "healthy"
		}
		
		// 这里可以发布领域事件
		// s.PublishEvent(domain.NewDomainEvent("service.health_changed", ...))
	}
}

// IsHealthy 检查是否健康
func (s *Service) IsHealthy() bool {
	s.mutex.RLock()
	defer s.mutex.RUnlock()
	
	return s.healthy
}

// GetName 获取服务名称
func (s *Service) GetName() string {
	return s.name
}

// GetHost 获取主机地址
func (s *Service) GetHost() string {
	return s.host
}

// GetPort 获取端口
func (s *Service) GetPort() int {
	return s.port
}

// GetPath 获取路径
func (s *Service) GetPath() string {
	return s.path
}

// GetWeight 获取权重
func (s *Service) GetWeight() int {
	s.mutex.RLock()
	defer s.mutex.RUnlock()
	
	return s.weight
}

// SetWeight 设置权重
func (s *Service) SetWeight(weight int) {
	if weight <= 0 {
		weight = 1
	}
	
	s.mutex.Lock()
	defer s.mutex.Unlock()
	
	s.weight = weight
	s.updatedAt = time.Now()
}

// GetLastCheck 获取上次检查时间
func (s *Service) GetLastCheck() time.Time {
	s.mutex.RLock()
	defer s.mutex.RUnlock()
	
	return s.lastCheck
}

// GetCreatedAt 获取创建时间
func (s *Service) GetCreatedAt() time.Time {
	return s.createdAt
}

// GetUpdatedAt 获取更新时间
func (s *Service) GetUpdatedAt() time.Time {
	s.mutex.RLock()
	defer s.mutex.RUnlock()
	
	return s.updatedAt
}

// GetMetadata 获取元数据
func (s *Service) GetMetadata() map[string]string {
	s.mutex.RLock()
	defer s.mutex.RUnlock()
	
	// 返回副本以避免并发修改
	metadata := make(map[string]string)
	for k, v := range s.metadata {
		metadata[k] = v
	}
	
	return metadata
}

// SetMetadata 设置元数据
func (s *Service) SetMetadata(key, value string) {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	
	s.metadata[key] = value
	s.updatedAt = time.Now()
}

// RemoveMetadata 删除元数据
func (s *Service) RemoveMetadata(key string) {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	
	delete(s.metadata, key)
	s.updatedAt = time.Now()
}

// GetURL 获取完整URL
func (s *Service) GetURL() string {
	return "http://" + s.host + ":" + string(rune(s.port))
}

// GetHealthCheckURL 获取健康检查URL
func (s *Service) GetHealthCheckURL() string {
	return s.GetURL() + "/health"
}

// IsRecentlyChecked 检查是否最近检查过
func (s *Service) IsRecentlyChecked(interval time.Duration) bool {
	s.mutex.RLock()
	defer s.mutex.RUnlock()
	
	return time.Since(s.lastCheck) < interval
}

// ShouldCheck 是否应该进行健康检查
func (s *Service) ShouldCheck(interval time.Duration) bool {
	s.mutex.RLock()
	defer s.mutex.RUnlock()
	
	// 如果从未检查过，或者距离上次检查超过间隔时间，则应该检查
	return s.lastCheck.IsZero() || time.Since(s.lastCheck) >= interval
}

// Clone 克隆服务实例
func (s *Service) Clone() *Service {
	s.mutex.RLock()
	defer s.mutex.RUnlock()
	
	metadata := make(map[string]string)
	for k, v := range s.metadata {
		metadata[k] = v
	}
	
	clone := &Service{
		Entity:    domain.NewEntity(),
		name:      s.name,
		host:      s.host,
		port:      s.port,
		path:      s.path,
		weight:    s.weight,
		healthy:   s.healthy,
		lastCheck: s.lastCheck,
		createdAt: s.createdAt,
		updatedAt: s.updatedAt,
		metadata:  metadata,
	}
	
	// 设置相同的ID
	clone.SetID(s.GetID())
	
	return clone
}
