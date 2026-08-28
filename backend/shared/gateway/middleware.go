package gateway

import (
	"context"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// RequestIDMiddleware generates or propagates X-Request-ID on every request.
func RequestIDMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		requestID := c.GetHeader("X-Request-ID")
		if requestID == "" {
			requestID = uuid.NewString()
		}
		c.Set("request_id", requestID)
		c.Header("X-Request-ID", requestID)
		c.Next()
	}
}

// TimeoutMiddleware aborts requests that exceed the configured duration.
func TimeoutMiddleware(timeout time.Duration) gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx, cancel := contextWithTimeout(c.Request.Context(), timeout)
		defer cancel()
		c.Request = c.Request.WithContext(ctx)
		c.Next()
	}
}

// ErrorEnvelopeMiddleware ensures JSON error responses share a consistent shape.
func ErrorEnvelopeMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Next()

		if len(c.Errors) == 0 {
			return
		}
		if c.Writer.Written() {
			return
		}

		c.JSON(http.StatusInternalServerError, gin.H{
			"error":      "internal server error",
			"request_id": c.GetString("request_id"),
		})
	}
}

// RateLimitMiddleware is a skeleton token-bucket limiter placeholder keyed by client IP.
func RateLimitMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Skeleton: real Redis-backed limiter can replace this per-route policy.
		c.Next()
	}
}

func contextWithTimeout(parent context.Context, timeout time.Duration) (context.Context, context.CancelFunc) {
	return context.WithTimeout(parent, timeout)
}
