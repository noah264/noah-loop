package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/health"
	"google.golang.org/grpc/health/grpc_health_v1"
	"google.golang.org/grpc/reflection"

	"github.com/noah-loop/backend/modules/rag/internal/wire"
	"github.com/noah-loop/backend/shared/pkg/infrastructure"
	"github.com/noah-loop/backend/shared/pkg/infrastructure/etcd"
	"github.com/noah-loop/backend/shared/pkg/infrastructure/tracing"
	"go.uber.org/zap"
)

const serviceName = "rag-service"

func main() {
	// 使用wire初始化应用
	app, cleanup, err := wire.InitializeRAGApp()
	if err != nil {
		log.Fatalf("Failed to initialize RAG app: %v", err)
	}
	defer cleanup()

	// 初始化基础设施组件
	infraApp, infraCleanup, err := initializeInfrastructure()
	if err != nil {
		log.Fatalf("Failed to initialize infrastructure: %v", err)
	}
	defer infraCleanup()

	app.Logger.Info("RAG service starting with full infrastructure support",
		zap.String("service", serviceName),
		zap.String("version", app.Config.App.Version))

	// 注册服务到etcd
	if err := registerService(infraApp.ServiceRegistry, infraApp.Config); err != nil {
		app.Logger.Fatal("Failed to register service", zap.Error(err))
	}
	defer deregisterService(infraApp.ServiceRegistry)

	// 设置HTTP服务器
	httpServer := setupHTTPServer(app, infraApp)

	// 设置gRPC服务器
	grpcServer := setupGRPCServer(app, infraApp)

	// 启动服务器
	go startHTTPServer(httpServer, infraApp.Config, app.Logger)
	go startGRPCServer(grpcServer, infraApp.Config, app.Logger)

	// 启动健康检查更新
	go startHealthUpdater(infraApp.ServiceRegistry, app.Logger)

	// 等待中断信号
	waitForShutdown(httpServer, grpcServer, infraApp.TracerManager, app.Logger)
}

// InfrastructureApp 基础设施应用组件
type InfrastructureApp struct {
	Config           *infrastructure.Config
	Logger           infrastructure.Logger
	TracerManager    *tracing.TracerManager
	EtcdClient       *etcd.Client
	ServiceRegistry  *etcd.ServiceRegistry
	ServiceDiscovery *etcd.ServiceDiscovery
	ConfigManager    *etcd.ConfigManager
	SecretManager    *etcd.SecretManager
}

// initializeInfrastructure 初始化基础设施组件
func initializeInfrastructure() (*InfrastructureApp, func(), error) {
	// 加载配置
	config, err := infrastructure.LoadConfig("../../configs")
	if err != nil {
		return nil, nil, fmt.Errorf("failed to load config: %w", err)
	}

	// 初始化日志
	logger := infrastructure.NewZapLogger(config)

	// 初始化链路追踪
	tracingConfig := tracing.NewTracingConfigFromInfrastructure(config, serviceName)
	tracerManager, err := tracing.NewTracerManager(tracingConfig, logger)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to initialize tracer: %w", err)
	}

	// 初始化etcd客户端
	etcdConfig := etcd.NewConfigFromInfrastructure(config)
	etcdClient, err := etcd.NewClient(etcdConfig, logger)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to initialize etcd client: %w", err)
	}

	// 初始化etcd组件
	serviceRegistry := etcd.NewServiceRegistry(etcdClient, serviceName, logger)
	serviceDiscovery := etcd.NewServiceDiscovery(etcdClient, logger)
	configManager := etcd.NewConfigManager(etcdClient, logger)
	secretManager := etcd.NewSecretManager(etcdClient, logger)

	app := &InfrastructureApp{
		Config:           config,
		Logger:           logger,
		TracerManager:    tracerManager,
		EtcdClient:       etcdClient,
		ServiceRegistry:  serviceRegistry,
		ServiceDiscovery: serviceDiscovery,
		ConfigManager:    configManager,
		SecretManager:    secretManager,
	}

	cleanup := func() {
		tracerManager.Close(context.Background())
		etcdClient.Close()
	}

	return app, cleanup, nil
}

// registerService 注册服务到etcd
func registerService(serviceRegistry *etcd.ServiceRegistry, config *infrastructure.Config) error {
	serviceInfo := etcd.ServiceInfo{
		Name:    serviceName,
		Version: config.App.Version,
		HTTP: etcd.EndpointInfo{
			Host: "localhost",
			Port: config.Services.RAG.Port,
		},
		GRPC: etcd.EndpointInfo{
			Host: "localhost",
			Port: config.Services.RAG.GRPCPort,
		},
		Metadata: map[string]string{
			"environment": config.App.Environment,
			"region":      "local",
		},
	}

	return serviceRegistry.Register(context.Background(), serviceInfo, 30*time.Second)
}

// deregisterService 注销服务
func deregisterService(serviceRegistry *etcd.ServiceRegistry) {
	if err := serviceRegistry.Deregister(context.Background()); err != nil {
		log.Printf("Failed to deregister service: %v", err)
	}
}

// setupHTTPServer 设置HTTP服务器
func setupHTTPServer(app *wire.RAGApp, infraApp *InfrastructureApp) *http.Server {
	return &http.Server{
		Addr:         fmt.Sprintf(":%d", infraApp.Config.Services.RAG.Port),
		Handler:      app.Router.GetEngine(),
		ReadTimeout:  time.Duration(infraApp.Config.HTTP.ReadTimeout) * time.Second,
		WriteTimeout: time.Duration(infraApp.Config.HTTP.WriteTimeout) * time.Second,
		IdleTimeout:  time.Duration(infraApp.Config.HTTP.IdleTimeout) * time.Second,
	}
}

// setupGRPCServer 设置gRPC服务器
func setupGRPCServer(app *wire.RAGApp, infraApp *InfrastructureApp) *grpc.Server {
	// gRPC拦截器选项
	var opts []grpc.ServerOption

	// 添加链路追踪拦截器
	if infraApp.TracerManager != nil {
		if unaryInterceptor := infraApp.TracerManager.UnaryServerInterceptor(); unaryInterceptor != nil {
			opts = append(opts, grpc.UnaryInterceptor(unaryInterceptor))
		}
		if streamInterceptor := infraApp.TracerManager.StreamServerInterceptor(); streamInterceptor != nil {
			opts = append(opts, grpc.StreamInterceptor(streamInterceptor))
		}
	}

	server := grpc.NewServer(opts...)

	// 注册健康检查服务
	healthServer := health.NewServer()
	grpc_health_v1.RegisterHealthServer(server, healthServer)
	healthServer.SetServingStatus(serviceName, grpc_health_v1.HealthCheckResponse_SERVING)

	// TODO: 注册RAG gRPC服务
	// ragpb.RegisterRAGServiceServer(server, grpcHandler)

	// 启用反射（开发环境）
	if infraApp.Config.App.Environment == "development" {
		reflection.Register(server)
	}

	return server
}

// startHTTPServer 启动HTTP服务器
func startHTTPServer(server *http.Server, config *infrastructure.Config, logger infrastructure.Logger) {
	logger.Info("Starting HTTP server", 
		zap.String("addr", server.Addr),
		zap.String("service", serviceName))

	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		logger.Fatal("Failed to start HTTP server", zap.Error(err))
	}
}

// startGRPCServer 启动gRPC服务器
func startGRPCServer(server *grpc.Server, config *infrastructure.Config, logger infrastructure.Logger) {
	addr := fmt.Sprintf(":%d", config.Services.RAG.GRPCPort)
	lis, err := net.Listen("tcp", addr)
	if err != nil {
		logger.Fatal("Failed to listen on gRPC port", zap.Error(err), zap.String("addr", addr))
	}

	logger.Info("Starting gRPC server",
		zap.String("addr", addr),
		zap.String("service", serviceName))

	if err := server.Serve(lis); err != nil {
		logger.Fatal("Failed to start gRPC server", zap.Error(err))
	}
}

// startHealthUpdater 启动健康状态更新器
func startHealthUpdater(serviceRegistry *etcd.ServiceRegistry, logger infrastructure.Logger) {
	ticker := time.NewTicker(15 * time.Second)
	defer ticker.Stop()

	for range ticker.C {
		if err := serviceRegistry.UpdateHealth(context.Background(), etcd.HealthStatusHealthy, ""); err != nil {
			logger.Error("Failed to update health status", zap.Error(err))
		}
	}
}

// waitForShutdown 等待关闭信号
func waitForShutdown(httpServer *http.Server, grpcServer *grpc.Server, tracerManager *tracing.TracerManager, logger infrastructure.Logger) {
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	
	<-quit
	logger.Info("Shutting down RAG service...")

	// 创建超时上下文
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// 优雅关闭HTTP服务器
	if err := httpServer.Shutdown(ctx); err != nil {
		logger.Error("HTTP server forced to shutdown", zap.Error(err))
	}

	// 优雅关闭gRPC服务器
	done := make(chan bool, 1)
	go func() {
		grpcServer.GracefulStop()
		done <- true
	}()

	select {
	case <-done:
		logger.Info("gRPC server stopped gracefully")
	case <-ctx.Done():
		logger.Warn("gRPC server forced to stop due to timeout")
		grpcServer.Stop()
	}

	// 关闭链路追踪
	if tracerManager != nil {
		if err := tracerManager.Close(ctx); err != nil {
			logger.Error("Failed to close tracer", zap.Error(err))
		}
	}

	logger.Info("RAG service stopped")
}
