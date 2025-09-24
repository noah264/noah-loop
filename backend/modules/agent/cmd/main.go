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

	"github.com/gin-gonic/gin"
	"github.com/noah-loop/backend/modules/agent/internal/wire"
	"github.com/noah-loop/backend/shared/pkg/infrastructure"
	"github.com/noah-loop/backend/shared/pkg/infrastructure/etcd"
	"github.com/noah-loop/backend/shared/pkg/infrastructure/tracing"
	"go.uber.org/zap"
)

const serviceName = "agent-service"

func main() {
	// 使用wire初始化应用
	app, cleanup, err := wire.InitializeAgentApp()
	if err != nil {
		log.Fatalf("Failed to initialize Agent app: %v", err)
	}
	defer cleanup()

	// 初始化基础设施组件
	infraApp, infraCleanup, err := initializeInfrastructure()
	if err != nil {
		log.Fatalf("Failed to initialize infrastructure: %v", err)
	}
	defer infraCleanup()

	app.Logger.Info("Agent service starting with full infrastructure support",
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
	Config          *infrastructure.Config
	Logger          infrastructure.Logger
	TracerManager   *tracing.TracerManager
	EtcdClient      *etcd.Client
	ServiceRegistry *etcd.ServiceRegistry
	ServiceDiscovery *etcd.ServiceDiscovery
	ConfigManager   *etcd.ConfigManager
	SecretManager   *etcd.SecretManager
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
		Config:          config,
		Logger:          logger,
		TracerManager:   tracerManager,
		EtcdClient:      etcdClient,
		ServiceRegistry: serviceRegistry,
		ServiceDiscovery: serviceDiscovery,
		ConfigManager:   configManager,
		SecretManager:   secretManager,
	}

	cleanup := func() {
		tracerManager.Close(context.Background())
		etcdClient.Close()
	}

	return app, cleanup, nil
}

// registerService 注册服务到etcd
func registerService(registry *etcd.ServiceRegistry, config *infrastructure.Config) error {
	serviceInfo := &etcd.ServiceInfo{
		Name:     serviceName,
		Address:  "localhost", // 在生产环境中应该是实际IP
		Port:     config.Services.Agent.Port,
		GRPCPort: config.Services.Agent.GRPCPort,
		Version:  config.App.Version,
		Health:   "healthy",
		Metadata: map[string]string{
			"environment": config.App.Environment,
			"region":      "local",
		},
	}

	return registry.Register(context.Background(), serviceInfo)
}

// deregisterService 注销服务
func deregisterService(registry *etcd.ServiceRegistry) {
	if err := registry.Deregister(context.Background()); err != nil {
		log.Printf("Failed to deregister service: %v", err)
	}
}

// setupHTTPServer 设置HTTP服务器
func setupHTTPServer(app *wire.AgentApp, infraApp *InfrastructureApp) *http.Server {
	// 设置Gin路由
	router := gin.New()
	
	// 添加追踪中间件
	router.Use(tracing.GinTracingMiddleware(infraApp.TracerManager))
	
	// 添加其他中间件
	router.Use(gin.Recovery())
	router.Use(gin.Logger())

	// 设置应用路由
	config := getConfigFromApp(app)
	appRouter := app.Router.SetupRouter(config)
	
	// 挂载应用路由（这里需要根据实际的路由结构调整）
	router.Any("/*path", gin.WrapH(appRouter))

	return &http.Server{
		Addr:         fmt.Sprintf(":%d", infraApp.Config.Services.Agent.Port),
		Handler:      router,
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 30 * time.Second,
		IdleTimeout:  120 * time.Second,
	}
}

// setupGRPCServer 设置gRPC服务器
func setupGRPCServer(app *wire.AgentApp, infraApp *InfrastructureApp) *grpc.Server {
	// 创建gRPC服务器，添加追踪拦截器
	server := grpc.NewServer(
		grpc.UnaryInterceptor(tracing.UnaryServerInterceptor(infraApp.TracerManager)),
		grpc.StreamInterceptor(tracing.StreamServerInterceptor(infraApp.TracerManager)),
	)

	// 注册健康检查服务
	healthServer := health.NewServer()
	grpc_health_v1.RegisterHealthServer(server, healthServer)

	// 设置服务健康状态
	healthServer.SetServingStatus(serviceName, grpc_health_v1.HealthCheckResponse_SERVING)
	healthServer.SetServingStatus("", grpc_health_v1.HealthCheckResponse_SERVING)

	// TODO: 注册Agent gRPC服务
	// agentpb.RegisterAgentServiceServer(server, app.GRPCHandler)

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
		logger.Fatal("HTTP server failed", zap.Error(err))
	}
}

// startGRPCServer 启动gRPC服务器
func startGRPCServer(server *grpc.Server, config *infrastructure.Config, logger infrastructure.Logger) {
	addr := fmt.Sprintf(":%d", config.Services.Agent.GRPCPort)
	listener, err := net.Listen("tcp", addr)
	if err != nil {
		logger.Fatal("Failed to listen for gRPC", zap.String("addr", addr), zap.Error(err))
	}

	logger.Info("Starting gRPC server",
		zap.String("addr", addr),
		zap.String("service", serviceName))

	if err := server.Serve(listener); err != nil {
		logger.Fatal("gRPC server failed", zap.Error(err))
	}
}

// startHealthUpdater 启动健康状态更新器
func startHealthUpdater(registry *etcd.ServiceRegistry, logger infrastructure.Logger) {
	ticker := time.NewTicker(10 * time.Second)
	defer ticker.Stop()

	for range ticker.C {
		// 检查服务健康状态
		health := "healthy"
		if !isServiceHealthy() {
			health = "unhealthy"
		}

		// 更新etcd中的健康状态
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		if err := registry.UpdateHealth(ctx, health); err != nil {
			logger.Error("Failed to update health status", zap.Error(err))
		}
		cancel()
	}
}

// isServiceHealthy 检查服务健康状态
func isServiceHealthy() bool {
	// TODO: 实现实际的健康检查逻辑
	// 例如：检查数据库连接、检查依赖服务等
	return true
}

// waitForShutdown 等待关闭信号
func waitForShutdown(httpServer *http.Server, grpcServer *grpc.Server, tracerManager *tracing.TracerManager, logger infrastructure.Logger) {
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.Info("Shutting down Agent service...")

	// 设置关闭超时
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// 优雅关闭HTTP服务器
	if err := httpServer.Shutdown(ctx); err != nil {
		logger.Error("HTTP server forced shutdown", zap.Error(err))
	}

	// 优雅关闭gRPC服务器
	grpcServer.GracefulStop()

	// 关闭追踪管理器
	if err := tracerManager.Close(ctx); err != nil {
		logger.Error("Failed to close tracer manager", zap.Error(err))
	}

	logger.Info("Agent service stopped gracefully")
}

// getConfigFromApp 从应用中获取配置(临时方案)
func getConfigFromApp(app *wire.AgentApp) *infrastructure.Config {
	// TODO: 改进配置获取方式
	config, err := infrastructure.LoadConfig("../../configs")
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}
	return config
}
