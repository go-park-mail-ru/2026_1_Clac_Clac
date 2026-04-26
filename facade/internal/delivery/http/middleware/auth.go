package middleware

import (
	"context"
	"net/http"

	authv1 "github.com/go-park-mail-ru/2026_1_Clac_Clac/pkg/contracts/auth"
	"github.com/rs/zerolog"

	grpcclient "github.com/go-park-mail-ru/2026_1_Clac_Clac/facade/internal/clients/grpc"
)

type contextKey string

const UserLinkKey contextKey = "user_link"

type AuthMiddleware struct {
	auth   *grpcclient.AuthClient
	logger *zerolog.Logger
}

func NewAuthMiddleware(auth *grpcclient.AuthClient, logger *zerolog.Logger) *AuthMiddleware {
	return &AuthMiddleware{auth: auth, logger: logger}
}

func (m *AuthMiddleware) Handle(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cookie, err := r.Cookie("session_id")
		if err != nil {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}

		resp, err := m.auth.GetUserLink(r.Context(), &authv1.GetUserLinkRequest{SessionId: cookie.Value})
		if err != nil {
			m.logger.Warn().Err(err).Msg("auth.GetUserLink")
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}

		ctx := context.WithValue(r.Context(), UserLinkKey, resp.UserLink)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
