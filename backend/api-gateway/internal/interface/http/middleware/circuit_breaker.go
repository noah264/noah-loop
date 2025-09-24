package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// CircuitBreaker 熔断器中间件（DDD版本）
func CircuitBreaker() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 在DDD架构中，熔断器逻辑已经移到领域服务层
		// 这个中间件主要用于全局熔断保护
		
		// 处理请求
		c.Next()

		// 可以在这里添加全局熔断逻辑
		// 但主要的熔断逻辑在应用服务层处理
	}
}
