package middleware

import (
	"context"
	"net/http"
	"time"

	"github.com/go-park-mail-ru/2026_1_Clac_Clac/internal/api"
	"github.com/go-park-mail-ru/2026_1_Clac_Clac/internal/auth/service"
	authSrv "github.com/go-park-mail-ru/2026_1_Clac_Clac/internal/auth/service"
	"github.com/google/uuid"
	"github.com/rs/zerolog"
)

const (
	unauthorizedMessage = "unauthorized"
)

type UserContextLink struct{}

// mockery --name=SessionCheker --output=mock_session_checker --outpkg=SessionCheker
type SessionCheker interface {
	GetUserLink(ctx context.Context, sessionID string) (uuid.UUID, error)
	RefreshSession(ctx context.Context, sessionID string) error
}

func AuthMiddleware(srv SessionCheker, logger *zerolog.Logger, sessionLifeTime time.Duration) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			cookie, err := r.Cookie(authSrv.SessiondIdKey)
			if err != nil {
				api.RespondError(w, http.StatusUnauthorized, unauthorizedMessage)
				return
			}

			tokenID := cookie.Value

			userLink, err := srv.GetUserLink(r.Context(), tokenID)
			if err != nil {
				api.RespondError(w, http.StatusUnauthorized, unauthorizedMessage)
				return
			}

			err = srv.RefreshSession(r.Context(), cookie.Value)
			if err != nil {
				logger.Warn().Err(err).Msg("failed to refresh session")
			}

			http.SetCookie(w, &http.Cookie{
				Name:     service.SessiondIdKey,
				Value:    tokenID,
				Path:     "/",
				Expires:  time.Now().Add(sessionLifeTime),
				Secure:   true,
				HttpOnly: true,
				SameSite: http.SameSiteLaxMode,
			})

			ctx := context.WithValue(r.Context(), UserContextLink{}, userLink)

			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}
