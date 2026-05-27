package middleware

import (
	"context"
	"net/http"

	"github.com/go-park-mail-ru/2026_1_Clac_Clac/realtime/internal/api"
	"github.com/google/uuid"
	"github.com/gorilla/mux"
)

const (
	boardPermissionDenied = "board permission denied"
	boardLinkKey          = "board_link"
)

type BoardContextLink struct{}

type BoardChecker interface {
	CanView(ctx context.Context, userLink, boardLink uuid.UUID) error
}

func BoardAccessMiddleware(client BoardChecker) mux.MiddlewareFunc {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			userLink, ok := r.Context().Value(UserContextLink{}).(uuid.UUID)
			if !ok {
				_, _ = api.RespondError(w, http.StatusUnauthorized, unauthorizedMessage)
				return
			}

			rawBoardLink, ok := mux.Vars(r)[boardLinkKey]
			if !ok {
				_, _ = api.RespondError(w, http.StatusBadRequest, "board link missing")
				return
			}

			boardLink, err := uuid.Parse(rawBoardLink)
			if err != nil {
				_, _ = api.RespondError(w, http.StatusBadRequest, "invalid board link")
				return
			}

			if err := client.CanView(r.Context(), userLink, boardLink); err != nil {
				_, _ = api.RespondError(w, http.StatusForbidden, boardPermissionDenied)
				return
			}

			ctx := context.WithValue(r.Context(), BoardContextLink{}, boardLink)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}
