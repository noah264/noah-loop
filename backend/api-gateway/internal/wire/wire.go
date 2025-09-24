//go:build wireinject

package wire

import (
	"github.com/google/wire"
	"github.com/noah-loop/backend/api-gateway/internal/application/service"
	"github.com/noah-loop/backend/api-gateway/internal/domain/repository"
	"github.com/noah-loop/backend/api-gateway/internal/infrastructure/config"
	infraRepo "github.com/noah-loop/backend/api-gateway/internal/infrastructure/repository"
	"github.com/noah-loop/backend/api-gateway/internal/interface/http/handler"
	"github.com/noah-loop/backend/api-gateway/internal/interface/http/router"
	"github.com/noah-loop/backend/shared/pkg/infrastructure"
	"github.com/noah-loop/backend/shared/pkg/infrastructure/etcd"
	"github.com/noah-loop/backend/shared/pkg/infrastructure/tracing"
)

// GatewayApp API网关应用结构
type GatewayApp struct {
	GatewayService      *service.GatewayService
	GRPCProxyService    *service.GRPCProxyService
	Handler             *handler.GatewayHandler
	GRPCHandler         *handler.GRPCGatewayHandler
	Router              *router.Router
	Metrics             *infrastructure.MetricsRegistry
	Config              *infrastructure.Config
	Logger              infrastructure.Logger
	
	// etcd相关组件
	EtcdClient          *etcd.Client
	ServiceRegistry     *etcd.ServiceRegistry
	ServiceDiscovery    *etcd.ServiceDiscovery
	ConfigManager       *etcd.ConfigManager
	SecretManager       *etcd.SecretManager
	
	// 链路追踪组件
	TracerManager       *tracing.TracerManager
	TracingWrapper      *tracing.TracingWrapper
}

// InitializeGatewayApp 初始化API网关应用
func InitializeGatewayApp() (*GatewayApp, func(), error) {
	wire.Build(
		// 基础设施
		infrastructure.InfrastructureProviderSet,
		
		// 网关配置
		GatewayConfigProviderSet,
		
		// 仓储
		GatewayRepositoryProviderSet,
		
		// 应用服务
		GatewayServiceProviderSet,
		
		// gRPC相关
		GRPCProviderSet,
		
		// etcd相关
		infrastructure.EtcdProviderSet,
		
		// 链路追踪相关
		infrastructure.TracingProviderSet,
		
		// HTTP处理器和路由
		GatewayHandlerProviderSet,
		
		// 应用结构
		wire.Struct(new(GatewayApp), "*"),
		
		// 提供服务名称
		wire.Value("api-gateway"),
	)
	
	return &GatewayApp{}, nil, nil
}

// GatewayConfigProviderSet 网关配置提供者集合
var GatewayConfigProviderSet = wire.NewSet(
	config.NewConfigAdapter,
	wire.Bind(new(service.GatewayConfig), new(*config.ConfigAdapter)),
)

// GatewayRepositoryProviderSet 仓储提供者集合
var GatewayRepositoryProviderSet = wire.NewSet(
	infraRepo.NewInMemoryServiceRepository,
	wire.Bind(new(repository.ServiceRepository), new(repository.ServiceRepository)),
)

// GatewayServiceProviderSet 应用服务提供者集合
var GatewayServiceProviderSet = wire.NewSet(
	service.NewGatewayService,
)

// GatewayHandlerProviderSet HTTP处理器提供者集合
var GatewayHandlerProviderSet = wire.NewSet(
	handler.NewGatewayHandler,
	router.NewRouter,
)
