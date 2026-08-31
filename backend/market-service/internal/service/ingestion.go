package service

import (
	"context"
	"log"
	"strings"
	"sync"
	"time"

	"market-service/internal/config"
	"market-service/internal/provider"
	"market-service/internal/repository"
)

// IngestionService consumes provider trades and fans out to Redis.
type IngestionService struct {
	cfg        config.Config
	provider   provider.MarketProvider
	normalizer *Normalizer
	cache      *repository.PriceCache
	publisher  *repository.Publisher
	mu         sync.Mutex
	subscribed map[string]bool
}

func NewIngestionService(
	cfg config.Config,
	provider provider.MarketProvider,
	normalizer *Normalizer,
	cache *repository.PriceCache,
	publisher *repository.Publisher,
) *IngestionService {
	subscribed := make(map[string]bool, len(cfg.DefaultSymbols))
	for _, symbol := range cfg.DefaultSymbols {
		subscribed[strings.ToUpper(strings.TrimSpace(symbol))] = true
	}
	return &IngestionService{
		cfg:        cfg,
		provider:   provider,
		normalizer: normalizer,
		cache:      cache,
		publisher:  publisher,
		subscribed: subscribed,
	}
}

// EnsureSubscribed adds a symbol to the Finnhub websocket feed when users view it.
func (s *IngestionService) EnsureSubscribed(symbol string) {
	symbol = strings.ToUpper(strings.TrimSpace(symbol))
	if symbol == "" {
		return
	}

	s.mu.Lock()
	if s.subscribed[symbol] {
		s.mu.Unlock()
		return
	}
	s.subscribed[symbol] = true
	s.mu.Unlock()

	if err := s.provider.Subscribe([]string{symbol}); err != nil {
		log.Printf("finnhub subscribe %s failed: %v", symbol, err)
	}
}

func (s *IngestionService) subscribedList() []string {
	s.mu.Lock()
	defer s.mu.Unlock()
	out := make([]string, 0, len(s.subscribed))
	for symbol := range s.subscribed {
		out = append(out, symbol)
	}
	return out
}

// Run starts ingestion with reconnect handling until context is cancelled.
func (s *IngestionService) Run(ctx context.Context) {
	for {
		if err := s.provider.Connect(ctx); err != nil {
			log.Printf("provider connect failed: %v", err)
			if !sleepOrDone(ctx, 3*time.Second) {
				return
			}
			continue
		}

		if err := s.provider.Subscribe(s.subscribedList()); err != nil {
			log.Printf("provider subscribe failed: %v", err)
			s.provider.Close()
			if !sleepOrDone(ctx, 3*time.Second) {
				return
			}
			continue
		}

		if err := s.consume(ctx); err != nil {
			log.Printf("ingestion stopped: %v", err)
			s.provider.Close()
			if !sleepOrDone(ctx, 3*time.Second) {
				return
			}
		}
	}
}

func (s *IngestionService) consume(ctx context.Context) error {
	trades := s.provider.Trades()
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case raw, ok := <-trades:
			if !ok {
				return context.Canceled
			}
			tick, err := s.normalizer.Normalize(raw)
			if err != nil {
				continue
			}
			if err := s.cache.SetLatest(ctx, tick); err != nil {
				log.Printf("cache update failed: %v", err)
			}
			if err := s.publisher.Publish(ctx, tick); err != nil {
				log.Printf("publish failed: %v", err)
			}
		}
	}
}

func sleepOrDone(ctx context.Context, d time.Duration) bool {
	timer := time.NewTimer(d)
	defer timer.Stop()
	select {
	case <-ctx.Done():
		return false
	case <-timer.C:
		return true
	}
}
