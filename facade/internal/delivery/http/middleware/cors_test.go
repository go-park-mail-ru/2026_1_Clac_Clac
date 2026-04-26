package middleware_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-park-mail-ru/2026_1_Clac_Clac/facade/internal/config"
	"github.com/go-park-mail-ru/2026_1_Clac_Clac/facade/internal/delivery/http/middleware"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCORSMiddleware(t *testing.T) {
	conf := config.CORS{
		Credentials: "true",
		Origin:      "http://localhost:3000",
		Methods:     "GET, POST, OPTIONS",
		Headers:     "Content-Type, X-CSRF-Token",
		MaxAge:      "86400",
	}

	successHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	t.Run("OPTIONS preflight — all CORS headers set, 204", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodOptions, "/", nil)
		res := httptest.NewRecorder()

		middleware.CORSMiddleware(&conf)(successHandler).ServeHTTP(res, req)

		assert.Equal(t, http.StatusNoContent, res.Code)

		expectedHeaders := []string{
			"Access-Control-Allow-Credentials",
			"Access-Control-Allow-Origin",
			"Access-Control-Allow-Methods",
			"Access-Control-Allow-Headers",
			"Access-Control-Max-Age",
		}
		for _, h := range expectedHeaders {
			require.NotEmpty(t, res.Header().Get(h), "header %q must be set", h)
		}
	})

	t.Run("GET request — credentials and origin headers set, handler called", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		res := httptest.NewRecorder()

		middleware.CORSMiddleware(&conf)(successHandler).ServeHTTP(res, req)

		assert.Equal(t, http.StatusOK, res.Code)
		assert.NotEmpty(t, res.Header().Get("Access-Control-Allow-Credentials"))
		assert.NotEmpty(t, res.Header().Get("Access-Control-Allow-Origin"))
	})

	t.Run("POST request — next handler is called", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "/api/test", nil)
		res := httptest.NewRecorder()

		middleware.CORSMiddleware(&conf)(successHandler).ServeHTTP(res, req)

		assert.Equal(t, http.StatusOK, res.Code)
	})

	t.Run("CORS header values match config", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodOptions, "/", nil)
		res := httptest.NewRecorder()

		middleware.CORSMiddleware(&conf)(successHandler).ServeHTTP(res, req)

		assert.Equal(t, conf.Credentials, res.Header().Get("Access-Control-Allow-Credentials"))
		assert.Equal(t, conf.Origin, res.Header().Get("Access-Control-Allow-Origin"))
		assert.Equal(t, conf.Methods, res.Header().Get("Access-Control-Allow-Methods"))
		assert.Equal(t, conf.Headers, res.Header().Get("Access-Control-Allow-Headers"))
		assert.Equal(t, conf.MaxAge, res.Header().Get("Access-Control-Max-Age"))
	})
}
