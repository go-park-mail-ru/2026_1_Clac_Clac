package middleware

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/go-park-mail-ru/2026_1_Clac_Clac/internal/auth/handler/dto"
	serviceDto "github.com/go-park-mail-ru/2026_1_Clac_Clac/internal/auth/service/dto"
	mockCheckLimit "github.com/go-park-mail-ru/2026_1_Clac_Clac/internal/middleware/mock_check_limit"
	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"
)

func TestRateLimiterMiddleware(t *testing.T) {
	t.Run("without error", func(t *testing.T) {
		mockCheckLimit := mockCheckLimit.NewCheckLimit(t)
		mockCheckLimit.On("UpdateCountRequests", context.Background(), serviceDto.RateLimiterConfig{
			Limit:  5,
			Action: "login",
			Window: 1 * time.Minute,
			UserIP: "192.0.2.1",
		}).Return(false, nil)

		nextReq := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		})

		limiter := RateLimiterMiddleware(mockCheckLimit, dto.RateLimitConfig{
			Limit:  5,
			Action: "login",
			Window: 1 * time.Minute,
		}, &zerolog.Logger{})

		testHandler := limiter(nextReq)

		res := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodPost, "/", nil)

		testHandler.ServeHTTP(res, req)

		assert.Equal(t, http.StatusOK, res.Code, "wait status 200")
	})
}

func TestRateLimiterMiddlewareError(t *testing.T) {
	t.Run("Exceeded limit requests", func(t *testing.T) {
		mockCheckLimit := mockCheckLimit.NewCheckLimit(t)
		mockCheckLimit.On("UpdateCountRequests", context.Background(), serviceDto.RateLimiterConfig{
			Limit:  5,
			Action: "login",
			Window: 1 * time.Minute,
			UserIP: "192.0.2.1",
		}).Return(true, nil)

		nextReq := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		})

		limiter := RateLimiterMiddleware(mockCheckLimit, dto.RateLimitConfig{
			Limit:  5,
			Action: "login",
			Window: 1 * time.Minute,
		}, &zerolog.Logger{})

		testHandler := limiter(nextReq)

		res := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodPost, "/", nil)

		testHandler.ServeHTTP(res, req)

		assert.Equal(t, http.StatusTooManyRequests, res.Code, "wait status 429")
	})

	t.Run("Fail Redis update count requests", func(t *testing.T) {
		mockCheckLimit := mockCheckLimit.NewCheckLimit(t)
		mockCheckLimit.On("UpdateCountRequests", context.Background(), serviceDto.RateLimiterConfig{
			Limit:  5,
			Action: "login",
			Window: 1 * time.Minute,
			UserIP: "192.0.2.1",
		}).Return(false, fmt.Errorf("error update count"))

		nextReq := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		})

		limiter := RateLimiterMiddleware(mockCheckLimit, dto.RateLimitConfig{
			Limit:  5,
			Action: "login",
			Window: 1 * time.Minute,
		}, &zerolog.Logger{})

		testHandler := limiter(nextReq)

		res := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodPost, "/", nil)

		testHandler.ServeHTTP(res, req)

		assert.Equal(t, http.StatusOK, res.Code, "wait status 200")
	})
}
