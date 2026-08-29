package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"auth-service/internal/database"
	"auth-service/internal/handler"
	"auth-service/internal/middleware"
	"auth-service/internal/repository"
	"auth-service/internal/service"
	"shared/health"
	"shared/metrics"
	"shared/monitoring"
	"shared/redis"

	"github.com/gin-gonic/gin"
)

func main() {
	serviceName := envOr("SERVICE_NAME", "auth-service")
	port := envOr("PORT", "8084")

	if os.Getenv("JWT_SECRET") == "" {
		log.Fatal("JWT_SECRET is required")
	}

	ctx := context.Background()

	db, err := database.Connect(ctx)
	if err != nil {
		log.Fatalf("database: %v", err)
	}
	defer database.Close()

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

	userRepo := repository.NewUserRepository(db)
	sessionRepo := repository.NewSessionRepository(redisClient)
	userService := service.NewUserService(userRepo, sessionRepo)
	authHandler := handler.NewAuthHandler(userService)

	r := gin.Default()
	mon.AttachGin(r)
	r.Use(metrics.Middleware())

	r.GET("/health", health.AliveHandler(serviceName))
	r.GET("/ready", health.ReadyHandler(serviceName, map[string]health.Checker{
		"postgres": func(ctx context.Context) error {
			sqlDB, err := db.DB()
			if err != nil {
				return err
			}
			return sqlDB.PingContext(ctx)
		},
		"redis": redis.Ping,
	}))
	r.GET("/metrics", metrics.Handler())

	r.POST("/register", authHandler.Register)
	r.POST("/signup", authHandler.Signup)
	r.POST("/login", authHandler.Login)
	r.POST("/refresh", authHandler.Refresh)
	r.POST("/logout", authHandler.Logout)
	r.GET("/me", middleware.AuthRequired(), authHandler.Me)

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

func envOr(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}
