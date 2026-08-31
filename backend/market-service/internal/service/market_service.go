package service

import (
	"context"
	"fmt"
	"log"
	"strings"

	"market-service/internal/model"
	"market-service/internal/provider"
	"market-service/internal/provider/yahoo"
	"market-service/internal/repository"
	"market-service/internal/service/candles"
	"market-service/internal/service/symbols"
)

// SymbolIngestor subscribes symbols to the live Finnhub websocket feed.
type SymbolIngestor interface {
	EnsureSubscribed(symbol string)
}

// MarketService exposes market read APIs.
type MarketService struct {
	rest   provider.RESTClient
	cache  *repository.PriceCache
	yahoo  *yahoo.Client
	ingest SymbolIngestor
}

func NewMarketService(rest provider.RESTClient, cache *repository.PriceCache, ingest SymbolIngestor) *MarketService {
	return &MarketService{
		rest:   rest,
		cache:  cache,
		yahoo:  yahoo.NewClient(),
		ingest: ingest,
	}
}

func (s *MarketService) trackSymbol(symbol string) {
	if s.ingest == nil {
		return
	}
	s.ingest.EnsureSubscribed(symbol)
}

func (s *MarketService) SearchSymbols(ctx context.Context, query string, limit int) ([]model.SymbolInfo, error) {
	query = strings.TrimSpace(query)
	if limit <= 0 {
		limit = 30
	}
	if limit > 50 {
		limit = 50
	}

	if query == "" {
		return symbols.FilterPopular("", limit), nil
	}

	results, err := s.rest.SearchSymbols(ctx, query)
	if err != nil {
		fallback := symbols.FilterPopular(query, limit)
		if len(fallback) > 0 {
			return fallback, nil
		}
		return nil, err
	}

	if len(results) == 0 {
		return symbols.FilterPopular(query, limit), nil
	}

	if len(results) > limit {
		results = results[:limit]
	}
	return results, nil
}

func (s *MarketService) ResolveSymbol(ctx context.Context, symbol string) (*model.SymbolInfo, error) {
	symbol = strings.ToUpper(strings.TrimSpace(symbol))
	s.trackSymbol(symbol)

	if info, ok := symbols.FindBySymbol(symbol); ok {
		copy := info
		return &copy, nil
	}

	info, err := s.rest.ResolveSymbol(ctx, symbol)
	if err == nil && info != nil {
		return info, nil
	}
	if err != nil {
		log.Printf("finnhub resolve failed for %s: %v", symbol, err)
	}

	return &model.SymbolInfo{
		Symbol:      symbol,
		Name:        symbol,
		Description: symbol,
		Type:        "Common Stock",
		Exchange:    "US",
		Currency:    "USD",
	}, nil
}

func (s *MarketService) GetQuote(ctx context.Context, symbol string) (map[string]float64, error) {
	symbol = strings.ToUpper(strings.TrimSpace(symbol))
	s.trackSymbol(symbol)

	quote, err := s.rest.GetQuote(ctx, symbol)
	if err == nil && isValidQuote(quote) {
		return quote, nil
	}
	if err != nil {
		log.Printf("finnhub quote failed for %s: %v", symbol, err)
	}

	fallback, yerr := s.yahoo.GetQuote(ctx, symbol)
	if yerr != nil {
		log.Printf("yahoo quote failed for %s: %v", symbol, yerr)
	}
	if isValidQuote(fallback) {
		return fallback, nil
	}

	if tick, cacheErr := s.cache.GetLatest(ctx, symbol); cacheErr == nil && tick.Price > 0 {
		return map[string]float64{"c": tick.Price}, nil
	}

	if err != nil {
		return nil, err
	}
	if yerr != nil {
		return nil, yerr
	}
	return nil, fmt.Errorf("no quote data for %s", symbol)
}

func isValidQuote(quote map[string]float64) bool {
	if quote == nil {
		return false
	}
	price, ok := quote["c"]
	return ok && price > 0
}

func (s *MarketService) GetCandles(ctx context.Context, symbol, resolution string, from, to int64) ([]model.Candle, error) {
	symbol = strings.ToUpper(strings.TrimSpace(symbol))
	s.trackSymbol(symbol)
	from, to = candles.NormalizeRange(from, to, resolution)

	result, err := s.rest.GetCandles(ctx, symbol, resolution, from, to)
	if err != nil {
		log.Printf("finnhub candles failed for %s: %v", symbol, err)
	}
	if len(result) > 0 {
		return result, nil
	}

	fallback, yerr := s.yahoo.GetCandles(ctx, symbol, resolution, from, to)
	if yerr != nil {
		log.Printf("yahoo candles failed for %s: %v", symbol, yerr)
		if err != nil {
			return nil, err
		}
		return nil, yerr
	}

	return fallback, nil
}
