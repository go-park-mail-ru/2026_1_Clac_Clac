package middleware_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-park-mail-ru/2026_1_Clac_Clac/monolith/internal/config"
	"github.com/go-park-mail-ru/2026_1_Clac_Clac/monolith/internal/middleware"
	"github.com/stretchr/testify/require"
)

func TestCORSMiddleware(t *testing.T) {
	conf := config.CORS{
		Credentials: "true",
		Origin:      "localhost",
		Methods:     "GET,OPTIONS",
		Headers:     "User-Agent",
		MaxAge:      "60",
	}

	successHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	tests := []struct {
		Name            string
		Method          string
		ExpectedHeaders []string
	}{
		{
			Name:   "preflight request",
			Method: http.MethodOptions,
			ExpectedHeaders: []string{
				"Access-Control-Allow-Credentials",
				"Access-Control-Allow-Origin",
				"Access-Control-Allow-Methods",
				"Access-Control-Allow-Headers",
				"Access-Control-Max-Age",
			},
		},
		{
			Name:   "regular request",
			Method: http.MethodGet,
			ExpectedHeaders: []string{
				"Access-Control-Allow-Credentials",
				"Access-Control-Allow-Origin",
			},
		},
	}

	for _, test := range tests {
		t.Run(test.Name, func(t *testing.T) {
			res := httptest.NewRecorder()
			req := httptest.NewRequest(test.Method, "/", http.NoBody)

			m := middleware.CORSMiddleware(&conf)
			h := m(successHandler)
			h.ServeHTTP(res, req)

			r := res.Result()

			for _, header := range test.ExpectedHeaders {
				exists := (r.Header.Get(header) != "")
				require.True(t, exists, "no header", header)
			}
		})
	}
}
