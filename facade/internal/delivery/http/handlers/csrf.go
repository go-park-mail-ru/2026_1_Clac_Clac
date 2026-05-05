package handlers

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/go-park-mail-ru/2026_1_Clac_Clac/facade/internal/api"
	handlerCommon "github.com/go-park-mail-ru/2026_1_Clac_Clac/facade/internal/delivery/http/handlers/common"
	"github.com/go-park-mail-ru/2026_1_Clac_Clac/facade/internal/middleware"
	sentryLogger "github.com/go-park-mail-ru/2026_1_Clac_Clac/pkg/logger"
	"github.com/rs/zerolog"
)

const (
	csrfCookieKey = "csrf_token"
)

type CSRFUsecase interface {
	GetExpireTime(ctx context.Context) time.Time
	Generate(ctx context.Context, sessionID string, expireAt int64) (string, error)
	Check(ctx context.Context, sessionID string, token string) error
}

type CSRF struct {
	csrf CSRFUsecase
}

func NewCSRF(csrf CSRFUsecase) *CSRF {
	return &CSRF{
		csrf: csrf,
	}
}

// SetCSRFCookieHandler генерирует и устанавливает CSRF токен
//
//	@Summary	Получить CSRF токен
//	@Tags		Auth
//	@Security	sessionCookie
//	@Produce	json
//	@Success	200	{object}	api.Response		"OK"
//	@Failure	401	{object}	api.ErrorResponse	"User not authorized"
//	@Failure	500	{object}	api.ErrorResponse	"Cannot create CSRF token"
//	@Router		/csrf [get]
func (c *CSRF) SetCSRFCookieHandler(w http.ResponseWriter, r *http.Request) {
	logger := zerolog.Ctx(r.Context())

	cookie, err := r.Cookie(middleware.SessiondIdKey)
	if err != nil {
		api.RespondError(w, http.StatusUnauthorized, handlerCommon.ErrUserNotAuthorized.Error())
		return
	}
	sessionID := cookie.Value

	expireTime := c.csrf.GetExpireTime(r.Context())

	token, err := c.csrf.Generate(r.Context(), sessionID, expireTime.Unix())
	if err != nil {
		errLog := fmt.Errorf("csrf.Generate: %w", err)
		logger.Error().Err(errLog).Msg("generate csrf token")
		sentryLogger.CaptureFromContext(r.Context(), errLog, "SetCSRFCookieHandler", map[string]interface{}{
			"action": "generate_csrf_token",
		})
		api.RespondError(w, http.StatusInternalServerError, handlerCommon.ErrCannotCreateCSRFToken.Error())
		return
	}

	http.SetCookie(w, api.NewCSRFCookie(csrfCookieKey, token, expireTime))
	api.Respond(w, http.StatusOK, api.StatusOK)
}
