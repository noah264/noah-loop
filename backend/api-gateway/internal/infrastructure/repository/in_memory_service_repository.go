package repository

import (
	"context"
	"sync"

	"github.com/noah-loop/backend/api-gateway/internal/domain/entity"
	"github.com/noah-loop/backend/api-gateway/internal/domain/repository"
	"github.com/noah-loop/backend/shared/pkg/domain"
)

// InMemoryServiceRepository 内存服务仓储实现
type InMemoryServiceRepository struct {
	services map[string]*entity.Service
	mutex    sync.RWMutex
}

// NewInMemoryServiceRepository 创建内存服务仓储
func NewInMemoryServiceRepository() repository.ServiceRepository {
	return &InMemoryServiceRepository{
		services: make(map[string]*entity.Service),
	}
}

// Save 保存服务
func (r *InMemoryServiceRepository) Save(ctx context.Context, service *entity.Service) error {
	r.mutex.Lock()
	defer r.mutex.Unlock()
	
	r.services[service.GetName()] = service.Clone()
	return nil
}

// FindByName 根据名称查找服务
func (r *InMemoryServiceRepository) FindByName(ctx context.Context, name string) (*entity.Service, error) {
	r.mutex.RLock()
	defer r.mutex.RUnlock()
	
	service, exists := r.services[name]
	if !exists {
		return nil, domain.NewDomainError("SERVICE_NOT_FOUND", "Service not found: "+name)
	}
	
	return service.Clone(), nil
}

// FindAll 查找所有服务
func (r *InMemoryServiceRepository) FindAll(ctx context.Context) ([]*entity.Service, error) {
	r.mutex.RLock()
	defer r.mutex.RUnlock()
	
	services := make([]*entity.Service, 0, len(r.services))
	for _, service := range r.services {
		services = append(services, service.Clone())
	}
	
	return services, nil
}

// FindHealthy 查找健康的服务
func (r *InMemoryServiceRepository) FindHealthy(ctx context.Context) ([]*entity.Service, error) {
	r.mutex.RLock()
	defer r.mutex.RUnlock()
	
	var healthyServices []*entity.Service
	for _, service := range r.services {
		if service.IsHealthy() {
			healthyServices = append(healthyServices, service.Clone())
		}
	}
	
	return healthyServices, nil
}

// Update 更新服务
func (r *InMemoryServiceRepository) Update(ctx context.Context, service *entity.Service) error {
	r.mutex.Lock()
	defer r.mutex.Unlock()
	
	if _, exists := r.services[service.GetName()]; !exists {
		return domain.NewDomainError("SERVICE_NOT_FOUND", "Service not found: "+service.GetName())
	}
	
	r.services[service.GetName()] = service.Clone()
	return nil
}

// Delete 删除服务
func (r *InMemoryServiceRepository) Delete(ctx context.Context, name string) error {
	r.mutex.Lock()
	defer r.mutex.Unlock()
	
	if _, exists := r.services[name]; !exists {
		return domain.NewDomainError("SERVICE_NOT_FOUND", "Service not found: "+name)
	}
	
	delete(r.services, name)
	return nil
}

// Exists 检查服务是否存在
func (r *InMemoryServiceRepository) Exists(ctx context.Context, name string) (bool, error) {
	r.mutex.RLock()
	defer r.mutex.RUnlock()
	
	_, exists := r.services[name]
	return exists, nil
}
