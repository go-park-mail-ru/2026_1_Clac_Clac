package middleware_test

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-park-mail-ru/2026_1_Clac_Clac/internal/middleware"
	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRecoverMiddleware(t *testing.T) {
	t.Run("handler without error", func(t *testing.T) {
		res := httptest.NewRecorder()
		req, err := http.NewRequest(http.MethodGet, "/", nil)

		require.NoError(t, err, "cannot create request")

		logger := zerolog.New(nil)

		h := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		})
		m := middleware.RecoveryMiddleware(&logger).Middleware(h)

		m.ServeHTTP(res, req)

		assert.NotEqual(t, http.StatusInternalServerError, res.Code, "should not return 500 status")
	})

	t.Run("handler with panic", func(t *testing.T) {
		res := httptest.NewRecorder()
		req, err := http.NewRequest(http.MethodGet, "/", nil)

		require.NoError(t, err, "cannot create request")

		buf := &bytes.Buffer{}
		logger := zerolog.New(buf)

		h := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			panic("oh no...")
		})
		m := middleware.RecoveryMiddleware(&logger).Middleware(h)

		m.ServeHTTP(res, req)

		assert.Equal(t, http.StatusInternalServerError, res.Code, "should return 500 status")
		assert.NotEmpty(t, buf.Bytes(), "logger should write error")
	})
}
