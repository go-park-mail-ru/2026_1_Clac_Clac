package middleware

import (
	"crypto/rand"
	"encoding/base64"
	"errors"
	"fmt"
	"net/http"

	"github.com/go-park-mail-ru/2026_1_Clac_Clac/internal/api"
	"github.com/gorilla/mux"
	"github.com/rs/zerolog"
)

const (
	csrfCookieKey = "csrf_token"
	xCsrfHeader   = "X-CSRF-Token"
)

var (
	ErrCSRFTokenIncorrect    = errors.New("csrf token incorrect")
	ErrCannotCreateCSRFToken = errors.New("cannot create csrf token")
)

var safeMethods = map[string]struct{}{
	http.MethodGet:     {},
	http.MethodHead:    {},
	http.MethodOptions: {},
	http.MethodTrace:   {},
}

func GenerateRandomCSRFToken() (string, error) {
	const tokenLength = 32

	b := make([]byte, tokenLength)

	if _, err := rand.Read(b); err != nil {
		return "", fmt.Errorf("rand.Read: %w", err)
	}

	return base64.URLEncoding.EncodeToString(b), nil
}

// Логику сделал следующую: только безопасный метод может установить куку,
// если запрос был не из безопасного метода, то куки не ставим
func CSRFMiddleware(tokenGenerator func() (string, error), logger *zerolog.Logger) mux.MiddlewareFunc {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			cookie, cookieErr := r.Cookie(csrfCookieKey)

			if _, isSafe := safeMethods[r.Method]; !isSafe {
				headerToken := r.Header.Get(xCsrfHeader)

				if cookie == nil || headerToken == "" || headerToken != cookie.Value {
					api.RespondError(w, http.StatusForbidden, ErrCSRFTokenIncorrect.Error())
					return
				}
			}

			if cookieErr != nil {
				token, err := tokenGenerator()
				if err != nil {
					logger.Error().Err(ErrCannotCreateCSRFToken).Msg("generate token")
					// Нет return - это не ошибка, так как метод безопасный, то мы ничего не теряем
				} else {
					// TODO: допилить api.NewCookie, чтобы можно было
					// создавать куки с разными полями
					http.SetCookie(w, &http.Cookie{
						Name:     csrfCookieKey,
						Value:    token,
						Secure:   true,
						HttpOnly: false,
						Path:     "/",
						SameSite: http.SameSiteLaxMode,
					})
				}
			}

			next.ServeHTTP(w, r)
		})
	}
}
