package usecase_test

import (
	"context"
	"errors"
	"testing"
	"time"

	pubsub "github.com/go-park-mail-ru/2026_1_Clac_Clac/pkg/pubsub"
	"github.com/go-park-mail-ru/2026_1_Clac_Clac/realtime/internal/common"
	"github.com/go-park-mail-ru/2026_1_Clac_Clac/realtime/internal/usecase"
	"github.com/go-park-mail-ru/2026_1_Clac_Clac/realtime/internal/usecase/dto"
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

func TestRealtimeService_Listen(t *testing.T) {
	eventPayload := common.BoardUpdateEvent{
		BoardLink:  testBoardLink.String(),
		EntityType: "card",
		EntityLink: "card-link-123",
		Action:     "create",
	}

	tests := []struct {
		name          string
		subscribeFn   func(ctx context.Context, channel pubsub.Channel, opts ...pubsub.SubOption) (pubsub.Subscription[common.BoardUpdateEvent], error)
		ctxTimeout    time.Duration
		expectErr     bool
		expectedErr   error
		expectedEvent *dto.BoardUpdateInfo
	}{
		{
			name: "Success_EventReceived",
			subscribeFn: func(ctx context.Context, channel pubsub.Channel, opts ...pubsub.SubOption) (pubsub.Subscription[common.BoardUpdateEvent], error) {
				ch := make(chan pubsub.Event[common.BoardUpdateEvent], 1)
				ch <- pubsub.Event[common.BoardUpdateEvent]{
					Type:    "board_update",
					Payload: eventPayload,
				}
				return &mockSubscription{ch: ch}, nil
			},
			ctxTimeout: 10 * time.Second,
			expectErr:  false,
			expectedEvent: &dto.BoardUpdateInfo{
				Type:    "board_update",
				Payload: eventPayload,
			},
		},
		{
			name: "Timeout_ContextCancelled",
			subscribeFn: func(ctx context.Context, channel pubsub.Channel, opts ...pubsub.SubOption) (pubsub.Subscription[common.BoardUpdateEvent], error) {
				ch := make(chan pubsub.Event[common.BoardUpdateEvent])
				return &mockSubscription{ch: ch}, nil
			},
			ctxTimeout: 1 * time.Millisecond,
			expectErr:  true,
			expectedErr: common.ErrTimeout,
		},
		{
			name: "Error_SubscribeFails",
			subscribeFn: func(ctx context.Context, channel pubsub.Channel, opts ...pubsub.SubOption) (pubsub.Subscription[common.BoardUpdateEvent], error) {
				return nil, errors.New("subscribe error")
			},
			ctxTimeout: 10 * time.Second,
			expectErr:  true,
		},
		{
			name: "Error_ChannelClosedWithErr",
			subscribeFn: func(ctx context.Context, channel pubsub.Channel, opts ...pubsub.SubOption) (pubsub.Subscription[common.BoardUpdateEvent], error) {
				ch := make(chan pubsub.Event[common.BoardUpdateEvent])
				close(ch)
				return &mockSubscription{ch: ch, err: errors.New("subscription error")}, nil
			},
			ctxTimeout: 10 * time.Second,
			expectErr:  true,
		},
		{
			name: "Error_ChannelClosedNoErr",
			subscribeFn: func(ctx context.Context, channel pubsub.Channel, opts ...pubsub.SubOption) (pubsub.Subscription[common.BoardUpdateEvent], error) {
				ch := make(chan pubsub.Event[common.BoardUpdateEvent])
				close(ch)
				return &mockSubscription{ch: ch, err: nil}, nil
			},
			ctxTimeout: 10 * time.Second,
			expectErr:  true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			subscriber := &mockSubscriber{subscribeFn: tc.subscribeFn}
			service := usecase.NewRealtimeService(subscriber)

			ctx, cancel := context.WithTimeout(context.Background(), tc.ctxTimeout)
			defer cancel()

			result, err := service.Listen(ctx, testBoardLink)

			if tc.expectErr {
				require.Error(t, err)
				if tc.expectedErr != nil {
					assert.True(t, errors.Is(err, tc.expectedErr), "expected %v, got %v", tc.expectedErr, err)
				}
			} else {
				require.NoError(t, err)
				assert.Equal(t, tc.expectedEvent.Type, result.Type)
				assert.Equal(t, tc.expectedEvent.Payload, result.Payload)
			}
		})
	}
}

func TestRealtimeService_Listen_CorrectChannel(t *testing.T) {
	var receivedChannel pubsub.Channel

	subscriber := &mockSubscriber{
		subscribeFn: func(ctx context.Context, channel pubsub.Channel, opts ...pubsub.SubOption) (pubsub.Subscription[common.BoardUpdateEvent], error) {
			receivedChannel = channel
			ch := make(chan pubsub.Event[common.BoardUpdateEvent], 1)
			ch <- pubsub.Event[common.BoardUpdateEvent]{
				Type:    "board_update",
				Payload: common.BoardUpdateEvent{BoardLink: testBoardLink.String(), Action: "update"},
			}
			return &mockSubscription{ch: ch}, nil
		},
	}

	service := usecase.NewRealtimeService(subscriber)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_, err := service.Listen(ctx, testBoardLink)
	require.NoError(t, err)
	assert.Equal(t, pubsub.Channel(testBoardLink.String()), receivedChannel)
}
