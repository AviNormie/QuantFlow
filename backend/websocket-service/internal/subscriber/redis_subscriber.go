package subscriber

import (
	"context"
	"encoding/json"
	"log"

	"websocket-service/internal/hub"

	"github.com/redis/go-redis/v9"
)

// RedisSubscriber listens to market updates and broadcasts to websocket clients.
type RedisSubscriber struct {
	client  *redis.Client
	channel string
	hub     *hub.Hub
}

func NewRedisSubscriber(client *redis.Client, channel string, hub *hub.Hub) *RedisSubscriber {
	return &RedisSubscriber{client: client, channel: channel, hub: hub}
}

func (s *RedisSubscriber) Run(ctx context.Context) {
	pubsub := s.client.Subscribe(ctx, s.channel)
	defer pubsub.Close()

	ch := pubsub.Channel()
	for {
		select {
		case <-ctx.Done():
			return
		case msg, ok := <-ch:
			if !ok {
				return
			}
			var tick struct {
				Symbol    string  `json:"symbol"`
				Price     float64 `json:"price"`
				Volume    float64 `json:"volume"`
				Timestamp int64   `json:"timestamp"`
			}
			if err := json.Unmarshal([]byte(msg.Payload), &tick); err != nil {
				continue
			}

			payload, err := json.Marshal(map[string]interface{}{
				"type": "trade",
				"data": []map[string]interface{}{
					{
						"s": tick.Symbol,
						"p": tick.Price,
						"v": tick.Volume,
						"t": tick.Timestamp,
					},
				},
			})
			if err != nil {
				continue
			}

			s.hub.Broadcast(tick.Symbol, payload)
		}
	}
}

func (s *RedisSubscriber) RunWithReconnect(ctx context.Context) {
	for {
		if ctx.Err() != nil {
			return
		}
		log.Printf("subscribing to redis channel %s", s.channel)
		s.Run(ctx)
		if ctx.Err() != nil {
			return
		}
		log.Printf("redis subscriber disconnected, retrying...")
	}
}
