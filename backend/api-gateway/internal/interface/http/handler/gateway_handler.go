package handler

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/noah-loop/backend/api-gateway/internal/application/service"
	"github.com/noah-loop/backend/shared/pkg/infrastructure"
	"go.uber.org/zap"
)

// GatewayHandler API网关HTTP处理器
type GatewayHandler struct {
	gatewayService *service.GatewayService
	logger         infrastructure.Logger
}

// NewGatewayHandler 创建网关处理器
func NewGatewayHandler(gatewayService *service.GatewayService, logger infrastructure.Logger) *GatewayHandler {
	return &GatewayHandler{
		gatewayService: gatewayService,
		logger:         logger,
	}
}

// HealthCheck 健康检查
func (h *GatewayHandler) HealthCheck(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status":    "healthy",
		"service":   "api-gateway",
		"time":      time.Now().UTC(),
		"version":   "1.0.0",
	})
}

// ServiceStatus 服务状态
func (h *GatewayHandler) ServiceStatus(c *gin.Context) {
	status := h.gatewayService.GetServiceStatus()
	
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    status,
	})
}

// GatewayInfo 网关信息
func (h *GatewayHandler) GatewayInfo(c *gin.Context) {
	info := h.gatewayService.GetGatewayInfo()
	
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    info,
	})
}

// ProxyRequest 代理请求到对应服务
func (h *GatewayHandler) ProxyRequest(serviceName string) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 记录开始时间
		start := time.Now()
		
		// 设置响应头
		c.Header("X-Proxy-Service", serviceName)
		c.Header("X-Gateway", "noah-loop-gateway")
		
		// 执行代理请求（简化版本）
		resp, err := h.gatewayService.ProxyRequest(serviceName, c.Request)
		if err != nil {
			h.handleProxyError(c, serviceName, err)
			return
		}
		
		// 设置响应状态码和头部
		for key, values := range resp.Header {
			for _, value := range values {
				c.Header(key, value)
			}
		}
		c.Status(resp.StatusCode)
		
		// 记录处理时间
		duration := time.Since(start)
		h.logger.Debug("Proxy request completed",
			zap.String("service", serviceName),
			zap.Int("status_code", resp.StatusCode),
			zap.Duration("duration", duration))
		
		// 返回成功响应（简化版本）
		c.JSON(resp.StatusCode, gin.H{
			"success":      true,
			"message":      "Request proxied successfully",
			"service":      serviceName,
			"request_id":   c.GetString("request_id"),
		})
	}
}

// handleProxyError 处理代理错误
func (h *GatewayHandler) handleProxyError(c *gin.Context, serviceName string, err error) {
	h.logger.Error("Proxy request failed",
		zap.String("service", serviceName),
		zap.String("path", c.Request.URL.Path),
		zap.String("method", c.Request.Method),
		zap.Error(err))
	
	// 根据错误类型返回不同的HTTP状态码
	switch err.(type) {
	case *service.ServiceUnavailableError:
		c.JSON(http.StatusServiceUnavailable, gin.H{
			"success":    false,
			"message":    err.Error(),
			"error":      "service_unavailable",
			"service":    serviceName,
			"request_id": c.GetString("request_id"),
		})
	default:
		// 检查是否为熔断器错误
		if isCircuitBreakerError(err) {
			c.JSON(http.StatusServiceUnavailable, gin.H{
				"success":    false,
				"message":    "Service temporarily unavailable due to circuit breaker",
				"error":      "circuit_breaker_open",
				"service":    serviceName,
				"request_id": c.GetString("request_id"),
			})
		} else {
			c.JSON(http.StatusBadGateway, gin.H{
				"success":    false,
				"message":    "Gateway error",
				"error":      "proxy_error",
				"service":    serviceName,
				"request_id": c.GetString("request_id"),
			})
		}
	}
}

// isCircuitBreakerError 检查是否为熔断器错误
func isCircuitBreakerError(err error) bool {
	// 检查错误消息中是否包含熔断器关键字
	errMsg := err.Error()
	return contains(errMsg, "CIRCUIT_BREAKER_OPEN") || contains(errMsg, "CIRCUIT_BREAKER_HALF_OPEN_LIMIT")
}

// contains 检查字符串是否包含子字符串
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(substr) == 0 || (len(s) > len(substr) && s[0:len(substr)] == substr) || contains(s[1:], substr))
}
