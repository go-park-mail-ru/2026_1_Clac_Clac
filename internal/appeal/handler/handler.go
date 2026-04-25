package handler

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"

	"github.com/go-park-mail-ru/2026_1_Clac_Clac/internal/api"
	"github.com/go-park-mail-ru/2026_1_Clac_Clac/internal/appeal/common"
	"github.com/go-park-mail-ru/2026_1_Clac_Clac/internal/appeal/handler/dto"
	serviceDto "github.com/go-park-mail-ru/2026_1_Clac_Clac/internal/appeal/service/dto"
	"github.com/go-park-mail-ru/2026_1_Clac_Clac/internal/middleware"
	"github.com/google/uuid"
	"github.com/rs/zerolog"
)

var (
	ErrInvalidRequestSchema = errors.New("invalid schema")          // Поправил опечатку shema -> schema
	ErrInvalidEmailOrName   = errors.New("incorrect email or name") // Поправил опечатку ot -> or
)

type AppealService interface {
	CreateAppeal(ctx context.Context, appeal serviceDto.EntityAppeal) error
	GetUserAppeals(ctx context.Context, userLink uuid.UUID) (serviceDto.Appeals, error)
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

// CreateAppeal godoc
// @Summary      Создать обращение
// @Description  Создает новое обращение (тикет) от лица авторизованного пользователя
// @Tags         appeals
// @Accept       json
// @Produce      json
// @Param        request body dto.EntityAppealRequest true "Данные обращения"
// @Success      200  {string} string "OK"
// @Failure      400  {string} string "Bad Request"
// @Failure      401  {string} string "Unauthorized"
// @Failure      500  {string} string "Internal Server Error"
// @Security     BearerAuth
// @Router       /appeals [post]
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
		logger.Error().Err(fmt.Errorf("srv.CreateAppeal: %w", err)).Msg("failed to create appeal")

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

// GetAppeals godoc
// @Summary      Получить список обращений
// @Description  Возвращает все обращения, созданные текущим авторизованным пользователем
// @Tags         appeals
// @Produce      json
// @Success      200  {object} dto.Appeals "Успешный ответ со списком обращений"
// @Failure      401  {string} string "Unauthorized"
// @Failure      500  {string} string "Internal Server Error"
// @Security     BearerAuth
// @Router       /appeals [get]
func (h *Handler) GetAppeals(w http.ResponseWriter, r *http.Request) {
	logger := zerolog.Ctx(r.Context())

	value := r.Context().Value(middleware.UserContextLink{})

	userLink, ok := value.(uuid.UUID)
	if !ok {
		api.RespondError(w, http.StatusUnauthorized, "unauthorised")
		return
	}

	appeals, err := h.srv.GetUserAppeals(r.Context(), userLink)
	if err != nil {
		logger.Error().Err(fmt.Errorf("srv.GetUserAppeals: %w", err)).Msg("failed to get user appeals")
		api.RespondError(w, http.StatusInternalServerError, "server error internal")
		return
	}

	response := dto.Appeals{
		Appeals: make([]dto.Appeal, 0, len(appeals.Appeals)),
	}

	for _, a := range appeals.Appeals {
		response.Appeals = append(response.Appeals, dto.Appeal{
			AppealID:      a.AppelID,
			AppealLink:    a.AppealLink,
			Email:         a.Email,
			DisplayName:   a.DisplayName,
			Category:      a.Category,
			Status:        a.Status,
			Description:   a.Description,
			AttachmentKey: a.AttachmentKey,
			CreatedAt:     a.CreatedAt,
		})
	}

	api.HandleError(api.RespondOk(w, response))
}

// DeleteAppeal godoc
// @Summary      Удалить обращение
// @Description  Удаляет конкретное обращение по его UUID
// @Tags         appeals
// @Param        id   path      string  true  "UUID обращения" format(uuid)
// @Success      200  {string}  string  "OK"
// @Failure      400  {string}  string  "Bad Request (невалидный UUID)"
// @Failure      401  {string}  string  "Unauthorized"
// @Failure      500  {string}  string  "Internal Server Error"
// @Security     BearerAuth
// @Router       /appeals/{id} [delete]
func (h *Handler) DeleteAppeal(w http.ResponseWriter, r *http.Request) {
	logger := zerolog.Ctx(r.Context())

	value := r.Context().Value(middleware.UserContextLink{})
	_, ok := value.(uuid.UUID)
	if !ok {
		api.RespondError(w, http.StatusUnauthorized, "unauthorised")
		return
	}

	appealLinkStr := r.PathValue("id")
	appealLink, err := uuid.Parse(appealLinkStr)
	if err != nil {
		api.RespondError(w, http.StatusBadRequest, "invalid appeal link format")
		return
	}

	err = h.srv.DeleteAppeal(r.Context(), appealLink)
	if err != nil {
		logger.Error().Err(fmt.Errorf("srv.DeleteAppeal: %w", err)).Msg("failed to delete appeal")
		api.RespondError(w, http.StatusInternalServerError, "server error internal")
		return
	}

	api.RespondOk(w, http.StatusOK)
}
