package middleware_test

import (
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-park-mail-ru/2026_1_Clac_Clac/internal/middleware"
	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	csrfCookieKey    = "csrf_token"
	xCSRFTokenHeader = "X-CSRF-Token"
)

func TestTestCSRFMiddleware(t *testing.T) {
	testLogger := zerolog.New(io.Discard)

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

	tests := []struct {
		Name           string
		Method         string
		HeaderValue    string
		RequestCookie  *http.Cookie
		TokenGenerator func() (string, error)
		ExpectedCode   int
		ExpectedCookie *http.Cookie
	}{
		{
			Name:           "send user csrf_token cookie, get method",
			Method:         http.MethodGet,
			HeaderValue:    "",
			RequestCookie:  nil,
			TokenGenerator: func() (string, error) { return "123", nil },
			ExpectedCode:   http.StatusOK,
			ExpectedCookie: newCSRFCookie("123"),
		},
		{
			Name:           "do nothing, get method",
			Method:         http.MethodGet,
			HeaderValue:    "123",
			RequestCookie:  newCSRFCookie("123"),
			TokenGenerator: func() (string, error) { return "testing", nil },
			ExpectedCode:   http.StatusOK,
			ExpectedCookie: nil,
		},
		{
			Name:           "header and cookie dont equal, skipped for get method",
			Method:         http.MethodGet,
			HeaderValue:    "123567",
			RequestCookie:  newCSRFCookie("123"),
			TokenGenerator: func() (string, error) { return "testing", nil },
			ExpectedCode:   http.StatusOK,
			ExpectedCookie: nil,
		},
		{
			Name:           "header and cookie correct, post method",
			Method:         http.MethodPost,
			HeaderValue:    "123",
			RequestCookie:  newCSRFCookie("123"),
			TokenGenerator: func() (string, error) { return "testing", nil },
			ExpectedCode:   http.StatusOK,
			ExpectedCookie: nil,
		},
		{
			Name:           "no cookie, post method",
			Method:         http.MethodPost,
			HeaderValue:    "123",
			RequestCookie:  nil,
			TokenGenerator: func() (string, error) { return "testing", nil },
			ExpectedCode:   http.StatusForbidden,
			ExpectedCookie: nil,
		},
		{
			Name:           "no header, post method",
			Method:         http.MethodPost,
			HeaderValue:    "",
			RequestCookie:  newCSRFCookie("123"),
			TokenGenerator: func() (string, error) { return "testing", nil },
			ExpectedCode:   http.StatusForbidden,
			ExpectedCookie: nil,
		},
		{
			Name:           "header and cookie dont equal, post method",
			Method:         http.MethodPost,
			HeaderValue:    "123567",
			RequestCookie:  newCSRFCookie("123"),
			TokenGenerator: func() (string, error) { return "testing", nil },
			ExpectedCode:   http.StatusForbidden,
			ExpectedCookie: nil,
		},
		{
			Name:           "header and cookie dont equal, post method",
			Method:         http.MethodPost,
			HeaderValue:    "123567",
			RequestCookie:  newCSRFCookie("123"),
			TokenGenerator: func() (string, error) { return "testing", nil },
			ExpectedCode:   http.StatusForbidden,
			ExpectedCookie: nil,
		},
		{
			Name:           "token generator error, get method",
			Method:         http.MethodGet,
			HeaderValue:    "",
			RequestCookie:  nil,
			TokenGenerator: func() (string, error) { return "", errors.New("cannot generate token") },
			ExpectedCode:   http.StatusOK,
			ExpectedCookie: nil,
		},
	}

	for _, test := range tests {
		successHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		})

		t.Run(test.Name, func(t *testing.T) {
			res := httptest.NewRecorder()
			req := httptest.NewRequest(test.Method, "/", nil)

			req.Header.Add(xCSRFTokenHeader, test.HeaderValue)

			if test.RequestCookie != nil {
				req.AddCookie(test.RequestCookie)
			}

			m := middleware.CSRFMiddleware(test.TokenGenerator, &testLogger)
			h := m(successHandler)

			h.ServeHTTP(res, req)

			r := res.Result()

			assert.Equal(t, test.ExpectedCode, r.StatusCode, "http codes must be equal")

			cookies := r.Cookies()

			var csrfCookie *http.Cookie
			var exists bool

			for _, c := range cookies {
				if c.Name == csrfCookieKey {
					csrfCookie = c
					exists = true
					break
				}
			}

			if test.ExpectedCookie != nil {
				require.True(t, exists, "cookie must be setted")
				assert.Equal(t, test.ExpectedCookie.Value, csrfCookie.Value, "tokens must be equal")
				assert.Equal(t, test.ExpectedCookie.HttpOnly, csrfCookie.HttpOnly, "httpOnly must be equal")
				assert.Equal(t, test.ExpectedCookie.Path, csrfCookie.Path, "path must be equal")
				assert.Equal(t, test.ExpectedCookie.Secure, csrfCookie.Secure, "secure must be equal")
				assert.Equal(t, test.ExpectedCookie.SameSite, csrfCookie.SameSite, "sameSite must be equal")
			} else {
				require.False(t, exists, "cookie must not be setted")
			}
		})
	}
}
