package handler

import "net/http"

type AppealService interface {
}

type Handler struct {
	srv AppealService
}

func NewHandler(srv AppealService) *Handler {
	return &Handler{
		srv: srv,
	}
}

func (h *Handler) CreateAppeal(w http.ResponseWriter, r *http.Request) {

}
