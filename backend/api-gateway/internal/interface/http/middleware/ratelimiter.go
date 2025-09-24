package middleware

import (
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"golang.org/x/time/rate"
)

// RateLimiter 限流中间件（DDD版本）
func RateLimiter() gin.HandlerFunc {
	// 创建基于IP的限流器映射
	limiters := &sync.Map{}
	
	// 每分钟最多100个请求，突发允许20个请求
	rps := rate.Every(time.Minute / 100)
	burst := 20

	return func(c *gin.Context) {
		ip := c.ClientIP()
		
		// 获取或创建限流器
		limiterInterface, _ := limiters.LoadOrStore(ip, rate.NewLimiter(rps, burst))
		limiter := limiterInterface.(*rate.Limiter)

		if !limiter.Allow() {
			c.JSON(http.StatusTooManyRequests, gin.H{
				"success":    false,
				"message":    "Rate limit exceeded. Please try again later.",
				"error":      "rate_limit_exceeded",
				"request_id": c.GetString("request_id"),
			})
			c.Abort()
			return
		}

		c.Next()
	}
}
