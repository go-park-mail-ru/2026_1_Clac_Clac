package middleware_test

import (
	"bytes"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-park-mail-ru/2026_1_Clac_Clac/monolith/internal/middleware"
	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLoggerMiddleware(t *testing.T) {
	t.Run("request without error", func(t *testing.T) {
		res := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/", nil)

		loggerBuffer := new(bytes.Buffer)
		logger := zerolog.New(loggerBuffer)

		h := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			require.NotEqual(t, zerolog.DefaultContextLogger, zerolog.Ctx(r.Context()), "logger should not be default")

			w.WriteHeader(http.StatusOK)
		})
		m := middleware.LoggerMiddleware(&logger).Middleware(h)

		m.ServeHTTP(res, req)

		require.NotContains(t, loggerBuffer.String(), "body", "must not write body when status is ok")
	})

	t.Run("request with error", func(t *testing.T) {
		res := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/", nil)

		loggerBuffer := new(bytes.Buffer)
		logger := zerolog.New(loggerBuffer)

		h := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			require.NotEqual(t, zerolog.DefaultContextLogger, zerolog.Ctx(r.Context()), "logger should not be default")

			w.WriteHeader(http.StatusInternalServerError)
		})
		m := middleware.LoggerMiddleware(&logger).Middleware(h)

		m.ServeHTTP(res, req)

		require.Contains(t, loggerBuffer.String(), "body", "must write body when error happend")
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

type ErrWriter struct {
	Err error
}

func (w *ErrWriter) Write(p []byte) (int, error) {
	return 0, w.Err
}

func TestLoggerLimitWriter(t *testing.T) {
	t.Run("dont reach limit", func(t *testing.T) {
		const writeLimit = 1 * 16 // 16 байт

		input := bytes.NewBuffer([]byte("some text"))
		output := new(bytes.Buffer)

		limitWriter := middleware.NewLoggerLimitWriter(output, writeLimit)
		n, err := limitWriter.Write(input.Bytes())

		require.NoError(t, err, "writer must not return error")
		require.Equal(t, len(input.Bytes()), n, "must read all input")
		assert.LessOrEqual(t, 0, limitWriter.Remaning, "must be non-negative")
	})

	t.Run("reach limit", func(t *testing.T) {
		const writeLimit = 1 // 1 байт

		input := bytes.NewBuffer([]byte("some text"))
		output := new(bytes.Buffer)

		limitWriter := middleware.NewLoggerLimitWriter(output, writeLimit)
		n, err := limitWriter.Write(input.Bytes())

		require.NoError(t, err, "writer must not return error")
		require.Equal(t, writeLimit, n, "must read until limit")

		assert.Equal(t, 0, limitWriter.Remaning, "must be zero after limit reached")
	})

	t.Run("after reaching limit", func(t *testing.T) {
		const writeLimit = 1 // 1 байт

		input := bytes.NewBuffer([]byte("some text"))
		output := new(bytes.Buffer)

		limitWriter := middleware.NewLoggerLimitWriter(output, writeLimit)
		n, err := limitWriter.Write(input.Bytes())

		require.NoError(t, err, "writer must not return error")
		require.Equal(t, writeLimit, n, "must read until limit")
		assert.Equal(t, 0, limitWriter.Remaning, "must be zero after limit reached")

		n, err = limitWriter.Write(input.Bytes())

		require.NoError(t, err, "writer must not return error")
		require.Equal(t, len(input.Bytes()), n, "must return len of data, when limit reached")
		assert.Equal(t, 0, limitWriter.Remaning, "must be zero after limit reached")
	})

	t.Run("reading error", func(t *testing.T) {
		const writeLimit = 1 // 1 байт

		input := bytes.NewBuffer([]byte("some text"))
		output := &ErrWriter{
			Err: errors.New("write error"),
		}

		limitWriter := middleware.NewLoggerLimitWriter(output, writeLimit)
		_, err := limitWriter.Write(input.Bytes())

		require.Error(t, err, "writer must not return error")
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
