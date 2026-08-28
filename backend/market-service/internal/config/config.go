package config

import (
	"os"
	"strings"
)

// Config holds market-service runtime configuration.
type Config struct {
	ServiceName       string
	Port              string
	FinnhubAPIKey     string
	PubSubChannel     string
	DefaultSymbols    []string
	PriceCachePrefix  string
	PriceCacheTTLHours int
}

func Load() Config {
	symbols := strings.Split(envOr("MARKET_DEFAULT_SYMBOLS", "AAPL,MSFT,GOOGL"), ",")
	clean := make([]string, 0, len(symbols))
	for _, s := range symbols {
		if trimmed := strings.TrimSpace(strings.ToUpper(s)); trimmed != "" {
			clean = append(clean, trimmed)
		}
	}

	return Config{
		ServiceName:        envOr("SERVICE_NAME", "market-service"),
		Port:               envOr("PORT", "8082"),
		FinnhubAPIKey:      os.Getenv("FINNHUB_API_KEY"),
		PubSubChannel:      envOr("MARKET_PUBSUB_CHANNEL", "market:updates"),
		DefaultSymbols:     clean,
		PriceCachePrefix:   envOr("MARKET_PRICE_PREFIX", "market:price:"),
		PriceCacheTTLHours: 24,
	}
}

func envOr(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}
