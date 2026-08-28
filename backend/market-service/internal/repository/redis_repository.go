package repository

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"market-service/internal/model"

	"github.com/redis/go-redis/v9"
)

// PriceCache stores latest normalized ticks in Redis.
type PriceCache struct {
	client *redis.Client
	prefix string
	ttl    time.Duration
}

func NewPriceCache(client *redis.Client, prefix string, ttl time.Duration) *PriceCache {
	return &PriceCache{client: client, prefix: prefix, ttl: ttl}
}

func (c *PriceCache) key(symbol string) string {
	return c.prefix + symbol
}

func (c *PriceCache) SetLatest(ctx context.Context, tick *model.NormalizedTick) error {
	payload, err := json.Marshal(tick)
	if err != nil {
		return err
	}
	return c.client.Set(ctx, c.key(tick.Symbol), payload, c.ttl).Err()
}

func (c *PriceCache) GetLatest(ctx context.Context, symbol string) (*model.NormalizedTick, error) {
	value, err := c.client.Get(ctx, c.key(symbol)).Result()
	if err == redis.Nil {
		return nil, fmt.Errorf("price not found")
	}
	if err != nil {
		return nil, err
	}

	var tick model.NormalizedTick
	if err := json.Unmarshal([]byte(value), &tick); err != nil {
		return nil, err
	}
	return &tick, nil
}

// Publisher publishes normalized ticks to Redis Pub/Sub.
type Publisher struct {
	client  *redis.Client
	channel string
}

func NewPublisher(client *redis.Client, channel string) *Publisher {
	return &Publisher{client: client, channel: channel}
}

func (p *Publisher) Publish(ctx context.Context, tick *model.NormalizedTick) error {
	payload, err := json.Marshal(tick)
	if err != nil {
		return err
	}
	return p.client.Publish(ctx, p.channel, payload).Err()
}
