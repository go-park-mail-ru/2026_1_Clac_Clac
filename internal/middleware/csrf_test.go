package middleware_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-park-mail-ru/2026_1_Clac_Clac/internal/middleware"
	"github.com/stretchr/testify/assert"
)

const (
	csrfCookieKey    = "csrf_token"
	xCSRFTokenHeader = "X-CSRF-Token"
)

func TestTestCSRFMiddleware(t *testing.T) {
	newCSRFCookie := func(value string) *http.Cookie {
		return &http.Cookie{
			Name:     csrfCookieKey,
			Value:    value,
			Path:     "/",
			Secure:   true,
			HttpOnly: false,
			SameSite: http.SameSiteLaxMode,
		}
	}

	successHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	tests := []struct {
		Name          string
		Method        string
		HeaderValue   string
		RequestCookie *http.Cookie
		ExpectedCode  int
	}{
		{
			Name:          "send user csrf_token cookie, get method",
			Method:        http.MethodGet,
			HeaderValue:   "",
			RequestCookie: nil,
			ExpectedCode:  http.StatusOK,
		},
		{
			Name:          "do nothing, get method",
			Method:        http.MethodGet,
			HeaderValue:   "123",
			RequestCookie: newCSRFCookie("123"),
			ExpectedCode:  http.StatusOK,
		},
		{
			Name:          "header and cookie dont equal, skipped for get method",
			Method:        http.MethodGet,
			HeaderValue:   "123567",
			RequestCookie: newCSRFCookie("123"),
			ExpectedCode:  http.StatusOK,
		},
		{
			Name:          "header and cookie correct, post method",
			Method:        http.MethodPost,
			HeaderValue:   "123",
			RequestCookie: newCSRFCookie("123"),
			ExpectedCode:  http.StatusOK,
		},
		{
			Name:          "no cookie, post method",
			Method:        http.MethodPost,
			HeaderValue:   "123",
			RequestCookie: nil,
			ExpectedCode:  http.StatusForbidden,
		},
		{
			Name:          "no header, post method",
			Method:        http.MethodPost,
			HeaderValue:   "",
			RequestCookie: newCSRFCookie("123"),
			ExpectedCode:  http.StatusForbidden,
		},
		{
			Name:          "header and cookie dont equal, post method",
			Method:        http.MethodPost,
			HeaderValue:   "123567",
			RequestCookie: newCSRFCookie("123"),
			ExpectedCode:  http.StatusForbidden,
		},
		{
			Name:          "header and cookie dont equal, post method",
			Method:        http.MethodPost,
			HeaderValue:   "123567",
			RequestCookie: newCSRFCookie("123"),
			ExpectedCode:  http.StatusForbidden,
		},
	}

	for _, test := range tests {
		t.Run(test.Name, func(t *testing.T) {
			res := httptest.NewRecorder()
			req := httptest.NewRequest(test.Method, "/", nil)

			req.Header.Add(xCSRFTokenHeader, test.HeaderValue)

			if test.RequestCookie != nil {
				req.AddCookie(test.RequestCookie)
			}

			h := middleware.CSRFMiddleware(successHandler)

			h.ServeHTTP(res, req)

			r := res.Result()
			assert.Equal(t, test.ExpectedCode, r.StatusCode, "http codes must be equal")
		})
	}
}
