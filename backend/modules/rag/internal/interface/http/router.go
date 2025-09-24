package http

import (
	"github.com/gin-gonic/gin"
	"github.com/noah-loop/backend/modules/rag/internal/interface/http/handler"
	"github.com/noah-loop/backend/shared/pkg/infrastructure"
	"github.com/noah-loop/backend/shared/pkg/infrastructure/tracing"
)

// Router HTTP路由器
type Router struct {
	engine     *gin.Engine
	ragHandler *handler.RAGHandler
	metrics    *infrastructure.MetricsRegistry
}

// NewRouter 创建HTTP路由器
func NewRouter(
	ragHandler *handler.RAGHandler,
	metrics *infrastructure.MetricsRegistry,
	tracingWrapper *tracing.TracingWrapper,
) *Router {
	gin.SetMode(gin.ReleaseMode)
	engine := gin.New()

	// 中间件
	engine.Use(gin.Logger())
	engine.Use(gin.Recovery())

	// 添加链路追踪中间件
	if tracingWrapper != nil {
		engine.Use(tracingWrapper.HTTPMiddleware())
	}

	// 添加指标中间件
	if metrics != nil {
		engine.Use(metrics.PrometheusMiddleware())
	}

	router := &Router{
		engine:     engine,
		ragHandler: ragHandler,
		metrics:    metrics,
	}

	router.setupRoutes()

	return router
}

// setupRoutes 设置路由
func (r *Router) setupRoutes() {
	// 健康检查
	r.engine.GET("/health", r.ragHandler.Health)
	r.engine.GET("/ready", r.ragHandler.Health)

	// API版本
	v1 := r.engine.Group("/api/v1")

	// 知识库相关路由
	kbRoutes := v1.Group("/knowledge-bases")
	{
		kbRoutes.POST("", r.ragHandler.CreateKnowledgeBase)
		kbRoutes.GET("", r.ragHandler.ListKnowledgeBases)
		kbRoutes.GET("/:id", r.ragHandler.GetKnowledgeBase)
		kbRoutes.PUT("/:id", r.ragHandler.UpdateKnowledgeBase)
		kbRoutes.DELETE("/:id", r.ragHandler.DeleteKnowledgeBase)
	}

	// 文档相关路由
	docRoutes := v1.Group("/documents")
	{
		docRoutes.POST("", r.ragHandler.AddDocument)
		docRoutes.GET("", r.ragHandler.ListDocuments)
		docRoutes.GET("/:id", r.ragHandler.GetDocument)
		docRoutes.PUT("/:id", r.ragHandler.UpdateDocument)
		docRoutes.DELETE("/:id", r.ragHandler.DeleteDocument)
		docRoutes.POST("/:id/process", r.ragHandler.ProcessDocument)
		
		// 批量操作
		docRoutes.POST("/batch", r.ragHandler.BatchAddDocuments)
		docRoutes.DELETE("/batch", r.ragHandler.BatchDeleteDocuments)
	}

	// 搜索路由
	searchRoutes := v1.Group("/search")
	{
		searchRoutes.POST("", r.ragHandler.Search)
	}

	// 指标路由（如果启用）
	if r.metrics != nil {
		r.engine.GET("/metrics", gin.WrapH(r.metrics.Handler()))
	}
}

// GetEngine 获取Gin引擎
func (r *Router) GetEngine() *gin.Engine {
	return r.engine
}
