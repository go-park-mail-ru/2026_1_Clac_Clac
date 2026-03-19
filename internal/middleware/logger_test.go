package middleware_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-park-mail-ru/2026_1_Clac_Clac/internal/middleware"
	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"
)

func TestLoggerMiddleware(t *testing.T) {
	t.Run("request without error", func(t *testing.T) {
		res := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/", nil)

		logger := zerolog.New(nil)

		h := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			assert.NotEqual(t, zerolog.DefaultContextLogger, zerolog.Ctx(r.Context()), "logger should not be default")
		})
		m := middleware.LoggerMiddleware(&logger).Middleware(h)

		m.ServeHTTP(res, req)
	})
}

func TestLoggerResponseWriter(t *testing.T) {
	t.Run("ok status", func(t *testing.T) {
		res := httptest.NewRecorder()
		loggerResponseWriter := &middleware.LoggerResponseWriter{
			ResponseWriter: res,
		}

		loggerResponseWriter.WriteHeader(http.StatusOK)

		assert.Equal(t, http.StatusOK, res.Code, "http codes must be equal")
	})

	t.Run("internal server error status", func(t *testing.T) {
		res := httptest.NewRecorder()
		loggerResponseWriter := &middleware.LoggerResponseWriter{
			ResponseWriter: res,
		}

		loggerResponseWriter.WriteHeader(http.StatusInternalServerError)

		assert.Equal(t, http.StatusInternalServerError, res.Code, "http codes must be equal")
	})
}

func TestGetRealIP(t *testing.T) {
	const (
		testIpAddr          = "8.8.8.8"
		multipleIpAddr      = "8.8.8.8,10.0.0.1"
		xRealIpHeader       = "X-Real-IP"
		xForwardedForHeader = "X-Forwarded-For"
	)

	t.Run("X-Real-IP header", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		req.Header.Set(xRealIpHeader, testIpAddr)

		actualIp := middleware.GetRealIP(req)

		assert.Equal(t, testIpAddr, actualIp, "IPs must be equal")
	})

	t.Run("X-Forwarded-For header, single ip", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		req.Header.Set(xForwardedForHeader, testIpAddr)

		actualIp := middleware.GetRealIP(req)

		assert.Equal(t, testIpAddr, actualIp, "IPs must be equal")
	})

	t.Run("X-Forwarded-For header, multiple ips", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		req.Header.Set(xForwardedForHeader, multipleIpAddr)

		actualIp := middleware.GetRealIP(req)

		assert.Equal(t, testIpAddr, actualIp, "IPs must be equal")
	})
}
