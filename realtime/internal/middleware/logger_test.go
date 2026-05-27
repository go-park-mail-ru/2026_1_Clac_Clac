package middleware_test

import (
	"bytes"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-park-mail-ru/2026_1_Clac_Clac/realtime/internal/middleware"
	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLoggerMiddleware(t *testing.T) {
	tests := []struct {
		name          string
		handlerStatus int
		expectBody    bool
	}{
		{
			name:          "Success_NoBodyLogged",
			handlerStatus: http.StatusOK,
			expectBody:    false,
		},
		{
			name:          "Error_BodyLogged",
			handlerStatus: http.StatusInternalServerError,
			expectBody:    true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			buf := &bytes.Buffer{}
			logger := zerolog.New(buf)

			h := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				require.NotEqual(t, zerolog.DefaultContextLogger, zerolog.Ctx(r.Context()))
				w.WriteHeader(tc.handlerStatus)
			})

			req := httptest.NewRequest(http.MethodGet, "/api/test", nil)
			res := httptest.NewRecorder()

			middleware.LoggerMiddleware(&logger)(h).ServeHTTP(res, req)

			assert.Equal(t, tc.handlerStatus, res.Code)
			assert.NotEmpty(t, buf.String())
			if tc.expectBody {
				assert.Contains(t, buf.String(), `"body"`)
			} else {
				assert.NotContains(t, buf.String(), `"body"`)
			}
		})
	}
}

func TestLoggerResponseWriter(t *testing.T) {
	tests := []struct {
		name           string
		statusCode     int
		expectedStatus int
	}{
		{
			name:           "200 OK",
			statusCode:     http.StatusOK,
			expectedStatus: http.StatusOK,
		},
		{
			name:           "500 InternalServerError",
			statusCode:     http.StatusInternalServerError,
			expectedStatus: http.StatusInternalServerError,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			res := httptest.NewRecorder()
			lrw := &middleware.LoggerResponseWriter{ResponseWriter: res}
			lrw.WriteHeader(tc.statusCode)
			assert.Equal(t, tc.expectedStatus, res.Code)
		})
	}
}

func TestLoggerLimitWriter(t *testing.T) {
	tests := []struct {
		name              string
		limit             int
		input             []byte
		writeTwice        bool
		dstErr            error
		expectedN         int
		expectErr         bool
		expectedRemaining int
	}{
		{
			name:              "UnderLimit",
			limit:             32,
			input:             []byte("hello world"),
			expectedN:         len([]byte("hello world")),
			expectedRemaining: 32 - len([]byte("hello world")),
		},
		{
			name:              "ExceedsLimit",
			limit:             1,
			input:             []byte("some text"),
			expectedN:         1,
			expectedRemaining: 0,
		},
		{
			name:              "AfterLimitReached",
			limit:             1,
			input:             []byte("some text"),
			writeTwice:        true,
			expectedN:         len([]byte("some text")),
			expectedRemaining: 0,
		},
		{
			name:      "DestinationError",
			limit:     16,
			input:     []byte("data"),
			dstErr:    errors.New("disk full"),
			expectErr: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			var dst interface {
				Write([]byte) (int, error)
			}

			if tc.dstErr != nil {
				dst = &errWriter{err: tc.dstErr}
			} else {
				dst = &bytes.Buffer{}
			}

			w := middleware.NewLoggerLimitWriter(dst, tc.limit)

			if tc.writeTwice {
				_, _ = w.Write(tc.input)
			}

			n, err := w.Write(tc.input)

			if tc.expectErr {
				assert.Error(t, err)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tc.expectedN, n)
				assert.Equal(t, tc.expectedRemaining, w.Remaining)
			}
		})
	}
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
	tests := []struct {
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

	for _, c := range tests {
		assert.Equal(t, c.expected, middleware.ShouldLogRequestBody(c.code), "code %d", c.code)
	}
}
