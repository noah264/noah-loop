package middleware

import (
	"context"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

// Timeout 超时中间件（DDD版本）
func Timeout(timeout time.Duration) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 创建带超时的上下文
		ctx, cancel := context.WithTimeout(c.Request.Context(), timeout)
		defer cancel()

		// 替换请求上下文
		c.Request = c.Request.WithContext(ctx)

		// 用于检测是否超时的通道
		finished := make(chan struct{})
		
		// 启动协程处理请求
		go func() {
			defer close(finished)
			c.Next()
		}()

		// 等待请求完成或超时
		select {
		case <-finished:
			// 请求正常完成
			return
		case <-ctx.Done():
			// 请求超时
			if ctx.Err() == context.DeadlineExceeded {
				c.JSON(http.StatusRequestTimeout, gin.H{
					"success":    false,
					"message":    "Request timeout",
					"error":      "request_timeout",
					"timeout":    timeout.String(),
					"request_id": c.GetString("request_id"),
				})
				c.Abort()
			}
			return
		}
	}
}
