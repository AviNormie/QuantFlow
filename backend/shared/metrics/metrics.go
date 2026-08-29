package metrics

import (
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

var (
	serviceName string

	requestsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "http_requests_total",
			Help: "Total HTTP requests processed.",
		},
		[]string{"service", "method", "path", "status"},
	)

	requestDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "http_request_duration_seconds",
			Help:    "HTTP request latency in seconds.",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"service", "method", "path"},
	)
)

// Init registers the service name used as a Prometheus label.
func Init(name string) {
	serviceName = name
}

// Middleware records request counts and latency for each route.
func Middleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		path := c.FullPath()
		if path == "" {
			path = c.Request.URL.Path
		}
		if path == "/metrics" || path == "/health" {
			c.Next()
			return
		}

		start := time.Now()
		c.Next()

		status := strconv.Itoa(c.Writer.Status())
		requestsTotal.WithLabelValues(serviceName, c.Request.Method, path, status).Inc()
		requestDuration.WithLabelValues(serviceName, c.Request.Method, path).Observe(time.Since(start).Seconds())
	}
}

// Handler exposes the Prometheus scrape endpoint.
func Handler() gin.HandlerFunc {
	return gin.WrapH(promhttp.Handler())
}
