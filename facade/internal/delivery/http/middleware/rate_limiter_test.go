package middleware_test

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-park-mail-ru/2026_1_Clac_Clac/facade/internal/delivery/http/middleware"
	"github.com/go-park-mail-ru/2026_1_Clac_Clac/facade/internal/domain"
	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"
)

type stubCheckLimit struct {
	fn func(ctx context.Context, check domain.RateLimitCheck) (bool, error)
}

func (s *stubCheckLimit) UpdateCountRequests(ctx context.Context, check domain.RateLimitCheck) (bool, error) {
	return s.fn(ctx, check)
}

var testRateCfg = domain.RateLimitConfig{
	Limit:   5,
	Action:  "login",
	WindowS: 60,
}

func TestRateLimiterMiddlewareNotExceeded(t *testing.T) {
	client := &stubCheckLimit{
		fn: func(_ context.Context, _ domain.RateLimitCheck) (bool, error) {
			return false, nil
		},
	}
	logger := zerolog.Nop()
	mw := middleware.RateLimiterMiddleware(client, testRateCfg, &logger)

	req := httptest.NewRequest(http.MethodPost, "/login", nil)
	res := httptest.NewRecorder()

	mw(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})).ServeHTTP(res, req)

	assert.Equal(t, http.StatusOK, res.Code)
}

func TestRateLimiterMiddlewareExceeded(t *testing.T) {
	client := &stubCheckLimit{
		fn: func(_ context.Context, _ domain.RateLimitCheck) (bool, error) {
			return true, nil
		},
	}
	logger := zerolog.Nop()
	mw := middleware.RateLimiterMiddleware(client, testRateCfg, &logger)

	req := httptest.NewRequest(http.MethodPost, "/login", nil)
	res := httptest.NewRecorder()

	mw(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})).ServeHTTP(res, req)

	assert.Equal(t, http.StatusTooManyRequests, res.Code)
}

func TestRateLimiterMiddlewareClientError(t *testing.T) {
	client := &stubCheckLimit{
		fn: func(_ context.Context, _ domain.RateLimitCheck) (bool, error) {
			return false, errors.New("redis unavailable")
		},
	}
	logger := zerolog.Nop()
	mw := middleware.RateLimiterMiddleware(client, testRateCfg, &logger)

	req := httptest.NewRequest(http.MethodPost, "/login", nil)
	res := httptest.NewRecorder()

	mw(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})).ServeHTTP(res, req)

	assert.Equal(t, http.StatusOK, res.Code)
}

func TestGetUserIP(t *testing.T) {
	tests := []struct {
		Name       string
		SetHeaders func(r *http.Request)
		Expected   string
	}{
		{
			Name: "X-Forwarded-For single IP",
			SetHeaders: func(r *http.Request) {
				r.Header.Set("X-Forwarded-For", "8.8.8.8")
			},
			Expected: "8.8.8.8",
		},
		{
			Name: "X-Forwarded-For multiple IPs — first one",
			SetHeaders: func(r *http.Request) {
				r.Header.Set("X-Forwarded-For", "8.8.8.8, 10.0.0.1, 172.16.0.1")
			},
			Expected: "8.8.8.8",
		},
		{
			Name: "X-Real-IP when no X-Forwarded-For",
			SetHeaders: func(r *http.Request) {
				r.Header.Set("X-Real-IP", "1.2.3.4")
			},
			Expected: "1.2.3.4",
		},
		{
			Name:       "RemoteAddr fallback (no proxy headers)",
			SetHeaders: func(r *http.Request) {},
			Expected:   "192.0.2.1",
		},
	}

	for _, tc := range tests {
		t.Run(tc.Name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/", nil)
			tc.SetHeaders(req)
			assert.Equal(t, tc.Expected, middleware.GetUserIP(req))
		})
	}
}
