package http

import (
	"github.com/gin-gonic/gin"
	"github.com/noah-loop/backend/modules/notify/internal/interface/http/handler"
	"github.com/noah-loop/backend/shared/pkg/infrastructure"
	"github.com/noah-loop/backend/shared/pkg/infrastructure/tracing"
)

// Router HTTP路由器
type Router struct {
	engine        *gin.Engine
	notifyHandler *handler.NotifyHandler
	metrics       *infrastructure.MetricsRegistry
}

// NewRouter 创建HTTP路由器
func NewRouter(
	notifyHandler *handler.NotifyHandler,
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
		engine:        engine,
		notifyHandler: notifyHandler,
		metrics:       metrics,
	}

	router.setupRoutes()
	return router
}

// setupRoutes 设置路由
func (r *Router) setupRoutes() {
	// 健康检查
	r.engine.GET("/health", r.notifyHandler.Health)
	r.engine.GET("/ready", r.notifyHandler.Health)

	// API版本
	v1 := r.engine.Group("/api/v1")

	// 通知相关路由
	notifications := v1.Group("/notifications")
	{
		notifications.POST("", r.notifyHandler.CreateNotification)
		notifications.POST("/template", r.notifyHandler.CreateNotificationFromTemplate)
		notifications.GET("", r.notifyHandler.ListNotifications)
		notifications.GET("/:id", r.notifyHandler.GetNotification)
		notifications.POST("/:id/send", r.notifyHandler.SendNotification)
	}

	// 模板相关路由
	templates := v1.Group("/templates")
	{
		templates.POST("", r.notifyHandler.CreateTemplate)
		// templates.GET("", r.notifyHandler.ListTemplates)
		// templates.GET("/:id", r.notifyHandler.GetTemplate)
		// templates.PUT("/:id", r.notifyHandler.UpdateTemplate)
	}

	// 渠道配置相关路由
	channels := v1.Group("/channels")
	{
		channels.POST("", r.notifyHandler.CreateChannelConfig)
		channels.POST("/test", r.notifyHandler.TestChannel)
		// channels.GET("", r.notifyHandler.ListChannelConfigs)
		// channels.GET("/:id", r.notifyHandler.GetChannelConfig)
		// channels.PUT("/:id", r.notifyHandler.UpdateChannelConfig)
	}

	// 指标路由
	if r.metrics != nil {
		r.engine.GET("/metrics", gin.WrapH(r.metrics.Handler()))
	}
}

// GetEngine 获取Gin引擎
func (r *Router) GetEngine() *gin.Engine {
	return r.engine
}
