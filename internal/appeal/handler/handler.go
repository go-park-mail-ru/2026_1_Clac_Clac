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
	"github.com/gorilla/mux"
	"github.com/rs/zerolog"
)

var (
	ErrInvalidRequestSchema = errors.New("invalid schema")
	ErrInvalidEmailOrName   = errors.New("incorrect email or name")
	ErrInvalidActions       = errors.New("this role can not do it")

	msgInternalError = "server error internal"
)

type AppealService interface {
	CreateAppeal(ctx context.Context, appeal serviceDto.EntityAppeal) error
	GetAppeals(ctx context.Context, userLink uuid.UUID) (serviceDto.Appeals, error)
	DeleteAppeal(ctx context.Context, appealLink uuid.UUID) error
	GetStats(ctx context.Context, userLink uuid.UUID) (serviceDto.AppealStats, error)
	ChangeAppealStatus(ctx context.Context, info serviceDto.ChangeAppealStatusInfo) error
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

	appeals, err := h.srv.GetAppeals(r.Context(), userLink)
	if err != nil {
		logger.Error().Err(fmt.Errorf("srv.GetUserAppeals: %w", err)).Msg("failed to get user appeals")
		api.RespondError(w, http.StatusInternalServerError, msgInternalError)
		return
	}

	response := dto.Appeals{
		Role:    appeals.Role,
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
// @Param        link path      string  true  "UUID обращения" format(uuid)
// @Success      200  {string}  string  "OK"
// @Failure      400  {string}  string  "Bad Request (невалидный UUID)"
// @Failure      401  {string}  string  "Unauthorized"
// @Failure      500  {string}  string  "Internal Server Error"
// @Security     BearerAuth
// @Router       /appeals/{link} [delete]
func (h *Handler) DeleteAppeal(w http.ResponseWriter, r *http.Request) {
	logger := zerolog.Ctx(r.Context())

	value := r.Context().Value(middleware.UserContextLink{})
	_, ok := value.(uuid.UUID)
	if !ok {
		api.RespondError(w, http.StatusUnauthorized, "unauthorised")
		return
	}

	vars := mux.Vars(r)
	linkParam := vars["link"]
	appealLink, err := uuid.Parse(linkParam)
	if err != nil {
		api.RespondError(w, http.StatusBadRequest, "invalid appeal link format")
		return
	}

	err = h.srv.DeleteAppeal(r.Context(), appealLink)
	if err != nil {
		logger.Error().Err(fmt.Errorf("srv.DeleteAppeal: %w", err)).Msg("failed to delete appeal")
		api.RespondError(w, http.StatusInternalServerError, msgInternalError)
		return
	}

	api.RespondOk(w, http.StatusOK)
}

// GetStats godoc
// @Summary      Получить статистику обращений
// @Description  Возвращает количество обращений по статусам (доступно только для support/admin)
// @Tags         appeals
// @Produce      json
// @Success      200  {object} dto.AppealStats "Успешный ответ со статистикой"
// @Failure      401  {string} string "Unauthorized"
// @Failure      403  {string} string "Forbidden (Недостаточно прав)"
// @Failure      500  {string} string "Internal Server Error"
// @Security     BearerAuth
// @Router       /stats [get]
func (h *Handler) GetStats(w http.ResponseWriter, r *http.Request) {
	logger := zerolog.Ctx(r.Context())

	value := r.Context().Value(middleware.UserContextLink{})

	userLink, ok := value.(uuid.UUID)
	if !ok {
		api.RespondError(w, http.StatusUnauthorized, "unauthorised")
		return
	}

	stats, err := h.srv.GetStats(r.Context(), userLink)
	if err != nil {
		if errors.Is(err, common.ErrorPermissionDenied) {
			logger.Error().Err(fmt.Errorf("srv.GetStats: %w", err)).Msg("your role can not do it")
			api.RespondError(w, http.StatusForbidden, ErrInvalidActions.Error())
			return
		}

		api.RespondError(w, http.StatusInternalServerError, msgInternalError)
		return
	}

	api.HandleError(api.RespondOk(w, dto.AppealStats{
		Open:   stats.Open,
		Close:  stats.Close,
		InWork: stats.InWork,
	}))
}

// ChangeAppealStatus godoc
// @Summary      Изменить статус обращения
// @Description  Меняет статус существующего обращения (доступно только для support/admin)
// @Tags         appeals
// @Accept       json
// @Produce      json
// @Param        link    path string                 true "UUID обращения" format(uuid)
// @Param        request body dto.ChangeAppealStatus true "Новый статус"
// @Success      200  {string} string "OK"
// @Failure      400  {string} string "Bad Request"
// @Failure      401  {string} string "Unauthorized"
// @Failure      403  {string} string "Forbidden (Недостаточно прав)"
// @Failure      500  {string} string "Internal Server Error"
// @Security     BearerAuth
// @Router       /appeals/{link} [patch]
func (h *Handler) ChangeAppealStatus(w http.ResponseWriter, r *http.Request) {
	logger := zerolog.Ctx(r.Context())

	value := r.Context().Value(middleware.UserContextLink{})

	userLink, ok := value.(uuid.UUID)
	if !ok {
		api.RespondError(w, http.StatusUnauthorized, "unauthorised")
		return
	}

	vars := mux.Vars(r)
	linkParam := vars["link"]

	appealLink, err := uuid.Parse(linkParam)
	if err != nil {
		logger.Error().Err(err).Msg("can not parse appeal link")
		api.RespondError(w, http.StatusBadRequest, common.IncorrectPath)
		return
	}

	var request dto.ChangeAppealStatus
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		logger.Error().Err(err).Msg("can not decode status")
		api.RespondError(w, http.StatusBadRequest, ErrInvalidRequestSchema.Error())
		return
	}

	err = h.srv.ChangeAppealStatus(r.Context(), serviceDto.ChangeAppealStatusInfo{
		SupporterLink: userLink,
		AppealLink:    appealLink,
		Status:        request.Status,
	})
	if err != nil {
		if errors.Is(err, common.ErrorPermissionDenied) {
			logger.Error().Err(fmt.Errorf("srv.ChangeAppealStatus: %w", err)).Msg("your role can not do it")
			api.RespondError(w, http.StatusForbidden, ErrInvalidActions.Error())
			return
		}

		api.RespondError(w, http.StatusInternalServerError, msgInternalError)
		return
	}

	api.RespondOk(w, http.StatusOK)
}
