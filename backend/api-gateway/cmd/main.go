package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"syscall"

	"github.com/noah-loop/backend/api-gateway/internal/wire"
	"github.com/noah-loop/backend/shared/pkg/infrastructure"
	"go.uber.org/zap"
)

func main() {
	// 使用Wire初始化应用
	app, cleanup, err := wire.InitializeGatewayApp()
	if err != nil {
		log.Fatalf("Failed to initialize Gateway app: %v", err)
	}
	defer cleanup()

	// 设置全局日志
	infrastructure.GlobalLogger = getLoggerFromApp(app)

	app.GatewayService.Initialize()

	// 启动应用
	if err := startGateway(app); err != nil {
		log.Fatalf("Failed to start Gateway: %v", err)
	}
}

// startGateway 启动网关服务
func startGateway(app *wire.GatewayApp) error {
	// 设置路由
	router := app.Router.SetupRouter()

	// 创建HTTP服务器
	srv := &http.Server{
		Addr:         ":" + strconv.Itoa(app.Config.HTTP.Port),
		Handler:      router,
		ReadTimeout:  app.Config.HTTP.ReadTimeout,
		WriteTimeout: app.Config.HTTP.WriteTimeout,
		IdleTimeout:  app.Config.HTTP.IdleTimeout,
	}

	// 启动服务器
	go func() {
		logger := getLoggerFromApp(app)
		logger.Info("API Gateway server starting",
			zap.String("addr", srv.Addr),
			zap.String("environment", app.Config.App.Environment))

		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Fatal("Server failed to start", zap.Error(err))
		}
	}()

	// 启动健康检查
	go app.GatewayService.StartHealthChecker()

	// 等待中断信号进行优雅关闭
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger := getLoggerFromApp(app)
	logger.Info("Shutting down API Gateway...")

	// 停止健康检查
	app.GatewayService.StopHealthChecker()

	// 设置关闭超时
	ctx, cancel := context.WithTimeout(context.Background(), app.Config.HTTP.ShutdownTimeout)
	defer cancel()

	// 优雅关闭服务器
	if err := srv.Shutdown(ctx); err != nil {
		logger.Error("Server forced to shutdown", zap.Error(err))
	} else {
		logger.Info("API Gateway stopped gracefully")
	}

	return nil
}

// getLoggerFromApp 从应用中获取日志器(临时方案)
func getLoggerFromApp(app *wire.GatewayApp) infrastructure.Logger {
	// TODO: 改进日志器获取方式
	logger, err := infrastructure.NewZapLogger(app.Config.Log.Level)
	if err != nil {
		log.Fatalf("Failed to create logger: %v", err)
	}
	return logger
}