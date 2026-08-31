package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"market-service/internal/config"
	"market-service/internal/handler"
	"market-service/internal/provider/finnhub"
	"market-service/internal/repository"
	"market-service/internal/service"
	"shared/health"
	"shared/metrics"
	"shared/monitoring"
	"shared/redis"

	"github.com/gin-gonic/gin"
)

func main() {
	cfg := config.Load()
	if cfg.FinnhubAPIKey == "" {
		log.Println("warning: FINNHUB_API_KEY is not set; REST and ingestion will fail")
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	redisClient, err := redis.Connect(ctx)
	if err != nil {
		log.Fatalf("redis: %v", err)
	}
	defer redis.Close()

	mon, err := monitoring.Init(cfg.ServiceName)
	if err != nil {
		log.Fatalf("monitoring init: %v", err)
	}
	defer mon.Close()

	metrics.Init(cfg.ServiceName)

	provider := finnhub.NewProvider(cfg.FinnhubAPIKey)
	normalizer := service.NewNormalizer()
	priceCache := repository.NewPriceCache(
		redisClient,
		cfg.PriceCachePrefix,
		time.Duration(cfg.PriceCacheTTLHours)*time.Hour,
	)
	publisher := repository.NewPublisher(redisClient, cfg.PubSubChannel)
	ingestion := service.NewIngestionService(cfg, provider, normalizer, priceCache, publisher)
	marketService := service.NewMarketService(provider, priceCache, ingestion)
	marketHandler := handler.NewMarketHandler(marketService)

	go ingestion.Run(ctx)

	r := gin.Default()
	mon.AttachGin(r)
	r.Use(metrics.Middleware())

	r.GET("/health", health.AliveHandler(cfg.ServiceName))
	r.GET("/ready", health.ReadyHandler(cfg.ServiceName, map[string]health.Checker{
		"redis": redis.Ping,
	}))
	r.GET("/metrics", metrics.Handler())

	r.GET("/symbols/search", marketHandler.SearchSymbols)
	r.GET("/symbols/:symbol", marketHandler.ResolveSymbol)
	r.GET("/quotes/:symbol", marketHandler.GetQuote)
	r.GET("/candles/:symbol", marketHandler.GetCandles)

	r.GET("/quote", func(c *gin.Context) {
		symbol := c.Query("symbol")
		if symbol == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "symbol is required"})
			return
		}
		c.Params = append(c.Params, gin.Param{Key: "symbol", Value: symbol})
		marketHandler.GetQuote(c)
	})
	r.GET("/candles", func(c *gin.Context) {
		symbol := c.Query("symbol")
		if symbol == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "symbol is required"})
			return
		}
		c.Params = append(c.Params, gin.Param{Key: "symbol", Value: symbol})
		marketHandler.GetCandles(c)
	})

	srv := &http.Server{
		Addr:    ":" + cfg.Port,
		Handler: r,
	}

	go func() {
		log.Printf("%s listening on :%s", cfg.ServiceName, cfg.Port)
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
