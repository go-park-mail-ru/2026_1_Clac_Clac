package delivery_test

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	pubsub "github.com/go-park-mail-ru/2026_1_Clac_Clac/pkg/pubsub"
	"github.com/go-park-mail-ru/2026_1_Clac_Clac/realtime/internal/common"
	delivery "github.com/go-park-mail-ru/2026_1_Clac_Clac/realtime/internal/delivery/http"
	"github.com/go-park-mail-ru/2026_1_Clac_Clac/realtime/internal/middleware"
	"github.com/google/uuid"
	"github.com/rs/zerolog"
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

type mockRealtimeService struct {
	subscribeFn func(ctx context.Context, boardLink uuid.UUID) (pubsub.Subscription[common.BoardUpdateEvent], error)
}

func (m *mockRealtimeService) Subscribe(ctx context.Context, boardLink uuid.UUID) (pubsub.Subscription[common.BoardUpdateEvent], error) {
	return m.subscribeFn(ctx, boardLink)
}

type sseRecorder struct {
	*httptest.ResponseRecorder
}

func (r *sseRecorder) Flush() {}

func setupHandlerTest(boardInCtx bool) (*httptest.ResponseRecorder, *http.Request) {
	res := httptest.NewRecorder()

	req := httptest.NewRequest(http.MethodGet, "/api/events/"+testBoardLink.String(), nil)
	if boardInCtx {
		logger := zerolog.Nop()
		ctx := req.Context()
		ctx = logger.WithContext(ctx)
		ctx = context.WithValue(ctx, middleware.BoardContextLink{}, testBoardLink)
		req = req.WithContext(ctx)
	}

	return res, req
}

func TestEventsSSE_MissingBoardLink(t *testing.T) {
	res, req := setupHandlerTest(false)

	ch := make(chan pubsub.Event[common.BoardUpdateEvent])
	mockService := &mockRealtimeService{
		subscribeFn: func(ctx context.Context, boardLink uuid.UUID) (pubsub.Subscription[common.BoardUpdateEvent], error) {
			return &mockSubscription{ch: ch}, nil
		},
	}
	h := delivery.NewRealtimeHandler(mockService)
	h.EventsSSE(res, req)

	assert.Equal(t, http.StatusBadRequest, res.Code)
}

func TestEventsSSE_SubscribeError(t *testing.T) {
	res, req := setupHandlerTest(true)
	sseRes := &sseRecorder{ResponseRecorder: res}

	mockService := &mockRealtimeService{
		subscribeFn: func(ctx context.Context, boardLink uuid.UUID) (pubsub.Subscription[common.BoardUpdateEvent], error) {
			return nil, errors.New("subscribe error")
		},
	}

	h := delivery.NewRealtimeHandler(mockService)
	h.EventsSSE(sseRes, req)

	assert.Equal(t, http.StatusInternalServerError, res.Code)
}

func TestEventsSSE_Streaming(t *testing.T) {
	res, req := setupHandlerTest(true)

	ctx, cancel := context.WithCancel(req.Context())
	req = req.WithContext(ctx)

	sseRes := &sseRecorder{ResponseRecorder: res}

	mockService := &mockRealtimeService{
		subscribeFn: func(ctx context.Context, boardLink uuid.UUID) (pubsub.Subscription[common.BoardUpdateEvent], error) {
			ch := make(chan pubsub.Event[common.BoardUpdateEvent])
			return &mockSubscription{ch: ch}, nil
		},
	}

	h := delivery.NewRealtimeHandler(mockService)

	done := make(chan struct{})
	go func() {
		h.EventsSSE(sseRes, req)
		close(done)
	}()

	time.Sleep(10 * time.Millisecond)
	cancel()

	select {
	case <-done:
	case <-time.After(5 * time.Second):
		t.Fatal("handler did not return after context cancellation")
	}

	assert.Equal(t, "text/event-stream", res.Header().Get("Content-Type"))
	assert.Equal(t, "no-cache", res.Header().Get("Cache-Control"))
	assert.Equal(t, "keep-alive", res.Header().Get("Connection"))
}

func TestEventsSSE_ContextPropagatesBoardLink(t *testing.T) {
	var receivedBoardLink uuid.UUID

	mockService := &mockRealtimeService{
		subscribeFn: func(ctx context.Context, boardLink uuid.UUID) (pubsub.Subscription[common.BoardUpdateEvent], error) {
			receivedBoardLink = boardLink
			ch := make(chan pubsub.Event[common.BoardUpdateEvent])
			return &mockSubscription{ch: ch}, nil
		},
	}

	handler := delivery.NewRealtimeHandler(mockService)
	res := httptest.NewRecorder()
	sseRes := &sseRecorder{ResponseRecorder: res}

	req := httptest.NewRequest(http.MethodGet, "/api/events/"+testBoardLink.String(), nil)
	logger := zerolog.Nop()
	ctx := logger.WithContext(req.Context())
	ctx = context.WithValue(ctx, middleware.BoardContextLink{}, testBoardLink)
	req = req.WithContext(ctx)

	ctx, cancel := context.WithCancel(req.Context())
	req = req.WithContext(ctx)

	done := make(chan struct{})
	go func() {
		handler.EventsSSE(sseRes, req)
		close(done)
	}()

	time.Sleep(10 * time.Millisecond)
	cancel()

	select {
	case <-done:
	case <-time.After(5 * time.Second):
		t.Fatal("handler did not return after context cancellation")
	}

	require.NotNil(t, receivedBoardLink)
	assert.Equal(t, testBoardLink, receivedBoardLink)
}
