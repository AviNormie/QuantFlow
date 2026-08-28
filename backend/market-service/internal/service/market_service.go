package service

import (
	"context"
	"strings"

	"market-service/internal/model"
	"market-service/internal/provider"
	"market-service/internal/repository"
)

// MarketService exposes market read APIs.
type MarketService struct {
	rest     provider.RESTClient
	cache    *repository.PriceCache
}

func NewMarketService(rest provider.RESTClient, cache *repository.PriceCache) *MarketService {
	return &MarketService{rest: rest, cache: cache}
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
	return s.rest.GetCandles(ctx, symbol, resolution, from, to)
}
