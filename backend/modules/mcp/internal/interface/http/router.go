package http

import (
	"time"

	"github.com/gin-gonic/gin"
	"github.com/noah-loop/backend/shared/pkg/infrastructure"
	"github.com/noah-loop/backend/shared/pkg/middleware"
)

// Router 路由结构
type Router struct {
	handler *MCPHandler
	metrics *infrastructure.MetricsRegistry
}

// NewRouter 创建路由实例
func NewRouter(handler *MCPHandler, metrics *infrastructure.MetricsRegistry) *Router {
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
			"service": "mcp",
			"time":    time.Now().UTC(),
		})
	})
}

// setupAPIRoutes 设置API路由
func (r *Router) setupAPIRoutes(router *gin.Engine) {
	api := router.Group("/api/v1")
	r.registerMCPRoutes(api)
}

// setupMetricsRoutes 设置指标路由
func (r *Router) setupMetricsRoutes(router *gin.Engine) {
	router.GET("/metrics", gin.WrapH(r.metrics.Handler()))
}

// registerMCPRoutes 注册MCP相关路由
func (r *Router) registerMCPRoutes(api *gin.RouterGroup) {
	mcp := api.Group("/mcp")

	// 会话管理路由
	sessions := mcp.Group("/sessions")
	{
		sessions.POST("", r.handler.CreateSession)
		sessions.GET("", r.handler.GetSessions)
		sessions.GET("/:id", r.handler.GetSession)
		sessions.PUT("/:id", r.handler.UpdateSession)
		sessions.DELETE("/:id", r.handler.DeleteSession)
		sessions.POST("/:id/extend", r.handler.ExtendSession)
	}

	// 上下文管理路由
	contexts := mcp.Group("/contexts")
	{
		contexts.POST("", r.handler.AddContext)
		contexts.GET("/:id", r.handler.GetContext)
		contexts.PUT("/:id", r.handler.UpdateContext)
		contexts.DELETE("/:id", r.handler.DeleteContext)
	}

	// 会话上下文路由
	sessionContexts := mcp.Group("/sessions/:session_id/contexts")
	{
		sessionContexts.GET("", r.handler.GetSessionContexts)
		sessionContexts.POST("", r.handler.AddContextToSession)
	}

	// 管理操作路由
	management := mcp.Group("/management")
	{
		management.POST("/cleanup-expired", r.handler.CleanupExpiredSessions)
		management.POST("/manage-idle", r.handler.ManageIdleSessions)
	}
}
