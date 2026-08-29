package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"shared/health"
	"shared/metrics"
	"shared/monitoring"
	"shared/redis"
	"websocket-service/internal/handler"
	"websocket-service/internal/hub"
	"websocket-service/internal/subscriber"

	"github.com/gin-gonic/gin"
)

func main() {
	serviceName := envOr("SERVICE_NAME", "websocket-service")
	port := envOr("PORT", "8083")
	pubSubChannel := envOr("MARKET_PUBSUB_CHANNEL", "market:updates")
	requireAuth := os.Getenv("WS_REQUIRE_AUTH") == "true"

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	redisClient, err := redis.Connect(ctx)
	if err != nil {
		log.Fatalf("redis: %v", err)
	}
	defer redis.Close()

	mon, err := monitoring.Init(serviceName)
	if err != nil {
		log.Fatalf("monitoring init: %v", err)
	}
	defer mon.Close()

	metrics.Init(serviceName)

	h := hub.NewHub()
	sub := subscriber.NewRedisSubscriber(redisClient, pubSubChannel, h)
	go sub.RunWithReconnect(ctx)

	wsHandler := handler.NewWSHandler(h, pubSubChannel, requireAuth)

	r := gin.Default()
	mon.AttachGin(r)
	r.Use(metrics.Middleware())

	r.GET("/health", health.AliveHandler(serviceName))
	r.GET("/ready", health.ReadyHandler(serviceName, map[string]health.Checker{
		"redis": redis.Ping,
	}))
	r.GET("/metrics", metrics.Handler())
	r.GET("/ws", wsHandler.Handle)

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

	cancel()

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer shutdownCancel()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		log.Fatalf("shutdown: %v", err)
	}
}

func envOr(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}
