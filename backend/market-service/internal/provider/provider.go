package provider

import (
	"context"

	"market-service/internal/model"
)

// RawTrade is a provider-specific trade before normalization.
type RawTrade struct {
	Symbol    string
	Price     float64
	Volume    float64
	Timestamp int64
}

// MarketProvider consumes external market data.
type MarketProvider interface {
	Connect(ctx context.Context) error
	Subscribe(symbols []string) error
	Trades() <-chan RawTrade
	Reconnect(ctx context.Context) error
	Close() error
}

// RESTClient fetches REST market data from the provider.
type RESTClient interface {
	SearchSymbols(ctx context.Context, query string) ([]model.SymbolInfo, error)
	ResolveSymbol(ctx context.Context, symbol string) (*model.SymbolInfo, error)
	GetQuote(ctx context.Context, symbol string) (map[string]float64, error)
	GetCandles(ctx context.Context, symbol, resolution string, from, to int64) ([]model.Candle, error)
}
