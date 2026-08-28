package service

import (
	"fmt"
	"strings"

	"market-service/internal/model"
	"market-service/internal/provider"
)

// Normalizer converts provider trades into canonical ticks.
type Normalizer struct{}

func NewNormalizer() *Normalizer {
	return &Normalizer{}
}

func (n *Normalizer) Normalize(raw provider.RawTrade) (*model.NormalizedTick, error) {
	symbol := strings.TrimSpace(strings.ToUpper(raw.Symbol))
	if symbol == "" {
		return nil, fmt.Errorf("missing symbol")
	}
	if raw.Price <= 0 {
		return nil, fmt.Errorf("invalid price")
	}
	if raw.Timestamp <= 0 {
		return nil, fmt.Errorf("invalid timestamp")
	}

	return &model.NormalizedTick{
		Symbol:    symbol,
		Price:     raw.Price,
		Volume:    raw.Volume,
		Timestamp: raw.Timestamp,
		Source:    "finnhub",
	}, nil
}
