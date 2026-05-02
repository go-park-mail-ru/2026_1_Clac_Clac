package middleware_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-park-mail-ru/2026_1_Clac_Clac/facade/internal/config"
	"github.com/go-park-mail-ru/2026_1_Clac_Clac/facade/internal/middleware"
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

	tests := []struct {
		name                  string
		method                string
		expectedStatus        int
		expectAllCORSHeaders  bool
		expectHandlerCalled   bool
	}{
		{
			name:                 "OPTIONS_preflight",
			method:               http.MethodOptions,
			expectedStatus:       http.StatusNoContent,
			expectAllCORSHeaders: true,
			expectHandlerCalled:  false,
		},
		{
			name:                "GET_request",
			method:              http.MethodGet,
			expectedStatus:      http.StatusOK,
			expectHandlerCalled: true,
		},
		{
			name:                "POST_request",
			method:              http.MethodPost,
			expectedStatus:      http.StatusOK,
			expectHandlerCalled: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			req := httptest.NewRequest(tc.method, "/", nil)
			res := httptest.NewRecorder()

			middleware.CORSMiddleware(&conf)(successHandler).ServeHTTP(res, req)

			assert.Equal(t, tc.expectedStatus, res.Code)
			assert.NotEmpty(t, res.Header().Get("Access-Control-Allow-Credentials"))
			assert.NotEmpty(t, res.Header().Get("Access-Control-Allow-Origin"))

			if tc.expectAllCORSHeaders {
				allHeaders := []string{
					"Access-Control-Allow-Methods",
					"Access-Control-Allow-Headers",
					"Access-Control-Max-Age",
				}
				for _, h := range allHeaders {
					require.NotEmpty(t, res.Header().Get(h), "header %q must be set", h)
				}
				assert.Equal(t, conf.Credentials, res.Header().Get("Access-Control-Allow-Credentials"))
				assert.Equal(t, conf.Origin, res.Header().Get("Access-Control-Allow-Origin"))
				assert.Equal(t, conf.Methods, res.Header().Get("Access-Control-Allow-Methods"))
				assert.Equal(t, conf.Headers, res.Header().Get("Access-Control-Allow-Headers"))
				assert.Equal(t, conf.MaxAge, res.Header().Get("Access-Control-Max-Age"))
			}
		})
	}
}
