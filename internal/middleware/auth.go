package middleware

import (
	"context"
	"net/http"

	"github.com/go-park-mail-ru/2026_1_Clac_Clac/internal/api"
	authSrv "github.com/go-park-mail-ru/2026_1_Clac_Clac/internal/auth/service"
	"github.com/google/uuid"
)

const (
	unauthorizedMessage = "unauthorized"
)

type UserContextLink struct{}

type SessionCheker interface {
	GetUserLink(ctx context.Context, sessionID string) (uuid.UUID, error)
}

func AuthMiddleware(srv SessionCheker) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			cookie, err := r.Cookie(authSrv.SessiondIdKey)
			if err != nil {
				api.RespondError(w, http.StatusUnauthorized, unauthorizedMessage)
				return
			}

			userLink, err := srv.GetUserLink(r.Context(), cookie.Value)
			if err != nil {
				api.RespondError(w, http.StatusUnauthorized, unauthorizedMessage)
				return
			}

			ctx := context.WithValue(r.Context(), UserContextLink{}, userLink)

			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}
