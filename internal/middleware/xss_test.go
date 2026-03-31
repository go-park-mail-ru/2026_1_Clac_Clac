package middleware_test

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/go-park-mail-ru/2026_1_Clac_Clac/internal/middleware"
	"github.com/stretchr/testify/assert"
)

func TestXSSMiddleware(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	tests := []struct {
		Name               string
		RequestBody        string
		ExpectedStatusCode int
	}{
		{
			Name:               "simple json",
			RequestBody:        `{"username": "user"}`,
			ExpectedStatusCode: http.StatusOK,
		},
		{
			Name: "json with array",
			RequestBody: `
			{
				"usernames": [
					"hello",
					"world",
				]
			}`,
			ExpectedStatusCode: http.StatusOK,
		},
		{
			Name:               "html special codes",
			RequestBody:        `{"username": "&lt;"}`,
			ExpectedStatusCode: http.StatusOK,
		},
		{
			Name: "simple markdown",
			RequestBody: `
			# Title 1

			**bold**, _italic_

			List:
			- Item 1
			- Item 2
			- Item 3
			`,
			ExpectedStatusCode: http.StatusOK,
		},
		{
			Name:               "username with html tags",
			RequestBody:        `{"username": "<b>Hello, World!</b>"}`,
			ExpectedStatusCode: http.StatusBadRequest,
		},
	}

	for _, test := range tests {
		t.Run(test.Name, func(t *testing.T) {
			res := httptest.NewRecorder()
			req := httptest.NewRequest(http.MethodGet, "/", strings.NewReader(test.RequestBody))

			h := middleware.XSSMiddleware(handler)
			h.ServeHTTP(res, req)

			r := res.Result()

			assert.Equal(t, test.ExpectedStatusCode, r.StatusCode, "status codes must be equal")
		})
	}
}
