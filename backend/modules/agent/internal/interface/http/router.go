package http

import (
	"time"

	"github.com/gin-gonic/gin"
	"github.com/noah-loop/backend/shared/pkg/infrastructure"
	"github.com/noah-loop/backend/shared/pkg/middleware"
)

// Router 路由结构
type Router struct {
	handler *AgentHandler
	metrics *infrastructure.MetricsRegistry
}

// NewRouter 创建路由实例
func NewRouter(handler *AgentHandler, metrics *infrastructure.MetricsRegistry) *Router {
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
			"service": "agent",
			"time":    time.Now().UTC(),
		})
	})
}

// setupAPIRoutes 设置API路由
func (r *Router) setupAPIRoutes(router *gin.Engine) {
	api := router.Group("/api/v1")
	r.registerAgentRoutes(api)
}

// setupMetricsRoutes 设置指标路由
func (r *Router) setupMetricsRoutes(router *gin.Engine) {
	router.GET("/metrics", gin.WrapH(r.metrics.Handler()))
}

// registerAgentRoutes 注册智能体相关路由
func (r *Router) registerAgentRoutes(api *gin.RouterGroup) {
	agent := api.Group("/agent")

	// 智能体管理路由
	agents := agent.Group("/agents")
	{
		agents.POST("", r.handler.CreateAgent)
		agents.GET("", r.handler.GetAgents)
		agents.GET("/:id", r.handler.GetAgent)
		agents.PUT("/:id", r.handler.UpdateAgent)
		agents.DELETE("/:id", r.handler.DeleteAgent)
		agents.POST("/:id/chat", r.handler.ChatWithAgent)
		agents.POST("/:id/learn", r.handler.LearnAgent)
	}

	// 工具管理路由
	tools := agent.Group("/tools")
	{
		tools.POST("", r.handler.CreateTool)
		tools.GET("", r.handler.GetTools)
		tools.GET("/:id", r.handler.GetTool)
		tools.PUT("/:id", r.handler.UpdateTool)
		tools.DELETE("/:id", r.handler.DeleteTool)
		tools.POST("/:id/execute", r.handler.ExecuteTool)
	}

	// 工具分配路由
	assign := agent.Group("/assign")
	{
		assign.POST("/tool", r.handler.AssignTool)
		assign.DELETE("/tool", r.handler.UnassignTool)
	}

	// 记忆管理路由
	memory := agent.Group("/memory")
	{
		memory.GET("/search", r.handler.SearchMemory)
		memory.GET("/:agent_id/recent", r.handler.GetRecentMemories)
	}

	// 执行历史路由
	executions := agent.Group("/executions")
	{
		executions.GET("", r.handler.GetExecutions)
		executions.GET("/:id", r.handler.GetExecution)
	}
}
