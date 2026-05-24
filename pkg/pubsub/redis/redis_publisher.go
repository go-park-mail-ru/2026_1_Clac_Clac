package pubsub

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/go-park-mail-ru/2026_1_Clac_Clac/pkg/pubsub"
	"github.com/redis/go-redis/v9"
)

type RedisPublisher[T any] struct {
	rdb        *redis.Client
	streamName string
}

func NewRedisPublisher[T any](rdb *redis.Client, streamName string) *RedisPublisher[T] {
	return &RedisPublisher[T]{
		rdb:        rdb,
		streamName: streamName,
	}
}

func (p *RedisPublisher[T]) Publish(ctx context.Context, channel pubsub.Channel, event pubsub.Event[T]) (pubsub.ID, error) {
	payloadBytes, err := json.Marshal(event.Payload)
	if err != nil {
		return "", fmt.Errorf("json.Marshal payload: %w", err)
	}

	values := map[string]any{
		pubsub.ChannelKey: string(channel),
		pubsub.TypeKey:    string(event.Type),
		pubsub.PayloadKey: payloadBytes,
	}

	messageID, err := p.rdb.XAdd(ctx, &redis.XAddArgs{
		Stream: p.streamName,
		ID:     "*",
		Values: values,
		MaxLen: 10_000,
		Approx: true,
	}).Result()

	if err != nil {
		return "", fmt.Errorf("redis.Client.XAdd: %w", err)
	}

	return pubsub.ID(messageID), nil
}
