package health

import (
	"context"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

// Checker validates a dependency for readiness probes.
type Checker func(ctx context.Context) error

// ReadyHandler returns a Gin handler that runs dependency checkers with a short timeout.
func ReadyHandler(serviceName string, checkers map[string]Checker) gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(c.Request.Context(), 2*time.Second)
		defer cancel()

		failed := make(map[string]string)
		for name, check := range checkers {
			if check == nil {
				continue
			}
			if err := check(ctx); err != nil {
				failed[name] = err.Error()
			}
		}

		if len(failed) > 0 {
			c.JSON(http.StatusServiceUnavailable, gin.H{
				"status":  "not_ready",
				"service": serviceName,
				"errors":  failed,
			})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"status":  "ready",
			"service": serviceName,
		})
	}
}

// AliveHandler returns a simple liveness response.
func AliveHandler(serviceName string) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok", "service": serviceName})
	}
}
