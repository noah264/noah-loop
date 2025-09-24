package wire

import (
	"github.com/google/wire"
	"github.com/noah-loop/backend/api-gateway/internal/application/service"
	grpcClients "github.com/noah-loop/backend/api-gateway/internal/infrastructure/grpc"
	"github.com/noah-loop/backend/api-gateway/internal/interface/http/handler"
	"github.com/noah-loop/backend/shared/pkg/infrastructure"
)

// GRPCProviderSet gRPC相关提供者集合
var GRPCProviderSet = wire.NewSet(
	// gRPC客户端管理器
	grpcClients.NewClientManager,
	
	// gRPC代理服务
	service.NewGRPCProxyService,
	
	// gRPC网关处理器
	handler.NewGRPCGatewayHandler,
)

// ProvideGRPCClientManager 提供gRPC客户端管理器
func ProvideGRPCClientManager(config *infrastructure.Config, logger infrastructure.Logger) (*grpcClients.ClientManager, error) {
	return grpcClients.NewClientManager(config, logger)
}

// ProvideGRPCProxyService 提供gRPC代理服务
func ProvideGRPCProxyService(clientManager *grpcClients.ClientManager, logger infrastructure.Logger) *service.GRPCProxyService {
	return service.NewGRPCProxyService(clientManager, logger)
}

// ProvideGRPCGatewayHandler 提供gRPC网关处理器
func ProvideGRPCGatewayHandler(grpcProxyService *service.GRPCProxyService, logger infrastructure.Logger) *handler.GRPCGatewayHandler {
	return handler.NewGRPCGatewayHandler(grpcProxyService, logger)
}
