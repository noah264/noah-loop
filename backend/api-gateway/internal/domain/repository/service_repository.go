package repository

import (
	"context"

	"github.com/noah-loop/backend/api-gateway/internal/domain/entity"
)

// ServiceRepository 服务仓储接口
type ServiceRepository interface {
	// Save 保存服务
	Save(ctx context.Context, service *entity.Service) error
	
	// FindByName 根据名称查找服务
	FindByName(ctx context.Context, name string) (*entity.Service, error)
	
	// FindAll 查找所有服务
	FindAll(ctx context.Context) ([]*entity.Service, error)
	
	// FindHealthy 查找健康的服务
	FindHealthy(ctx context.Context) ([]*entity.Service, error)
	
	// Update 更新服务
	Update(ctx context.Context, service *entity.Service) error
	
	// Delete 删除服务
	Delete(ctx context.Context, name string) error
	
	// Exists 检查服务是否存在
	Exists(ctx context.Context, name string) (bool, error)
}

// GatewayRepository 网关仓储接口
type GatewayRepository interface {
	// Save 保存网关
	Save(ctx context.Context, gateway *entity.Gateway) error
	
	// FindByID 根据ID查找网关
	FindByID(ctx context.Context, id string) (*entity.Gateway, error)
	
	// Update 更新网关
	Update(ctx context.Context, gateway *entity.Gateway) error
	
	// Delete 删除网关
	Delete(ctx context.Context, id string) error
}
