package redis

import (
	"context"
	"fmt"
	"os"

	"github.com/redis/go-redis/v9"
)

// Client is the shared Redis client for services that call Connect.
var Client *redis.Client

// Connect parses REDIS_URL, pings Redis, and stores the client globally.
func Connect(ctx context.Context) (*redis.Client, error) {
	url := os.Getenv("REDIS_URL")
	if url == "" {
		return nil, fmt.Errorf("REDIS_URL is not set")
	}

	opts, err := redis.ParseURL(url)
	if err != nil {
		return nil, fmt.Errorf("parse redis url: %w", err)
	}

	client := redis.NewClient(opts)
	if err := client.Ping(ctx).Err(); err != nil {
		client.Close()
		return nil, fmt.Errorf("ping redis: %w", err)
	}

	Client = client
	return client, nil
}

// Ping checks connectivity using the global client.
func Ping(ctx context.Context) error {
	if Client == nil {
		return fmt.Errorf("redis client not initialized")
	}
	return Client.Ping(ctx).Err()
}

// Close shuts down the global Redis client.
func Close() {
	if Client != nil {
		Client.Close()
		Client = nil
	}
}
