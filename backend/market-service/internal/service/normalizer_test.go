package service_test

import (
	"testing"

	"market-service/internal/provider"
	"market-service/internal/service"
)

func TestNormalizerValidTrade(t *testing.T) {
	n := service.NewNormalizer()
	tick, err := n.Normalize(provider.RawTrade{
		Symbol:    "aapl",
		Price:     150.25,
		Volume:    10,
		Timestamp: 1700000000000,
	})
	if err != nil {
		t.Fatalf("normalize: %v", err)
	}
	if tick.Symbol != "AAPL" || tick.Price != 150.25 {
		t.Fatalf("unexpected tick: %+v", tick)
	}
}

func TestNormalizerInvalidTrade(t *testing.T) {
	n := service.NewNormalizer()
	_, err := n.Normalize(provider.RawTrade{Symbol: "", Price: 1, Timestamp: 1})
	if err == nil {
		t.Fatal("expected error for missing symbol")
	}
	_, err = n.Normalize(provider.RawTrade{Symbol: "AAPL", Price: 0, Timestamp: 1})
	if err == nil {
		t.Fatal("expected error for invalid price")
	}
}
