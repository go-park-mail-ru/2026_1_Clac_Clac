package middleware_test

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-park-mail-ru/2026_1_Clac_Clac/monolith/internal/auth/service"
	"github.com/go-park-mail-ru/2026_1_Clac_Clac/monolith/internal/middleware"
	"github.com/stretchr/testify/assert"
)

const (
	csrfCookieKey    = "csrf_token"
	xCSRFTokenHeader = "X-CSRF-Token"
)

func TestCSRFMiddleware(t *testing.T) {
	newCookie := func(name, value string) *http.Cookie {
		return &http.Cookie{
			Name:  name,
			Value: value,
		}
	}

	successHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	type tokenCheckerBehavior int
	const (
		CheckSuccess tokenCheckerBehavior = iota
		CheckFail
	)

	tests := []struct {
		Name          string
		Method        string
		HeaderValue   string
		SessionCookie *http.Cookie
		CSRFCookie    *http.Cookie
		CheckerResult tokenCheckerBehavior
		ExpectedCode  int
	}{
		{
			Name:          "no session cookie - fail even on GET",
			Method:        http.MethodGet,
			SessionCookie: nil,
			ExpectedCode:  http.StatusUnauthorized,
		},
		{
			Name:          "valid session, GET method - success (CSRF ignored)",
			Method:        http.MethodGet,
			SessionCookie: newCookie(service.SessiondIdKey, "sess-123"),
			ExpectedCode:  http.StatusOK,
		},
		{
			Name:          "POST, all valid",
			Method:        http.MethodPost,
			SessionCookie: newCookie(service.SessiondIdKey, "sess-123"),
			CSRFCookie:    newCookie(csrfCookieKey, "valid-token"),
			HeaderValue:   "valid-token",
			CheckerResult: CheckSuccess,
			ExpectedCode:  http.StatusOK,
		},
		{
			Name:          "POST, no CSRF cookie",
			Method:        http.MethodPost,
			SessionCookie: newCookie(service.SessiondIdKey, "sess-123"),
			CSRFCookie:    nil,
			HeaderValue:   "some-token",
			ExpectedCode:  http.StatusForbidden,
		},
		{
			Name:          "POST, header and cookie mismatch",
			Method:        http.MethodPost,
			SessionCookie: newCookie(service.SessiondIdKey, "sess-123"),
			CSRFCookie:    newCookie(csrfCookieKey, "token-a"),
			HeaderValue:   "token-b",
			ExpectedCode:  http.StatusForbidden,
		},
		{
			Name:          "POST, tokenChecker failed (expired or invalid HMAC)",
			Method:        http.MethodPost,
			SessionCookie: newCookie(service.SessiondIdKey, "sess-123"),
			CSRFCookie:    newCookie(csrfCookieKey, "invalid-hmac-token"),
			HeaderValue:   "invalid-hmac-token",
			CheckerResult: CheckFail,
			ExpectedCode:  http.StatusForbidden,
		},
	}

	for _, test := range tests {
		t.Run(test.Name, func(t *testing.T) {
			mockChecker := func(ctx context.Context, sid string, token string) error {
				if test.CheckerResult == CheckFail {
					return errors.New("check failed")
				}
				return nil
			}

			req := httptest.NewRequest(test.Method, "/", nil)
			if test.SessionCookie != nil {
				req.AddCookie(test.SessionCookie)
			}
			if test.CSRFCookie != nil {
				req.AddCookie(test.CSRFCookie)
			}
			if test.HeaderValue != "" {
				req.Header.Add(xCSRFTokenHeader, test.HeaderValue)
			}

			res := httptest.NewRecorder()

			mw := middleware.CSRFMiddleware(mockChecker)
			h := mw(successHandler)

			h.ServeHTTP(res, req)

			assert.Equal(t, test.ExpectedCode, res.Code, "Response code mismatch")
		})
	}
}
