package http

import (
	"time"

	"github.com/gin-gonic/gin"
	"github.com/noah-loop/backend/shared/pkg/infrastructure"
	"github.com/noah-loop/backend/shared/pkg/middleware"
)

// Router 路由结构
type Router struct {
	handler *OrchestratorHandler
	metrics *infrastructure.MetricsRegistry
}

// NewRouter 创建路由实例
func NewRouter(handler *OrchestratorHandler, metrics *infrastructure.MetricsRegistry) *Router {
	return &Router{
		handler: handler,
		metrics: metrics,
	}
}

// SetupRouter 设置路由
func (r *Router) SetupRouter(config *infrastructure.Config) *gin.Engine {
	// 设置Gin模式
	if config.App.Environment == "production" {
		gin.SetMode(gin.ReleaseMode)
	}

	// 创建路由引擎
	router := gin.New()

	// 注册中间件
	r.setupMiddlewares(router)

	// 注册健康检查路由
	r.setupHealthRoutes(router)

	// 注册API路由
	r.setupAPIRoutes(router)

	// 注册指标端点
	r.setupMetricsRoutes(router)

	return router
}

// setupMiddlewares 设置中间件
func (r *Router) setupMiddlewares(router *gin.Engine) {
	router.Use(middleware.RequestLogging())
	router.Use(middleware.Recovery())
	router.Use(middleware.CORS(middleware.DefaultCORSConfig()))
	router.Use(middleware.MetricsMiddleware(r.metrics))
}

// setupHealthRoutes 设置健康检查路由
func (r *Router) setupHealthRoutes(router *gin.Engine) {
	router.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"status":  "healthy",
			"service": "orchestrator",
			"time":    time.Now().UTC(),
		})
	})
}

// setupAPIRoutes 设置API路由
func (r *Router) setupAPIRoutes(router *gin.Engine) {
	api := router.Group("/api/v1")
	r.registerOrchestratorRoutes(api)
}

// setupMetricsRoutes 设置指标路由
func (r *Router) setupMetricsRoutes(router *gin.Engine) {
	router.GET("/metrics", gin.WrapH(r.metrics.Handler()))
}

// registerOrchestratorRoutes 注册编排器相关路由
func (r *Router) registerOrchestratorRoutes(api *gin.RouterGroup) {
	orchestrator := api.Group("/orchestrator")

	// 工作流管理路由
	workflows := orchestrator.Group("/workflows")
	{
		workflows.POST("", r.handler.CreateWorkflow)
		workflows.GET("", r.handler.GetWorkflows)
		workflows.GET("/:id", r.handler.GetWorkflow)
		workflows.PUT("/:id", r.handler.UpdateWorkflow)
		workflows.DELETE("/:id", r.handler.DeleteWorkflow)
		workflows.POST("/:id/execute", r.handler.ExecuteWorkflow)
	}

	// 触发器管理路由
	triggers := orchestrator.Group("/triggers")
	{
		triggers.POST("", r.handler.CreateTrigger)
		triggers.GET("", r.handler.GetTriggers)
	}

	// 执行历史路由
	executions := orchestrator.Group("/executions")
	{
		executions.GET("", r.handler.GetExecutions)
		executions.GET("/:id", r.handler.GetExecution)
	}
}
