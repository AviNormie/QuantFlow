package service

import (
	"context"
	"log"
	"strings"

	"market-service/internal/model"
	"market-service/internal/provider"
	"market-service/internal/provider/yahoo"
	"market-service/internal/repository"
	"market-service/internal/service/candles"
)

// MarketService exposes market read APIs.
type MarketService struct {
	rest   provider.RESTClient
	cache  *repository.PriceCache
	yahoo  *yahoo.Client
}

func NewMarketService(rest provider.RESTClient, cache *repository.PriceCache) *MarketService {
	return &MarketService{
		rest:  rest,
		cache: cache,
		yahoo: yahoo.NewClient(),
	}
}

func (s *MarketService) SearchSymbols(ctx context.Context, query string) ([]model.SymbolInfo, error) {
	return s.rest.SearchSymbols(ctx, strings.TrimSpace(query))
}

func (s *MarketService) ResolveSymbol(ctx context.Context, symbol string) (*model.SymbolInfo, error) {
	return s.rest.ResolveSymbol(ctx, symbol)
}

func (s *MarketService) GetQuote(ctx context.Context, symbol string) (map[string]float64, error) {
	symbol = strings.ToUpper(strings.TrimSpace(symbol))
	quote, err := s.rest.GetQuote(ctx, symbol)
	if err == nil {
		return quote, nil
	}
	if tick, cacheErr := s.cache.GetLatest(ctx, symbol); cacheErr == nil {
		return map[string]float64{"c": tick.Price}, nil
	}
	return nil, err
}

func (s *MarketService) GetCandles(ctx context.Context, symbol, resolution string, from, to int64) ([]model.Candle, error) {
	symbol = strings.ToUpper(strings.TrimSpace(symbol))
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
