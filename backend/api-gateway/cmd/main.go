package main

import (
	"context"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"shared/monitoring"

	"github.com/gin-gonic/gin"
)

func main() {
	serviceName := envOr("SERVICE_NAME", "api-gateway")
	port := envOr("PORT", "8080")
	authURL := envOr("AUTH_SERVICE_URL", "http://localhost:8084")
	marketURL := envOr("NEXTJS_URL", "http://localhost:3000")

	mon, err := monitoring.Init(serviceName)
	if err != nil {
		log.Fatalf("monitoring init: %v", err)
	}
	defer mon.Close()

	r := gin.Default()
	mon.AttachGin(r)
	r.Use(corsMiddleware())

	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok", "service": serviceName})
	})

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

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Fatalf("shutdown: %v", err)
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
		c.Header("Access-Control-Allow-Headers", "Content-Type, Authorization")
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
