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
	"auth-service/internal/repository"
	"auth-service/internal/service"
	"shared/monitoring"

	"github.com/gin-gonic/gin"
)

func main() {
	serviceName := envOr("SERVICE_NAME", "auth-service")
	port := envOr("PORT", "8084")

	if os.Getenv("JWT_SECRET") == "" {
		log.Fatal("JWT_SECRET is required")
	}

	ctx := context.Background()
	pool, err := database.Connect(ctx)
	if err != nil {
		log.Fatalf("database: %v", err)
	}
	defer database.Close()

	mon, err := monitoring.Init(serviceName)
	if err != nil {
		log.Fatalf("monitoring init: %v", err)
	}
	defer mon.Close()

	userRepo := repository.NewUserRepository(pool)
	userService := service.NewUserService(userRepo)
	authHandler := handler.NewAuthHandler(userService)

	r := gin.Default()
	mon.AttachGin(r)

	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok", "service": serviceName})
	})

	r.POST("/register", authHandler.Register)
	r.POST("/login", authHandler.Login)

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
