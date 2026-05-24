package pubsub

import (
	"context"
	"sync"

	"github.com/go-park-mail-ru/2026_1_Clac_Clac/pkg/pubsub"
)

type redisSubscription[T any] struct {
	ch     chan pubsub.Event[T]
	cancel context.CancelFunc
	err    error
	mu     sync.RWMutex
}

func (s *redisSubscription[T]) C() <-chan pubsub.Event[T] {
	return s.ch
}

func (s *redisSubscription[T]) Close() error {
	s.cancel()
	return nil
}

func (s *redisSubscription[T]) Err() error {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.err
}

func (s *redisSubscription[T]) setError(err error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.err = err
}
