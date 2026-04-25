package handler

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"

	"github.com/go-park-mail-ru/2026_1_Clac_Clac/internal/api"
	"github.com/go-park-mail-ru/2026_1_Clac_Clac/internal/appeal/handler/dto"
	serviceDto "github.com/go-park-mail-ru/2026_1_Clac_Clac/internal/appeal/service/dto"
	"github.com/go-park-mail-ru/2026_1_Clac_Clac/internal/common"
	"github.com/go-park-mail-ru/2026_1_Clac_Clac/internal/middleware"
	"github.com/google/uuid"
	"github.com/rs/zerolog"
)

var (
	ErrInvalidRequestSchema = errors.New("invalid schema")
	ErrInvalidEmailOrName   = errors.New("invalid email or name")
)

type AppealService interface {
	CreateAppeal(ctx context.Context, appeal serviceDto.EntityAppeal) error
	GetAppeals(ctx context.Context, appeals serviceDto.Appeals)
	DeleteAppeal(ctx context.Context, appealLink uuid.UUID) error
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
	logger := zerolog.Ctx(r.Context())

	value := r.Context().Value(middleware.UserContextLink{})

	userLink, ok := value.(uuid.UUID)
	if !ok {
		api.RespondError(w, http.StatusUnauthorized, "unauthorised")
		return
	}

	var request dto.EntityAppealRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		api.RespondError(w, http.StatusBadRequest, ErrInvalidRequestSchema.Error())
		return
	}

	request.Sanitize()

	if err := ValidatorRequestAppeal(request.Mail, request.DisplayName); err != nil {
		api.RespondError(w, http.StatusBadRequest, ErrInvalidEmailOrName.Error())
		return
	}

	err := h.srv.CreateAppeal(r.Context(), serviceDto.EntityAppeal{
		UserLink:    userLink,
		DisplayName: request.DisplayName,
		Mail:        request.Mail,
		Description: request.Description,
		Category:    request.Category,
	})
	if err != nil {
		logger.Err(fmt.Errorf("srv.CreateAppeal: %w", err))
		if errors.Is(err, common.ErrorExistingUser) {
			api.RespondError(w, http.StatusBadRequest, common.ErrorExistingUser.Error())
			return
		}

		if errors.Is(err, common.ErrorNotNullValue) {
			api.RespondError(w, http.StatusBadRequest, common.ErrorNotNullValue.Error())
			return
		}

		api.RespondError(w, http.StatusInternalServerError, "server error internal")
		return
	}

	api.RespondOk(w, api.StatusOK)
}

func (h *Handler) GetAppeals(w http.ResponseWriter, r *http.Request) {
	logger := zerolog.Ctx(r.Context())

	value := r.Context().Value(middleware.UserContextLink{})

	userLink, ok := value.(uuid.UUID)
	if !ok {
		api.RespondError(w, http.StatusUnauthorized, "unauthorised")
		return
	}
}

func (h *Handler) DeleteAppeal(w http.ResponseWriter, r *http.Request) {

}
