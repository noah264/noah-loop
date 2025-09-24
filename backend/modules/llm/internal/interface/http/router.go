package http

import (
	"fmt"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/noah-loop/backend/shared/pkg/infrastructure"
	"github.com/noah-loop/backend/shared/pkg/middleware"
)

// Router 路由结构
type Router struct {
	handler *LLMHandler
	metrics *infrastructure.MetricsRegistry
}

// NewRouter 创建路由实例
func NewRouter(handler *LLMHandler, metrics *infrastructure.MetricsRegistry) *Router {
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
			"service": "llm",
			"time":    time.Now().UTC(),
		})
	})
}

// setupAPIRoutes 设置API路由
func (r *Router) setupAPIRoutes(router *gin.Engine) {
	api := router.Group("/api/v1")
	r.registerLLMRoutes(api)
}

// setupMetricsRoutes 设置指标路由
func (r *Router) setupMetricsRoutes(router *gin.Engine) {
	router.GET("/metrics", gin.WrapH(r.metrics.Handler()))
}

// registerLLMRoutes 注册LLM相关路由
func (r *Router) registerLLMRoutes(api *gin.RouterGroup) {
	llm := api.Group("/llm")

	// 模型管理路由
	models := llm.Group("/models")
	{
		models.POST("", r.handler.CreateModel)
		models.GET("", r.handler.GetModels)
		models.GET("/:id", r.handler.GetModel)
		models.PUT("/:id", r.handler.UpdateModel)
		models.DELETE("/:id", r.handler.DeleteModel)
	}

	// 请求处理路由
	requests := llm.Group("/requests")
	{
		requests.POST("", r.handler.ProcessRequest)
		requests.GET("", r.handler.GetRequests)
		requests.GET("/:id", r.handler.GetRequest)
	}

	// 统计信息路由
	stats := llm.Group("/stats")
	{
		stats.GET("/usage/:user_id", r.handler.GetUsageStats)
	}
}
