package middleware

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var fixedUserLink = uuid.MustParse("11111111-1111-1111-1111-111111111111")

type stubSessionCheker struct {
	checkFn   func(ctx context.Context, sessionID string) (uuid.UUID, error)
	refreshFn func(ctx context.Context, sessionID string) error
}

func (s *stubSessionCheker) CheckSession(ctx context.Context, sessionID string) (uuid.UUID, error) {
	return s.checkFn(ctx, sessionID)
}

func (s *stubSessionCheker) RefreshSession(ctx context.Context, sessionID string) error {
	return s.refreshFn(ctx, sessionID)
}

func newAuthMiddleware(checkFn func(context.Context, string) (uuid.UUID, error),
	refreshFn func(context.Context, string) error) func(http.Handler) http.Handler {
	logger := zerolog.Nop()
	stub := &stubSessionCheker{checkFn: checkFn, refreshFn: refreshFn}
	return AuthMiddleware(stub, &logger, 24*time.Hour)
}

func TestAuthMiddlewareNoCookie(t *testing.T) {
	mw := newAuthMiddleware(nil, nil)

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	res := httptest.NewRecorder()

	mw(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})).ServeHTTP(res, req)

	assert.Equal(t, http.StatusUnauthorized, res.Code)
}

func TestAuthMiddlewareSuccess(t *testing.T) {
	mw := newAuthMiddleware(
		func(ctx context.Context, sessionID string) (uuid.UUID, error) {
			return fixedUserLink, nil
		},
		func(ctx context.Context, sessionID string) error {
			return nil
		},
	)

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.AddCookie(&http.Cookie{Name: SessiondIdKey, Value: "session-abc"})
	res := httptest.NewRecorder()

	var gotLink uuid.UUID
	mw(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		val := r.Context().Value(UserContextLink{})
		link, ok := val.(uuid.UUID)
		require.True(t, ok, "UserContextLink must be set in context")
		gotLink = link
		w.WriteHeader(http.StatusOK)
	})).ServeHTTP(res, req)

	assert.Equal(t, http.StatusOK, res.Code)
	assert.Equal(t, fixedUserLink, gotLink)
}

func TestAuthMiddlewareCheckSessionError(t *testing.T) {
	mw := newAuthMiddleware(
		func(ctx context.Context, sessionID string) (uuid.UUID, error) {
			return uuid.Nil, errors.New("session not found")
		},
		nil,
	)

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.AddCookie(&http.Cookie{Name: SessiondIdKey, Value: "bad-session"})
	res := httptest.NewRecorder()

	mw(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})).ServeHTTP(res, req)

	assert.Equal(t, http.StatusUnauthorized, res.Code)
}

func TestAuthMiddlewareRefreshSessionError(t *testing.T) {
	mw := newAuthMiddleware(
		func(ctx context.Context, sessionID string) (uuid.UUID, error) {
			return fixedUserLink, nil
		},
		func(ctx context.Context, sessionID string) error {
			return errors.New("redis timeout")
		},
	)

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.AddCookie(&http.Cookie{Name: SessiondIdKey, Value: "session-abc"})
	res := httptest.NewRecorder()

	mw(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})).ServeHTTP(res, req)

	// refresh error must not block the request
	assert.Equal(t, http.StatusOK, res.Code)
}

func TestAuthMiddlewareSetsRefreshCookie(t *testing.T) {
	mw := newAuthMiddleware(
		func(ctx context.Context, sessionID string) (uuid.UUID, error) {
			return fixedUserLink, nil
		},
		func(ctx context.Context, sessionID string) error { return nil },
	)

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.AddCookie(&http.Cookie{Name: SessiondIdKey, Value: "session-abc"})
	res := httptest.NewRecorder()

	mw(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})).ServeHTTP(res, req)

	var found bool
	for _, c := range res.Result().Cookies() {
		if c.Name == SessiondIdKey {
			found = true
			assert.Equal(t, "session-abc", c.Value)
		}
	}
	assert.True(t, found, "session cookie must be refreshed in response")
}
