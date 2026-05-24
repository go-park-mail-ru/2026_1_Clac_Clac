package pubsub

import (
	"context"
	"encoding/json"

	"github.com/go-park-mail-ru/2026_1_Clac_Clac/pkg/pubsub"
	"github.com/rs/zerolog"
)

type MuxSubscriber[T any] struct {
	mux *RedisMultiplexor
}

func NewMuxSubscriber[T any](mux *RedisMultiplexor) *MuxSubscriber[T] {
	return &MuxSubscriber[T]{
		mux: mux,
	}
}

func (s *MuxSubscriber[T]) Subscribe(ctx context.Context, channel pubsub.Channel, opts ...pubsub.SubOption) (pubsub.Subscription[T], error) {
	options := &pubsub.SubOptions{BufferSize: pubsub.DefaultBufferSize}
	for _, opt := range opts {
		opt(options)
	}

	rawCh, unsubscribe := s.mux.subscribe(string(channel))

	subCtx, cancel := context.WithCancel(ctx)
	sub := &redisSubscription[T]{
		ch:     make(chan pubsub.Event[T], options.BufferSize),
		cancel: cancel,
	}

	go s.listen(subCtx, sub, unsubscribe, rawCh)

	return sub, nil
}

func (s *MuxSubscriber[T]) listen(ctx context.Context, sub *redisSubscription[T], unsubscribe func(), rawCh <-chan RawEvent) {
	logger := zerolog.Ctx(ctx)

	defer unsubscribe()
	defer close(sub.ch)

	for {
		select {
		case <-ctx.Done():
			sub.setError(ctx.Err())
			return
		case raw, ok := <-rawCh:
			if !ok {
				return
			}

			var payload T
			if err := json.Unmarshal(raw.Payload, &payload); err != nil {
				logger.Error().Err(err).Msg("MuxSubscriber.listen: json.Unmarshal")
				continue
			}

			event := pubsub.Event[T]{
				Type:    pubsub.Type(raw.Type),
				Payload: payload,
			}

			select {
			case sub.ch <- event:
			case <-ctx.Done():
				sub.setError(ctx.Err())
				return
			}
		}
	}
}
