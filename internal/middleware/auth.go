package middleware

import (
	"context"
	"net/http"

	"github.com/go-park-mail-ru/2026_1_Clac_Clac/internal/api"
	"github.com/go-park-mail-ru/2026_1_Clac_Clac/internal/service/auth"
	"github.com/google/uuid"
)

const (
	unauthorizedMessage = "unauthorized"
)

type UserIDKey struct{}

type SessionCheker interface {
	GetUserID(ctx context.Context, sessionID string) (uuid.UUID, error)
}

func AuthMiddleware(srv SessionCheker) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			cookie, err := r.Cookie(auth.SessiondIdKey)
			if err != nil {
				api.RespondError(w, http.StatusUnauthorized, unauthorizedMessage)
				return
			}

			userID, err := srv.GetUserID(r.Context(), cookie.Value)
			if err != nil {
				api.RespondError(w, http.StatusUnauthorized, unauthorizedMessage)
				return
			}

			ctx := context.WithValue(r.Context(), UserIDKey{}, userID)

			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}
