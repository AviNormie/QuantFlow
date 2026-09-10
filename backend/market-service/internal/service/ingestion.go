package service

import (
	"context"
	"fmt"
	"log"
	"strings"
	"sync"
	"time"

	"market-service/internal/config"
	"market-service/internal/model"
	"market-service/internal/provider"
	"market-service/internal/provider/yahoo"
	"market-service/internal/repository"
)

// IngestionService consumes provider trades and fans out to Redis.
// It also polls REST quotes so prices keep updating when Finnhub WS is quiet.
type IngestionService struct {
	cfg        config.Config
	market     provider.MarketProvider
	rest       provider.RESTClient
	yahoo      *yahoo.Client
	normalizer *Normalizer
	cache      *repository.PriceCache
	publisher  *repository.Publisher
	mu         sync.Mutex
	subscribed map[string]bool
}

func NewIngestionService(
	cfg config.Config,
	market provider.MarketProvider,
	normalizer *Normalizer,
	cache *repository.PriceCache,
	publisher *repository.Publisher,
) *IngestionService {
	subscribed := make(map[string]bool, len(cfg.DefaultSymbols))
	for _, symbol := range cfg.DefaultSymbols {
		subscribed[strings.ToUpper(strings.TrimSpace(symbol))] = true
	}

	rest, _ := market.(provider.RESTClient)

	return &IngestionService{
		cfg:        cfg,
		market:     market,
		rest:       rest,
		yahoo:      yahoo.NewClient(),
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

	if err := s.market.Subscribe([]string{symbol}); err != nil {
		log.Printf("finnhub subscribe %s failed: %v", symbol, err)
	} else {
		log.Printf("finnhub subscribed %s", symbol)
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

// Run starts Finnhub WS ingestion + quote polling until context is cancelled.
func (s *IngestionService) Run(ctx context.Context) {
	go s.pollQuotes(ctx)

	for {
		if err := s.market.Connect(ctx); err != nil {
			log.Printf("provider connect failed: %v", err)
			if !sleepOrDone(ctx, 3*time.Second) {
				return
			}
			continue
		}
		log.Printf("finnhub ws connected; subscribing %v", s.subscribedList())

		if err := s.market.Subscribe(s.subscribedList()); err != nil {
			log.Printf("provider subscribe failed: %v", err)
			s.market.Close()
			if !sleepOrDone(ctx, 3*time.Second) {
				return
			}
			continue
		}

		if err := s.consume(ctx); err != nil {
			log.Printf("ingestion stopped: %v", err)
			s.market.Close()
			if !sleepOrDone(ctx, 3*time.Second) {
				return
			}
		}
	}
}

func (s *IngestionService) consume(ctx context.Context) error {
	trades := s.market.Trades()
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case raw, ok := <-trades:
			if !ok {
				return fmt.Errorf("finnhub trades channel closed")
			}
			tick, err := s.normalizer.Normalize(raw)
			if err != nil {
				continue
			}
			s.publishTick(ctx, tick)
		}
	}
}

func (s *IngestionService) publishTick(ctx context.Context, tick *model.NormalizedTick) {
	if err := s.cache.SetLatest(ctx, tick); err != nil {
		log.Printf("cache update failed: %v", err)
	}
	if err := s.publisher.Publish(ctx, tick); err != nil {
		log.Printf("publish failed: %v", err)
	}
}

// pollQuotes publishes REST/Yahoo quotes as ticks so the live feed keeps moving
// even when Finnhub websocket trades are missing (common on free tier / after hours).
func (s *IngestionService) pollQuotes(ctx context.Context) {
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			for _, symbol := range s.subscribedList() {
				price, ok := s.fetchLivePrice(ctx, symbol)
				if !ok {
					continue
				}
				tick := &model.NormalizedTick{
					Symbol:    symbol,
					Price:     price,
					Volume:    0,
					Timestamp: time.Now().UnixMilli(),
					Source:    "quote-poll",
				}
				s.publishTick(ctx, tick)
			}
		}
	}
}

func (s *IngestionService) fetchLivePrice(ctx context.Context, symbol string) (float64, bool) {
	// Prefer Yahoo for polling to avoid Finnhub free-tier REST rate limits.
	quote, err := s.yahoo.GetQuote(ctx, symbol)
	if err == nil && quote["c"] > 0 {
		return quote["c"], true
	}
	if s.rest != nil {
		quote, err := s.rest.GetQuote(ctx, symbol)
		if err == nil && quote["c"] > 0 {
			return quote["c"], true
		}
	}
	return 0, false
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
