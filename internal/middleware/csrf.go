package middleware

import (
	"crypto/rand"
	"encoding/base64"
	"errors"
	"fmt"
	"net/http"

	"github.com/go-park-mail-ru/2026_1_Clac_Clac/internal/api"
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

// SetCSRFCookieHandler устанавливает CSRF куку
//
//	@Summary		Установка CSRF куки
//	@Description	Генерирует новый CSRF токен и записывает его в Cookie
//	@Tags			csrf
//	@Produce		json
//	@Success		200	{object}	api.Response		"ok"	"Успешная установка куки"
//	@Header			200	{string}	Set-Cookie			"csrf_token=...; Path=/; Secure; SameSite=Lax"
//	@Failure		500	{object}	api.ErrorResponse	"internal server error - cannot create token"
//	@Router			/csrf [get]
func SetCSRFCookieHandler(tokenGenerator func() (string, error), logger *zerolog.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		token, err := tokenGenerator()
		if err != nil {
			logger.Error().Err(ErrCannotCreateCSRFToken).Msg("generate token")
			api.RespondError(w, http.StatusInternalServerError, ErrCannotCreateCSRFToken.Error())
			return
		}

		http.SetCookie(w, &http.Cookie{
			Name:     csrfCookieKey,
			Value:    token,
			Secure:   true,
			HttpOnly: false,
			Path:     "/",
			SameSite: http.SameSiteLaxMode,
		})

		api.HandleError(api.Respond(w, http.StatusOK, api.StatusOK))
	}
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
