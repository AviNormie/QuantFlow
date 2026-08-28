package repository_test

import (
	"context"
	"testing"
	"time"

	"market-service/internal/model"
	"market-service/internal/repository"

	"github.com/alicebob/miniredis/v2"
	"github.com/redis/go-redis/v9"
)

func TestPriceCacheSetAndGet(t *testing.T) {
	mr := miniredis.RunT(t)
	client := redis.NewClient(&redis.Options{Addr: mr.Addr()})

	cache := repository.NewPriceCache(client, "market:price:", time.Hour)
	tick := &model.NormalizedTick{
		Symbol:    "AAPL",
		Price:     123.45,
		Volume:    1,
		Timestamp: 1700000000000,
		Source:    "finnhub",
	}

	ctx := context.Background()
	if err := cache.SetLatest(ctx, tick); err != nil {
		t.Fatalf("set latest: %v", err)
	}

	got, err := cache.GetLatest(ctx, "AAPL")
	if err != nil {
		t.Fatalf("get latest: %v", err)
	}
	if got.Price != 123.45 {
		t.Fatalf("unexpected price: %v", got.Price)
	}
}
