package middleware

import (
	"context"
	"net/http"
	"time"

	"github.com/go-park-mail-ru/2026_1_Clac_Clac/realtime/internal/api"
	"github.com/google/uuid"
	"github.com/rs/zerolog"
)

const (
	SessionIdKey = "session_id"

	unauthorizedMessage = "unauthorized"
)

type UserContextLink struct{}

// mockery --name=SessionChecker --output=mock_session_checker --outpkg=SessionChecker
type SessionChecker interface {
	CheckSession(ctx context.Context, sessionID string) (uuid.UUID, error)
}

func AuthMiddleware(client SessionChecker, logger *zerolog.Logger, sessionLifeTime time.Duration) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			cookie, err := r.Cookie(SessionIdKey)
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

			http.SetCookie(w, &http.Cookie{
				Name:     SessionIdKey,
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
