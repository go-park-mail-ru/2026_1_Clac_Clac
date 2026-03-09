package profile

import (
	"context"
	"net/http"

	"github.com/go-park-mail-ru/2026_1_Clac_Clac/internal/api"
	"github.com/go-park-mail-ru/2026_1_Clac_Clac/internal/middleware"
	"github.com/go-park-mail-ru/2026_1_Clac_Clac/internal/models"
	"github.com/google/uuid"
)

const (
	unauthorizedMessage = "unauthorized"
	somethingWentWrong  = "something went wrong"
)

type ProfileService interface {
	GetProfileUser(ctx context.Context, userID uuid.UUID) (models.User, error)
}

func NewProfileHandler(srv ProfileService) *ProfileHandler {
	return &ProfileHandler{
		srv: srv,
	}
}

type ProfileHandler struct {
	srv ProfileService
}

func (ps *ProfileHandler) GetProfile(w http.ResponseWriter, r *http.Request) {
	value := r.Context().Value(middleware.UserIDKey{})

	userID, ok := value.(uuid.UUID)
	if !ok {
		api.RespondError(w, http.StatusUnauthorized, unauthorizedMessage)
		return
	}

	user, err := ps.srv.GetProfileUser(r.Context(), userID)
	if err != nil {
		api.RespondError(w, http.StatusInternalServerError, somethingWentWrong)
		return
	}

	api.HandleError(api.RespondOk(w, user))
}
