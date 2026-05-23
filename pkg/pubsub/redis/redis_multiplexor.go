package pubsub

import (
	"context"
	"errors"
	"sync"
	"time"

	"github.com/go-park-mail-ru/2026_1_Clac_Clac/pkg/pubsub"
	"github.com/redis/go-redis/v9"
	"github.com/rs/zerolog"
)

// Так как мультиплексор будет общим для всего микросервиса
// то в него может прилететь вообще разные события
// поэтому в нем храним все в исходном виде, но вот
// уже в реализации MuxSubscriber будет конкретный тип
type RawEvent struct {
	ID      string
	Type    string
	Channel string
	Payload []byte
}

type RedisMultiplexor struct {
	rdb        *redis.Client
	streamName string
	mu         sync.RWMutex
	listeners  map[string][]chan<- RawEvent
}

func NewRedisMultiplexor(rdb *redis.Client, streamName string) *RedisMultiplexor {
	return &RedisMultiplexor{
		rdb:        rdb,
		streamName: streamName,
		listeners:  make(map[string][]chan<- RawEvent),
	}
}

func (m *RedisMultiplexor) Start(ctx context.Context) {
	go m.startReading(ctx)
}

func (m *RedisMultiplexor) startReading(ctx context.Context) {
	logger := zerolog.Ctx(ctx)

	lastID := "$"

	for {
		select {
		case <-ctx.Done():
			return
		default:
		}

		streams, err := m.rdb.XRead(ctx, &redis.XReadArgs{
			Streams: []string{m.streamName, lastID},
			Count:   100,
			Block:   2 * time.Second,
		}).Result()
		if err != nil {
			if errors.Is(err, redis.Nil) {
				continue
			}

			if errors.Is(err, context.Canceled) {
				return
			}

			logger.Error().Err(err).Msg("RedisMultiplexor.startReading: redis.Client.XRead")
			continue
		}

		for _, stream := range streams {
			for _, msg := range stream.Messages {
				lastID = msg.ID

				channelStr, ok := msg.Values[pubsub.ChannelKey].(string)
				if !ok {
					logger.Error().Msg("RedisMultiplexor.startReading: parse channel")
					continue
				}

				typeStr, ok := msg.Values[pubsub.TypeKey].(string)
				if !ok {
					logger.Error().Msg("RedisMultiplexor.startReading: parse type")
					continue
				}

				payloadStr, ok := msg.Values[pubsub.PayloadKey].(string)
				if !ok {
					logger.Error().Msg("RedisMultiplexor.startReading: parse payload")
					continue
				}

				raw := RawEvent{
					ID:      msg.ID,
					Type:    typeStr,
					Channel: channelStr,
					Payload: []byte(payloadStr),
				}

				m.mu.RLock()
				for _, listener := range m.listeners[raw.Channel] {
					select {
					case listener <- raw:
					default:
						logger.Warn().Str("channel", raw.Channel).Msg("RedisMultiplexor.startReading: client freeze")
					}
				}
				m.mu.RUnlock()
			}
		}
	}
}

func (m *RedisMultiplexor) subscribe(channel string) (<-chan RawEvent, func()) {
	m.mu.Lock()
	defer m.mu.Unlock()

	ch := make(chan RawEvent, pubsub.DefaultBufferSize)
	m.listeners[channel] = append(m.listeners[channel], ch)

	unsubscribe := func() {
		m.mu.Lock()
		defer m.mu.Unlock()
		subs := m.listeners[channel]
		for i, sub := range subs {
			if sub == ch {
				m.listeners[channel] = append(subs[:i], subs[i+1:]...)
				close(ch)
				break
			}
		}
	}

	return ch, unsubscribe
}
