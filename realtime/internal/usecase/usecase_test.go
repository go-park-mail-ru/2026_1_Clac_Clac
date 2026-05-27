package usecase_test

import (
	"context"
	"errors"
	"testing"

	pubsub "github.com/go-park-mail-ru/2026_1_Clac_Clac/pkg/pubsub"
	"github.com/go-park-mail-ru/2026_1_Clac_Clac/realtime/internal/common"
	"github.com/go-park-mail-ru/2026_1_Clac_Clac/realtime/internal/usecase"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var testBoardLink = uuid.MustParse("bbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbbbbbb")

type mockSubscription struct {
	ch  chan pubsub.Event[common.BoardUpdateEvent]
	err error
}

func (s *mockSubscription) C() <-chan pubsub.Event[common.BoardUpdateEvent] {
	return s.ch
}

func (s *mockSubscription) Close() error {
	return nil
}

func (s *mockSubscription) Err() error {
	return s.err
}

type mockSubscriber struct {
	subscribeFn func(ctx context.Context, channel pubsub.Channel, opts ...pubsub.SubOption) (pubsub.Subscription[common.BoardUpdateEvent], error)
}

func (m *mockSubscriber) Subscribe(ctx context.Context, channel pubsub.Channel, opts ...pubsub.SubOption) (pubsub.Subscription[common.BoardUpdateEvent], error) {
	return m.subscribeFn(ctx, channel, opts...)
}

func TestRealtimeService_Subscribe_Success(t *testing.T) {
	ch := make(chan pubsub.Event[common.BoardUpdateEvent], 1)
	expectedSub := &mockSubscription{ch: ch}

	subscriber := &mockSubscriber{
		subscribeFn: func(ctx context.Context, channel pubsub.Channel, opts ...pubsub.SubOption) (pubsub.Subscription[common.BoardUpdateEvent], error) {
			return expectedSub, nil
		},
	}

	service := usecase.NewRealtimeService(subscriber)
	sub, err := service.Subscribe(context.Background(), testBoardLink)

	require.NoError(t, err)
	assert.Equal(t, expectedSub, sub)
}

func TestRealtimeService_Subscribe_Error(t *testing.T) {
	subscriber := &mockSubscriber{
		subscribeFn: func(ctx context.Context, channel pubsub.Channel, opts ...pubsub.SubOption) (pubsub.Subscription[common.BoardUpdateEvent], error) {
			return nil, errors.New("subscribe error")
		},
	}

	service := usecase.NewRealtimeService(subscriber)
	sub, err := service.Subscribe(context.Background(), testBoardLink)

	require.Error(t, err)
	assert.Nil(t, sub)
}

func TestRealtimeService_Subscribe_CorrectChannel(t *testing.T) {
	var receivedChannel pubsub.Channel

	subscriber := &mockSubscriber{
		subscribeFn: func(ctx context.Context, channel pubsub.Channel, opts ...pubsub.SubOption) (pubsub.Subscription[common.BoardUpdateEvent], error) {
			receivedChannel = channel
			ch := make(chan pubsub.Event[common.BoardUpdateEvent])
			return &mockSubscription{ch: ch}, nil
		},
	}

	service := usecase.NewRealtimeService(subscriber)
	_, err := service.Subscribe(context.Background(), testBoardLink)

	require.NoError(t, err)
	assert.Equal(t, pubsub.Channel(testBoardLink.String()), receivedChannel)
}
