package profile

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/go-park-mail-ru/2026_1_Clac_Clac/internal/common"
	"github.com/go-park-mail-ru/2026_1_Clac_Clac/internal/handler"
	"github.com/go-park-mail-ru/2026_1_Clac_Clac/internal/middleware"
	"github.com/go-park-mail-ru/2026_1_Clac_Clac/internal/models"
	"github.com/google/uuid"
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
	value := r.Context().Value(middleware.UserIDKey)

	userID, ok := value.(uuid.UUID)
	if !ok {
		common.MakeJSONError(w, http.StatusUnauthorized, handler.ErrorNotAuth)
		return
	}

	user, err := ps.srv.GetProfileUser(r.Context(), userID)
	if err != nil {
		common.MakeJSONError(w, http.StatusInternalServerError, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	if err := json.NewEncoder(w).Encode(user); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
	}
}
