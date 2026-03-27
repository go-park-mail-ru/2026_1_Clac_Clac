package usecase

import (
	"errors"
	"net/http"

	"github.com/go-park-mail-ru/2026_1_Clac_Clac/internal/api"
	"github.com/rs/zerolog"
)

const (
	csrfCookieKey = "csrf_token"
)

var (
	ErrCannotCreateCSRFToken = errors.New("cannot create csrf token")
)

// SetCSRFCookieHandler устанавливает CSRF куку
//
//	@Summary		Установка CSRF куки
//	@Description	Генерирует новый CSRF токен и записывает его в Cookie.
//	@Description	Вместе с кукой также надо отправлять X-CSRF-Token
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
