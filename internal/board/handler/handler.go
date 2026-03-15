package handler

import (
	"context"
	"net/http"

	"github.com/go-park-mail-ru/2026_1_Clac_Clac/internal/api"
	"github.com/go-park-mail-ru/2026_1_Clac_Clac/internal/board/models"
	"github.com/go-park-mail-ru/2026_1_Clac_Clac/internal/middleware"
	"github.com/google/uuid"
)

const (
	unauthorizedMessage = "unauthorized"
)

type BoardService interface {
	GetBoards(ctx context.Context, userID uuid.UUID) ([]models.Board, error)
}

func NewHandler(srv BoardService) *BoardHandler {
	return &BoardHandler{
		srv: srv,
	}
}

type BoardHandler struct {
	srv BoardService
}

func (bh *BoardHandler) GetUserBoards(w http.ResponseWriter, r *http.Request) {
	value := r.Context().Value(middleware.UserIDKey{})

	userID, ok := value.(uuid.UUID)
	if !ok {
		api.RespondError(w, http.StatusUnauthorized, unauthorizedMessage)
		return
	}

	boards, err := bh.srv.GetBoards(r.Context(), userID)
	if err != nil {
		api.RespondError(w, http.StatusUnauthorized, unauthorizedMessage)
		return
	}

	api.HandleError(api.RespondOk(w, boards))
}
