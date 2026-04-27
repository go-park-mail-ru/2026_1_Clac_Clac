package middleware_test

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-park-mail-ru/2026_1_Clac_Clac/facade/internal/delivery/http/middleware"
	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRecoveryMiddlewareNoPanic(t *testing.T) {
	buf := &bytes.Buffer{}
	logger := zerolog.New(buf)

	h := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	req, err := http.NewRequest(http.MethodGet, "/", nil)
	require.NoError(t, err)
	res := httptest.NewRecorder()

	middleware.RecoveryMiddleware(&logger)(h).ServeHTTP(res, req)

	assert.Equal(t, http.StatusOK, res.Code)
}

func TestRecoveryMiddlewarePanic(t *testing.T) {
	buf := &bytes.Buffer{}
	logger := zerolog.New(buf)

	h := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		panic("something exploded")
	})

	req, err := http.NewRequest(http.MethodGet, "/", nil)
	require.NoError(t, err)
	res := httptest.NewRecorder()

	middleware.RecoveryMiddleware(&logger)(h).ServeHTTP(res, req)

	assert.Equal(t, http.StatusInternalServerError, res.Code)
	assert.NotEmpty(t, buf.Bytes(), "panic must be logged")
}
