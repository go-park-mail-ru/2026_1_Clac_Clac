package middleware

import (
	"context"
	"errors"
	"net/http"

	common "github.com/go-park-mail-ru/2026_1_Clac_Clac/internal/common"
	"github.com/google/uuid"
)

type SessionCheker interface {
	GetUserIDBySession(ctx context.Context, sessionID string) (uuid.UUID, error)
}

func AuthMiddleware(db SessionCheker) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			cookie, err := r.Cookie("session_id")
			if err != nil {
				common.MakeJSONError(w, http.StatusUnauthorized, errors.New("missing cookie"))
				return
			}

			userID, err := db.GetUserIDBySession(r.Context(), cookie.Value)
			if err != nil {
				common.MakeJSONError(w, http.StatusUnauthorized, errors.New("invalid session"))
				return
			}

			ctx := context.WithValue(r.Context(), "userID", userID)

			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}
