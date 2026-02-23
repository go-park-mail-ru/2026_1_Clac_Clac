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
	req, err := http.NewRequest(http.MethodGet, "/", nil)
	if err != nil {
		t.Fatalf("error when creating request: %v", err)
	}

	logger := zerolog.New(nil)

	h := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.NotEqual(t, zerolog.DefaultContextLogger, zerolog.Ctx(r.Context()), "logger should not be default")
	})
	m := middleware.LoggerMiddleware(&logger).Middleware(h)

	m.ServeHTTP(res, req)
}
