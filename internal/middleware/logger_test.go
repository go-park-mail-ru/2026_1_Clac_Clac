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
	res := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/", nil)

	logger := zerolog.New(nil)

	h := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.NotEqual(t, zerolog.DefaultContextLogger, zerolog.Ctx(r.Context()), "logger should not be default")
	})
	m := middleware.LoggerMiddleware(&logger).Middleware(h)

	m.ServeHTTP(res, req)
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
