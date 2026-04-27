package middleware

import (
	"context"
	"errors"
	"net/http"

	"github.com/go-park-mail-ru/2026_1_Clac_Clac/facade/internal/api"
	"github.com/gorilla/mux"
)

const (
	csrfCookieKey = "csrf_token"
	xCsrfHeader   = "X-CSRF-Token"
)

var (
	ErrUserNotAuthorized  = errors.New("user not authorized")
	ErrCSRFTokenIncorrect = errors.New("csrf token incorrect")
)

var safeMethods = map[string]struct{}{
	http.MethodGet:     {},
	http.MethodHead:    {},
	http.MethodOptions: {},
	http.MethodTrace:   {},
}

func CSRFMiddleware(tokenChecker func(ctx context.Context, sessionId string, token string) error) mux.MiddlewareFunc {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			sessionCookie, err := r.Cookie(SessiondIdKey)
			if err != nil {
				api.RespondError(w, http.StatusUnauthorized, ErrUserNotAuthorized.Error())
				return
			}
			sessionId := sessionCookie.Value

			if _, isSafe := safeMethods[r.Method]; !isSafe {
				csrfCookie, err := r.Cookie(csrfCookieKey)
				if err != nil {
					api.RespondError(w, http.StatusForbidden, ErrCSRFTokenIncorrect.Error())
					return
				}

				headerToken := r.Header.Get(xCsrfHeader)

				if headerToken == "" || headerToken != csrfCookie.Value {
					api.RespondError(w, http.StatusForbidden, ErrCSRFTokenIncorrect.Error())
					return
				}

				err = tokenChecker(r.Context(), sessionId, csrfCookie.Value)
				if err != nil {
					api.RespondError(w, http.StatusForbidden, ErrCSRFTokenIncorrect.Error())
					return
				}
			}

			next.ServeHTTP(w, r)
		})
	}
}
