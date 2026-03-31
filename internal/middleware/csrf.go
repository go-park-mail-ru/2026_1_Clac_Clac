package middleware

import (
	"errors"
	"net/http"

	"github.com/go-park-mail-ru/2026_1_Clac_Clac/internal/api"
)

const (
	csrfCookieKey = "csrf_token"
	xCsrfHeader   = "X-CSRF-Token"
)

var (
	ErrCSRFTokenIncorrect = errors.New("csrf token incorrect")
)

var safeMethods = map[string]struct{}{
	http.MethodGet:     {},
	http.MethodHead:    {},
	http.MethodOptions: {},
	http.MethodTrace:   {},
}

func CSRFMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if _, isSafe := safeMethods[r.Method]; !isSafe {
			cookie, err := r.Cookie(csrfCookieKey)
			if err != nil {
				api.RespondError(w, http.StatusForbidden, ErrCSRFTokenIncorrect.Error())
				return
			}

			headerToken := r.Header.Get(xCsrfHeader)

			if cookie == nil || headerToken == "" || headerToken != cookie.Value {
				api.RespondError(w, http.StatusForbidden, ErrCSRFTokenIncorrect.Error())
				return
			}
		}

		next.ServeHTTP(w, r)
	})
}
