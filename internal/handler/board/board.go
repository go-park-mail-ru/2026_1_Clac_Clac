package board

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"

	"github.com/go-park-mail-ru/2026_1_Clac_Clac/internal/common"
	"github.com/go-park-mail-ru/2026_1_Clac_Clac/internal/middleware"
	"github.com/go-park-mail-ru/2026_1_Clac_Clac/internal/models"
	"github.com/google/uuid"
)

var ErrorNotAuth = errors.New("user was not authorized")

type BoardService interface {
	GetBoards(ctx context.Context, userID uuid.UUID) []models.Board
}

func NewBoardHandler(srv BoardService) *BoardHandler {
	return &BoardHandler{
		srv: srv,
	}
}

type BoardHandler struct {
	srv BoardService
}

func (bh *BoardHandler) GetUserBoards(w http.ResponseWriter, r *http.Request) {
	value := r.Context().Value(middleware.UserIDKey)

	userID, ok := value.(uuid.UUID)
	if !ok {
		common.MakeJSONError(w, http.StatusUnauthorized, ErrorNotAuth)
		return
	}

	boards := bh.srv.GetBoards(r.Context(), userID)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	if err := json.NewEncoder(w).Encode(boards); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
	}
}
