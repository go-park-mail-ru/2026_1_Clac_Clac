package middleware_test

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-park-mail-ru/2026_1_Clac_Clac/facade/internal/delivery/http/middleware"
	"github.com/stretchr/testify/assert"
)

func TestCSRFMiddleware(t *testing.T) {
	const (
		csrfCookieName = "csrf_token"
		csrfHeader     = "X-CSRF-Token"
		sessionID      = "sess-123"
		csrfToken      = "valid-csrf-token"
	)

	successHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	newCookie := func(name, value string) *http.Cookie {
		return &http.Cookie{Name: name, Value: value}
	}

	type checkerResult int
	const (
		checkerOK checkerResult = iota
		checkerFail
	)

	tests := []struct {
		Name          string
		Method        string
		SessionCookie *http.Cookie
		CSRFCookie    *http.Cookie
		HeaderValue   string
		Checker       checkerResult
		ExpectedCode  int
	}{
		{
			Name:         "no session cookie — 401 even for GET",
			Method:       http.MethodGet,
			ExpectedCode: http.StatusUnauthorized,
		},
		{
			Name:          "GET with session — CSRF skipped, 200",
			Method:        http.MethodGet,
			SessionCookie: newCookie(middleware.SessiondIdKey, sessionID),
			ExpectedCode:  http.StatusOK,
		},
		{
			Name:          "POST all valid — 200",
			Method:        http.MethodPost,
			SessionCookie: newCookie(middleware.SessiondIdKey, sessionID),
			CSRFCookie:    newCookie(csrfCookieName, csrfToken),
			HeaderValue:   csrfToken,
			Checker:       checkerOK,
			ExpectedCode:  http.StatusOK,
		},
		{
			Name:          "POST no CSRF cookie — 403",
			Method:        http.MethodPost,
			SessionCookie: newCookie(middleware.SessiondIdKey, sessionID),
			HeaderValue:   csrfToken,
			ExpectedCode:  http.StatusForbidden,
		},
		{
			Name:          "POST header empty — 403",
			Method:        http.MethodPost,
			SessionCookie: newCookie(middleware.SessiondIdKey, sessionID),
			CSRFCookie:    newCookie(csrfCookieName, csrfToken),
			HeaderValue:   "",
			ExpectedCode:  http.StatusForbidden,
		},
		{
			Name:          "POST header and cookie mismatch — 403",
			Method:        http.MethodPost,
			SessionCookie: newCookie(middleware.SessiondIdKey, sessionID),
			CSRFCookie:    newCookie(csrfCookieName, "token-a"),
			HeaderValue:   "token-b",
			ExpectedCode:  http.StatusForbidden,
		},
		{
			Name:          "POST tokenChecker fails (expired HMAC) — 403",
			Method:        http.MethodPost,
			SessionCookie: newCookie(middleware.SessiondIdKey, sessionID),
			CSRFCookie:    newCookie(csrfCookieName, "bad-token"),
			HeaderValue:   "bad-token",
			Checker:       checkerFail,
			ExpectedCode:  http.StatusForbidden,
		},
		{
			Name:          "DELETE all valid — 200",
			Method:        http.MethodDelete,
			SessionCookie: newCookie(middleware.SessiondIdKey, sessionID),
			CSRFCookie:    newCookie(csrfCookieName, csrfToken),
			HeaderValue:   csrfToken,
			Checker:       checkerOK,
			ExpectedCode:  http.StatusOK,
		},
		{
			Name:          "PUT all valid — 200",
			Method:        http.MethodPut,
			SessionCookie: newCookie(middleware.SessiondIdKey, sessionID),
			CSRFCookie:    newCookie(csrfCookieName, csrfToken),
			HeaderValue:   csrfToken,
			Checker:       checkerOK,
			ExpectedCode:  http.StatusOK,
		},
	}

	for _, tc := range tests {
		t.Run(tc.Name, func(t *testing.T) {
			tokenChecker := func(_ context.Context, _ string, _ string) error {
				if tc.Checker == checkerFail {
					return errors.New("invalid token")
				}
				return nil
			}

			req := httptest.NewRequest(tc.Method, "/", nil)
			if tc.SessionCookie != nil {
				req.AddCookie(tc.SessionCookie)
			}
			if tc.CSRFCookie != nil {
				req.AddCookie(tc.CSRFCookie)
			}
			if tc.HeaderValue != "" {
				req.Header.Set(csrfHeader, tc.HeaderValue)
			}

			res := httptest.NewRecorder()
			middleware.CSRFMiddleware(tokenChecker)(successHandler).ServeHTTP(res, req)

			assert.Equal(t, tc.ExpectedCode, res.Code)
		})
	}
}
