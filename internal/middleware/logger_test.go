package middleware_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-park-mail-ru/2026_1_Clac_Clac/internal/middleware"
	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLoggerMiddleware(t *testing.T) {
	res := httptest.NewRecorder()
	req, err := http.NewRequest(http.MethodGet, "/", nil)

	require.NoError(t, err, "cannot create request")

	logger := zerolog.New(nil)

	h := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.NotEqual(t, zerolog.DefaultContextLogger, zerolog.Ctx(r.Context()), "logger should not be default")
	})
	m := middleware.LoggerMiddleware(&logger).Middleware(h)

	m.ServeHTTP(res, req)
}
