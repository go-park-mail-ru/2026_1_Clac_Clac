package middleware_test

import (
	"bytes"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-park-mail-ru/2026_1_Clac_Clac/facade/internal/middleware"
	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLoggerMiddlewareSuccess(t *testing.T) {
	buf := &bytes.Buffer{}
	logger := zerolog.New(buf)

	h := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		require.NotEqual(t, zerolog.DefaultContextLogger, zerolog.Ctx(r.Context()))
		w.WriteHeader(http.StatusOK)
	})

	req := httptest.NewRequest(http.MethodGet, "/api/test", nil)
	res := httptest.NewRecorder()

	middleware.LoggerMiddleware(&logger)(h).ServeHTTP(res, req)

	assert.Equal(t, http.StatusOK, res.Code)
	logged := buf.String()
	assert.NotEmpty(t, logged)
	assert.NotContains(t, logged, `"body"`, "body must not be logged on success")
}

func TestLoggerMiddlewareError(t *testing.T) {
	buf := &bytes.Buffer{}
	logger := zerolog.New(buf)

	h := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	})

	req := httptest.NewRequest(http.MethodPost, "/api/test", nil)
	res := httptest.NewRecorder()

	middleware.LoggerMiddleware(&logger)(h).ServeHTTP(res, req)

	assert.Equal(t, http.StatusInternalServerError, res.Code)
	assert.Contains(t, buf.String(), `"body"`, "body field must be logged on error")
}

func TestLoggerResponseWriterStatusCode(t *testing.T) {
	t.Run("200 OK", func(t *testing.T) {
		res := httptest.NewRecorder()
		lrw := &middleware.LoggerResponseWriter{ResponseWriter: res}
		lrw.WriteHeader(http.StatusOK)
		assert.Equal(t, http.StatusOK, res.Code)
	})

	t.Run("500 Internal Server Error", func(t *testing.T) {
		res := httptest.NewRecorder()
		lrw := &middleware.LoggerResponseWriter{ResponseWriter: res}
		lrw.WriteHeader(http.StatusInternalServerError)
		assert.Equal(t, http.StatusInternalServerError, res.Code)
	})
}

func TestLoggerLimitWriterUnderLimit(t *testing.T) {
	const limit = 32

	input := []byte("hello world")
	dst := &bytes.Buffer{}
	w := middleware.NewLoggerLimitWriter(dst, limit)

	n, err := w.Write(input)

	require.NoError(t, err)
	assert.Equal(t, len(input), n)
	assert.Greater(t, w.Remaning, 0)
}

func TestLoggerLimitWriterExceedsLimit(t *testing.T) {
	const limit = 1

	input := []byte("some text")
	dst := &bytes.Buffer{}
	w := middleware.NewLoggerLimitWriter(dst, limit)

	n, err := w.Write(input)

	require.NoError(t, err)
	assert.Equal(t, limit, n)
	assert.Equal(t, 0, w.Remaning)
}

func TestLoggerLimitWriterAfterLimitReached(t *testing.T) {
	const limit = 1

	input := []byte("some text")
	dst := &bytes.Buffer{}
	w := middleware.NewLoggerLimitWriter(dst, limit)

	_, _ = w.Write(input)

	n, err := w.Write(input)
	require.NoError(t, err)
	assert.Equal(t, len(input), n)
	assert.Equal(t, 0, w.Remaning)
}

func TestLoggerLimitWriterDestinationError(t *testing.T) {
	const limit = 16
	errDst := &errWriter{err: errors.New("disk full")}
	w := middleware.NewLoggerLimitWriter(errDst, limit)

	_, err := w.Write([]byte("data"))
	assert.Error(t, err)
}

type errWriter struct{ err error }

func (e *errWriter) Write(p []byte) (int, error) { return 0, e.err }

func TestGetRealIP(t *testing.T) {
	tests := []struct {
		Name       string
		SetHeaders func(r *http.Request)
		Expected   string
	}{
		{
			Name: "X-Real-IP header",
			SetHeaders: func(r *http.Request) {
				r.Header.Set("X-Real-IP", "8.8.8.8")
			},
			Expected: "8.8.8.8",
		},
		{
			Name: "X-Forwarded-For single IP",
			SetHeaders: func(r *http.Request) {
				r.Header.Set("X-Forwarded-For", "1.2.3.4")
			},
			Expected: "1.2.3.4",
		},
		{
			Name: "X-Forwarded-For multiple IPs — first one",
			SetHeaders: func(r *http.Request) {
				r.Header.Set("X-Forwarded-For", "1.2.3.4,10.0.0.1")
			},
			Expected: "1.2.3.4",
		},
		{
			Name: "X-Real-IP takes priority over X-Forwarded-For",
			SetHeaders: func(r *http.Request) {
				r.Header.Set("X-Real-IP", "9.9.9.9")
				r.Header.Set("X-Forwarded-For", "1.2.3.4")
			},
			Expected: "9.9.9.9",
		},
		{
			Name:       "fallback to RemoteAddr",
			SetHeaders: func(r *http.Request) {},
			Expected:   "192.0.2.1:1234",
		},
	}

	for _, tc := range tests {
		t.Run(tc.Name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/", nil)
			tc.SetHeaders(req)
			assert.Equal(t, tc.Expected, middleware.GetRealIP(req))
		})
	}
}

func TestShouldLogRequestBody(t *testing.T) {
	cases := []struct {
		code     int
		expected bool
	}{
		{http.StatusOK, false},
		{http.StatusCreated, false},
		{http.StatusNoContent, false},
		{http.StatusBadRequest, true},
		{http.StatusUnauthorized, true},
		{http.StatusForbidden, true},
		{http.StatusNotFound, true},
		{http.StatusInternalServerError, true},
		{http.StatusBadGateway, true},
	}

	for _, c := range cases {
		assert.Equal(t, c.expected, middleware.ShouldLogRequestBody(c.code), "code %d", c.code)
	}
}
