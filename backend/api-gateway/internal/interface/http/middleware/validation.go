package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// ServiceValidation 服务验证中间件（DDD版本）
func ServiceValidation(serviceName string) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 添加服务名到上下文
		c.Set("target_service", serviceName)
		
		// 验证服务名称
		validServices := map[string]bool{
			"agent":        true,
			"llm":          true,
			"mcp":          true,
			"orchestrator": true,
		}

		if !validServices[serviceName] {
			c.JSON(http.StatusBadRequest, gin.H{
				"success":    false,
				"message":    "Invalid service: " + serviceName,
				"error":      "invalid_service",
				"request_id": c.GetString("request_id"),
			})
			c.Abort()
			return
		}

		c.Next()
	}
}

// RequestValidation 请求验证中间件
func RequestValidation() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 检查Content-Type for POST/PUT requests
		if (c.Request.Method == "POST" || c.Request.Method == "PUT" || c.Request.Method == "PATCH") &&
			c.Request.ContentLength > 0 {
			contentType := c.GetHeader("Content-Type")
			if contentType == "" {
				c.JSON(http.StatusBadRequest, gin.H{
					"success":    false,
					"message":    "Missing Content-Type header",
					"error":      "missing_content_type",
					"request_id": c.GetString("request_id"),
				})
				c.Abort()
				return
			}
		}

		c.Next()
	}
}
