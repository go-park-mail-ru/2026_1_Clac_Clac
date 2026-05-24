package delivery_test

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-park-mail-ru/2026_1_Clac_Clac/realtime/internal/common"
	delivery "github.com/go-park-mail-ru/2026_1_Clac_Clac/realtime/internal/delivery/http"
	"github.com/go-park-mail-ru/2026_1_Clac_Clac/realtime/internal/middleware"
	serviceDto "github.com/go-park-mail-ru/2026_1_Clac_Clac/realtime/internal/usecase/dto"
	"github.com/google/uuid"
	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"
)

var testBoardLink = uuid.MustParse("bbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbbbbbb")

type mockRealtimeService struct {
	listenFn func(ctx context.Context, boardLink uuid.UUID) (serviceDto.BoardUpdateInfo, error)
}

func (m *mockRealtimeService) Listen(ctx context.Context, boardLink uuid.UUID) (serviceDto.BoardUpdateInfo, error) {
	return m.listenFn(ctx, boardLink)
}

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

func TestEventsLongPolling(t *testing.T) {
	eventPayload := common.BoardUpdateEvent{
		BoardLink: testBoardLink.String(),
		UserLink:  "user-link-123",
		Action:    "create",
	}

	tests := []struct {
		name           string
		boardInCtx     bool
		listenFn       func(ctx context.Context, boardLink uuid.UUID) (serviceDto.BoardUpdateInfo, error)
		expectedStatus int
		expectBody     bool
	}{
		{
			name:       "MissingBoardInContext",
			boardInCtx: false,
			listenFn:   nil,
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:       "Success_EventReceived",
			boardInCtx: true,
			listenFn: func(ctx context.Context, boardLink uuid.UUID) (serviceDto.BoardUpdateInfo, error) {
				return serviceDto.BoardUpdateInfo{
					Type:    "board_update",
					Payload: eventPayload,
				}, nil
			},
			expectedStatus: http.StatusOK,
			expectBody:     true,
		},
		{
			name:       "Timeout_Returns204",
			boardInCtx: true,
			listenFn: func(ctx context.Context, boardLink uuid.UUID) (serviceDto.BoardUpdateInfo, error) {
				return serviceDto.BoardUpdateInfo{}, common.ErrTimeout
			},
			expectedStatus: http.StatusNoContent,
		},
		{
			name:       "Error_ServiceFails",
			boardInCtx: true,
			listenFn: func(ctx context.Context, boardLink uuid.UUID) (serviceDto.BoardUpdateInfo, error) {
				return serviceDto.BoardUpdateInfo{}, errors.New("internal error")
			},
			expectedStatus: http.StatusBadRequest,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			res, req := setupHandlerTest(tc.boardInCtx)

			mockService := &mockRealtimeService{listenFn: tc.listenFn}

			h := delivery.NewRealtimeHandler(mockService)
			h.EventsLongPolling(res, req)

			assert.Equal(t, tc.expectedStatus, res.Code)

			if tc.expectBody && tc.expectedStatus == http.StatusOK {
				assert.Contains(t, res.Body.String(), `"status":"ok"`)
				assert.Contains(t, res.Body.String(), `"type":"board_update"`)
				assert.Contains(t, res.Body.String(), `"board_link":"`+testBoardLink.String()+`"`)
			}
		})
	}
}

func TestEventsLongPolling_ContextPropagatesBoardLink(t *testing.T) {
	var receivedBoardLink uuid.UUID

	mockService := &mockRealtimeService{
		listenFn: func(ctx context.Context, boardLink uuid.UUID) (serviceDto.BoardUpdateInfo, error) {
			receivedBoardLink = boardLink
			return serviceDto.BoardUpdateInfo{
				Type: "board_update",
				Payload: common.BoardUpdateEvent{
					BoardLink: boardLink.String(),
					Action:    "update",
				},
			}, nil
		},
	}

	handler := delivery.NewRealtimeHandler(mockService)
	res := httptest.NewRecorder()

	req := httptest.NewRequest(http.MethodGet, "/api/events/"+testBoardLink.String(), nil)
	logger := zerolog.Nop()
	ctx := logger.WithContext(req.Context())
	ctx = context.WithValue(ctx, middleware.BoardContextLink{}, testBoardLink)
	req = req.WithContext(ctx)

	handler.EventsLongPolling(res, req)

	assert.Equal(t, http.StatusOK, res.Code)
	assert.Equal(t, testBoardLink, receivedBoardLink)
}
