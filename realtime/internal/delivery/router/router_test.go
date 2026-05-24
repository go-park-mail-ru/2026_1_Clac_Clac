package delivery_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/go-park-mail-ru/2026_1_Clac_Clac/realtime/internal/config"
	router "github.com/go-park-mail-ru/2026_1_Clac_Clac/realtime/internal/delivery/router"
	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type stubAuthChecker struct{}

func (s *stubAuthChecker) CheckSession(ctx context.Context, sessionID string) (uuid.UUID, error) {
	return uuid.MustParse("11111111-1111-1111-1111-111111111111"), nil
}

type stubBoardChecker struct{}

func (s *stubBoardChecker) CanView(ctx context.Context, userLink, boardLink uuid.UUID) error {
	return nil
}

type stubRealtimeHandler struct{}

func (s *stubRealtimeHandler) EventsSSE(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
}

func TestNewRouter(t *testing.T) {
	logger := zerolog.Nop()
	conf := config.DefaultConfig()

	deps := router.Tools{
		Realtime:     &stubRealtimeHandler{},
		AuthChecker:  &stubAuthChecker{},
		BoardChecker: &stubBoardChecker{},
	}

	r := router.NewRouter(deps, &conf, &logger)
	require.NotNil(t, r)
}

func TestNewRouter_RouteRegistration(t *testing.T) {
	logger := zerolog.Nop()
	conf := config.DefaultConfig()

	deps := router.Tools{
		Realtime:     &stubRealtimeHandler{},
		AuthChecker:  &stubAuthChecker{},
		BoardChecker: &stubBoardChecker{},
	}

	r := router.NewRouter(deps, &conf, &logger)

	tests := []struct {
		name         string
		method       string
		path         string
		addCookie    bool
		expectStatus int
	}{
		{
			name:         "EventsSSE route exists with board_link",
			method:       http.MethodGet,
			path:         "/api/events/bbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbbbbbb",
			addCookie:    true,
			expectStatus: http.StatusOK,
		},
		{
			name:         "Events route without board_link returns 404",
			method:       http.MethodGet,
			path:         "/api/events",
			expectStatus: http.StatusNotFound,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			req := httptest.NewRequest(tc.method, tc.path, nil)
			if tc.addCookie {
				req.AddCookie(&http.Cookie{Name: "session_id", Value: "test-session"})
			}
			res := httptest.NewRecorder()

			r.ServeHTTP(res, req)
			assert.Equal(t, tc.expectStatus, res.Code)
		})
	}
}

func TestNewRouter_MiddlewareChain_RouteNotFound(t *testing.T) {
	logger := zerolog.Nop()
	conf := config.DefaultConfig()

	deps := router.Tools{
		Realtime:     &stubRealtimeHandler{},
		AuthChecker:  &stubAuthChecker{},
		BoardChecker: &stubBoardChecker{},
	}

	r := router.NewRouter(deps, &conf, &logger)

	req := httptest.NewRequest(http.MethodGet, "/unknown-path", nil)
	res := httptest.NewRecorder()

	r.ServeHTTP(res, req)
	assert.Equal(t, http.StatusNotFound, res.Code)
}

func TestNewRouter_MuxSetup(t *testing.T) {
	logger := zerolog.Nop()
	conf := config.DefaultConfig()

	deps := router.Tools{
		Realtime:     &stubRealtimeHandler{},
		AuthChecker:  &stubAuthChecker{},
		BoardChecker: &stubBoardChecker{},
	}

	r := router.NewRouter(deps, &conf, &logger)

	assert.IsType(t, &mux.Router{}, r)
	assert.NotNil(t, r)

	walkFn := func(route *mux.Route, router *mux.Router, ancestors []*mux.Route) error {
		pathTemplate, _ := route.GetPathTemplate()
		methods, _ := route.GetMethods()
		if pathTemplate == "/api/events/{board_link}" {
			assert.Contains(t, methods, http.MethodGet)
		}
		return nil
	}
	require.NoError(t, r.Walk(walkFn))
}

func TestNewRouter_WithValidSession(t *testing.T) {
	logger := zerolog.Nop()
	conf := config.DefaultConfig()
	conf.Services.Auth.SessionLifetime = 24 * time.Hour

	deps := router.Tools{
		Realtime:     &stubRealtimeHandler{},
		AuthChecker:  &stubAuthChecker{},
		BoardChecker: &stubBoardChecker{},
	}

	r := router.NewRouter(deps, &conf, &logger)

	req := httptest.NewRequest(http.MethodGet, "/api/events/"+uuid.New().String(), nil)
	req.AddCookie(&http.Cookie{Name: "session_id", Value: "test-session-value"})
	res := httptest.NewRecorder()

	r.ServeHTTP(res, req)
	assert.Equal(t, http.StatusOK, res.Code)
}
