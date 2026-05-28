package middleware

import (
	"context"
	"net/http"
	"time"

	"github.com/go-park-mail-ru/2026_1_Clac_Clac/facade/internal/api"
	"github.com/google/uuid"
	"github.com/rs/zerolog"
)

const (
	SessiondIdKey = "session_id"

	unauthorizedMessage = "unauthorized"
)

type UserContextLink struct{}

// mockery --name=SessionCheker --output=mock_session_checker --outpkg=SessionCheker
type SessionCheker interface {
	CheckSession(ctx context.Context, sessionID string) (uuid.UUID, error)
	RefreshSession(ctx context.Context, sessionID string) error
}

func AuthMiddleware(client SessionCheker, logger *zerolog.Logger, sessionLifeTime time.Duration) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			cookie, err := r.Cookie(SessiondIdKey)
			if err != nil {
				_, _ = api.RespondError(w, http.StatusUnauthorized, unauthorizedMessage)
				return
			}

			tokenID := cookie.Value

			userLink, err := client.CheckSession(r.Context(), tokenID)
			if err != nil {
				_, _ = api.RespondError(w, http.StatusUnauthorized, unauthorizedMessage)
				return
			}

			err = client.RefreshSession(r.Context(), cookie.Value)
			if err != nil {
				logger.Warn().Err(err).Msg("failed to refresh session")
			}

			http.SetCookie(w, &http.Cookie{
				Name:     SessiondIdKey,
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
