package pubsub

type Channel string
type ID string
type Type string

type Event[T any] struct {
	Type    Type
	Payload T
}

type SubOptions struct {
	BufferSize int
}

type SubOption func(*SubOptions)
