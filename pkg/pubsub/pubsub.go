package pubsub

import "context"

type Subscription[T any] interface {
	C() <-chan Event[T]
	Close() error
	Err() error
}

type Publisher[T any] interface {
	Publish(ctx context.Context, channel Channel, event Event[T]) (ID, error)
}

type Subscriber[T any] interface {
	Subscribe(ctx context.Context, channel Channel, opts ...SubOption) (Subscription[T], error)
}
