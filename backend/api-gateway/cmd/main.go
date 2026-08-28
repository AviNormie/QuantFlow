package main

import (
	"context"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"shared/gateway"
	"shared/health"
	"shared/monitoring"
	"shared/redis"

	"github.com/gin-gonic/gin"
)

func main() {
	serviceName := envOr("SERVICE_NAME", "api-gateway")
	port := envOr("PORT", "8080")
	authURL := ensureHTTPURL(envOr("AUTH_SERVICE_URL", "http://localhost:8084"))
	marketURL := ensureHTTPURL(envOr("MARKET_SERVICE_URL", "http://localhost:8082"))

	ctx := context.Background()
	if _, err := redis.Connect(ctx); err != nil {
		log.Printf("redis connect warning: %v", err)
	}
	defer redis.Close()

	mon, err := monitoring.Init(serviceName)
	if err != nil {
		log.Fatalf("monitoring init: %v", err)
	}
	defer mon.Close()

	r := gin.Default()
	mon.AttachGin(r)
	r.Use(gateway.RequestIDMiddleware())
	r.Use(gateway.TimeoutMiddleware(30 * time.Second))
	r.Use(gateway.RateLimitMiddleware())
	r.Use(gateway.ErrorEnvelopeMiddleware())
	r.Use(corsMiddleware())

	r.GET("/health", health.AliveHandler(serviceName))
	r.GET("/ready", health.ReadyHandler(serviceName, map[string]health.Checker{
		"redis": redis.Ping,
		"auth":  downstreamHealthChecker(authURL + "/ready"),
		"market": downstreamHealthChecker(marketURL + "/ready"),
	}))

	r.Any("/api/auth/*proxyPath", makePathProxy(authURL, "/api/auth"))
	r.Any("/api/market/*proxyPath", makePathProxy(marketURL, "/api/market"))

	srv := &http.Server{
		Addr:    ":" + port,
		Handler: r,
	}

	go func() {
		log.Printf("%s listening on :%s", serviceName, port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("server: %v", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		log.Fatalf("shutdown: %v", err)
	}
}

func downstreamHealthChecker(url string) health.Checker {
	return func(ctx context.Context) error {
		req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
		if err != nil {
			return err
		}
		res, err := http.DefaultClient.Do(req)
		if err != nil {
			return err
		}
		defer res.Body.Close()
		io.Copy(io.Discard, res.Body)
		if res.StatusCode >= 400 {
			return fmt.Errorf("status %d", res.StatusCode)
		}
		return nil
	}
}

func makePathProxy(targetBase, stripPrefix string) gin.HandlerFunc {
	target, err := url.Parse(targetBase)
	if err != nil {
		log.Fatalf("invalid proxy target %q: %v", targetBase, err)
	}

	proxy := httputil.NewSingleHostReverseProxy(target)
	originalDirector := proxy.Director
	proxy.Director = func(req *http.Request) {
		originalDirector(req)
		path := strings.TrimPrefix(req.URL.Path, stripPrefix)
		if path == "" {
			path = "/"
		}
		req.URL.Path = path
		req.URL.RawPath = ""
	}

	return func(c *gin.Context) {
		if requestID := c.GetString("request_id"); requestID != "" {
			c.Request.Header.Set("X-Request-ID", requestID)
		}
		proxy.ServeHTTP(c.Writer, c.Request)
	}
}

func corsMiddleware() gin.HandlerFunc {
	allowed := parseAllowedOrigins(envOr("ALLOWED_ORIGINS", "http://localhost:3000"))

	return func(c *gin.Context) {
		origin := c.GetHeader("Origin")
		if origin != "" && isAllowedOrigin(origin, allowed) {
			c.Header("Access-Control-Allow-Origin", origin)
			c.Header("Access-Control-Allow-Credentials", "true")
		}
		c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, PATCH, DELETE, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Content-Type, Authorization, X-Request-ID")
		c.Header("Vary", "Origin")

		if c.Request.Method == http.MethodOptions {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}

		c.Next()
	}
}

func parseAllowedOrigins(value string) []string {
	parts := strings.Split(value, ",")
	origins := make([]string, 0, len(parts))
	for _, part := range parts {
		if trimmed := strings.TrimSpace(part); trimmed != "" {
			origins = append(origins, trimmed)
		}
	}
	return origins
}

func isAllowedOrigin(origin string, allowed []string) bool {
	for _, pattern := range allowed {
		if pattern == "*" || pattern == origin {
			return true
		}
		if strings.HasPrefix(pattern, "https://*.") {
			suffix := strings.TrimPrefix(pattern, "https://*.")
			if strings.HasPrefix(origin, "https://") && strings.HasSuffix(origin, "."+suffix) {
				return true
			}
		}
	}
	return false
}

func envOr(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}

func ensureHTTPURL(raw string) string {
	if strings.HasPrefix(raw, "http://") || strings.HasPrefix(raw, "https://") {
		return raw
	}
	return "http://" + raw
}
