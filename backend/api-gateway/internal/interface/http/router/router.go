package router

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/noah-loop/backend/api-gateway/internal/application/service"
	"github.com/noah-loop/backend/api-gateway/internal/interface/http/handler"
	"github.com/noah-loop/backend/api-gateway/internal/interface/http/middleware"
	"github.com/noah-loop/backend/shared/pkg/infrastructure"
	sharedMiddleware "github.com/noah-loop/backend/shared/pkg/middleware"
)

// Router API网关路由器
type Router struct {
	gatewayService *service.GatewayService
	handler        *handler.GatewayHandler
	logger         infrastructure.Logger
	metrics        *infrastructure.MetricsRegistry
}

// NewRouter 创建路由器实例
func NewRouter(gatewayService *service.GatewayService, logger infrastructure.Logger, metrics *infrastructure.MetricsRegistry) *Router {
	handler := handler.NewGatewayHandler(gatewayService, logger)
	
	return &Router{
		gatewayService: gatewayService,
		handler:        handler,
		logger:         logger,
		metrics:        metrics,
	}
}

// SetupRouter 设置路由
func (r *Router) SetupRouter() *gin.Engine {
	// 设置Gin模式
	gin.SetMode(gin.ReleaseMode)

	// 创建路由引擎
	router := gin.New()

	// 注册全局中间件
	r.setupGlobalMiddlewares(router)

	// 注册健康检查路由
	r.setupHealthRoutes(router)

	// 注册管理路由
	r.setupManagementRoutes(router)

	// 注册API路由（代理路由）
	r.setupProxyRoutes(router)

	// 注册指标端点
	r.setupMetricsRoutes(router)

	return router
}

// setupGlobalMiddlewares 设置全局中间件
func (r *Router) setupGlobalMiddlewares(router *gin.Engine) {
	// 基础中间件
	router.Use(sharedMiddleware.RequestLogging())
	router.Use(sharedMiddleware.Recovery())
	router.Use(sharedMiddleware.CORS(sharedMiddleware.DefaultCORSConfig()))
	
	if r.metrics != nil {
		router.Use(sharedMiddleware.MetricsMiddleware(r.metrics))
	}
	
	// API网关专用中间件
	router.Use(middleware.RateLimiter())
	router.Use(middleware.Timeout(30 * time.Second))
	router.Use(middleware.CircuitBreaker())
}

// setupHealthRoutes 设置健康检查路由
func (r *Router) setupHealthRoutes(router *gin.Engine) {
	router.GET("/health", r.handler.HealthCheck)
	router.GET("/health/services", r.handler.ServiceStatus)
}

// setupManagementRoutes 设置管理路由
func (r *Router) setupManagementRoutes(router *gin.Engine) {
	management := router.Group("/management")
	{
		management.GET("/info", r.handler.GatewayInfo)
		management.GET("/services", r.handler.ServiceStatus)
	}
}

// setupProxyRoutes 设置代理路由
func (r *Router) setupProxyRoutes(router *gin.Engine) {
	api := router.Group("/api/v1")
	
	// 可选的认证中间件（目前注释掉，后续可以启用）
	// api.Use(middleware.Authentication())

	// Agent服务路由代理
	agentGroup := api.Group("/agent")
	agentGroup.Use(middleware.ServiceValidation("agent"))
	agentGroup.Any("/*path", r.handler.ProxyRequest("agent"))

	// LLM服务路由代理
	llmGroup := api.Group("/llm")  
	llmGroup.Use(middleware.ServiceValidation("llm"))
	llmGroup.Any("/*path", r.handler.ProxyRequest("llm"))

	// MCP服务路由代理
	mcpGroup := api.Group("/mcp")
	mcpGroup.Use(middleware.ServiceValidation("mcp"))
	mcpGroup.Any("/*path", r.handler.ProxyRequest("mcp"))

	// Orchestrator服务路由代理
	orchestratorGroup := api.Group("/orchestrator")
	orchestratorGroup.Use(middleware.ServiceValidation("orchestrator"))
	orchestratorGroup.Any("/*path", r.handler.ProxyRequest("orchestrator"))

	// RAG服务路由代理
	ragGroup := api.Group("/rag")
	ragGroup.Use(middleware.ServiceValidation("rag"))
	ragGroup.Any("/*path", r.handler.ProxyRequest("rag"))

	// Notify服务路由代理
	notifyGroup := api.Group("/notify")
	notifyGroup.Use(middleware.ServiceValidation("notify"))
	notifyGroup.Any("/*path", r.handler.ProxyRequest("notify"))
}

// setupMetricsRoutes 设置指标路由
func (r *Router) setupMetricsRoutes(router *gin.Engine) {
	if r.metrics != nil {
		router.GET("/metrics", gin.WrapH(r.metrics.Handler()))
	}
}
