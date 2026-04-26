package handlers

import (
	"net/http"

	"github.com/rs/zerolog"

	"github.com/go-park-mail-ru/2026_1_Clac_Clac/facade/internal/usecase"
)

type BoardHandler struct {
	uc     usecase.BoardUsecase
	logger *zerolog.Logger
}

func NewBoardHandler(uc usecase.BoardUsecase, logger *zerolog.Logger) *BoardHandler {
	return &BoardHandler{uc: uc, logger: logger}
}

func (h *BoardHandler) GetBoards(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusNotImplemented)
}

func (h *BoardHandler) GetBoard(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusNotImplemented)
}
