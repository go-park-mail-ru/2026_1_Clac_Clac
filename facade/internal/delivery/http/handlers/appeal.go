package handlers

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"

	"github.com/go-park-mail-ru/2026_1_Clac_Clac/facade/internal/api"
	"github.com/go-park-mail-ru/2026_1_Clac_Clac/facade/internal/common"
	"github.com/go-park-mail-ru/2026_1_Clac_Clac/facade/internal/delivery/http/dto"
	handlerCommon "github.com/go-park-mail-ru/2026_1_Clac_Clac/facade/internal/delivery/http/handlers/common"
	"github.com/go-park-mail-ru/2026_1_Clac_Clac/facade/internal/domain"
	"github.com/go-park-mail-ru/2026_1_Clac_Clac/facade/internal/middleware"
	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/rs/zerolog"
)

var (
	ErrCannotCreateAppeal   = errors.New("cannot create appeal")
	ErrCannotGetAppeals     = errors.New("cannot get appeals")
	ErrAppealLinkMissing    = errors.New("appeal link missing")
	ErrInvalidAppealLink    = errors.New("invalid appeal link")
	ErrCannotUploadFile     = errors.New("cannot upload file")
	ErrParseMultipartForm   = errors.New("cannot parse multipart form")
	ErrCannotFindAttachment = errors.New("cannot find attachment in multipart")
	ErrCannotDeleteAppeal   = errors.New("cannot delete appeal")
	ErrActionDenied         = errors.New("action denied")
	ErrCannotGetStats       = errors.New("cannot get stats")
	ErrCannotChangeStatus   = errors.New("cannot change status")
)

//go:generate mockery --name=AppealUsecase --output=mock_appeal_use_case
type AppealUsecase interface {
	CreateAppeal(ctx context.Context, newAppeal domain.CreateAppealInfo) (uuid.UUID, error)
	GetAppeal(ctx context.Context, userLink uuid.UUID) (string, []domain.AppealInfo, error)
	UploadAttachment(ctx context.Context, attachmentInfo domain.UploadAttachmentInfo, attachment io.Reader) (string, error)
	DeleteAppeal(ctx context.Context, deleteInfo domain.DeleteInfo) error
	GetStats(ctx context.Context, userLink uuid.UUID) (domain.AppealsStats, error)
	ChangeAppealStatus(ctx context.Context, changeStatusInfo domain.ChangeAppealStatusInfo) error
}

type AppealConfig struct {
	MultipartAttachmentFileKey string
	MaxAttachmentSize          int64
}

type Appeal struct {
	service AppealUsecase
	conf    AppealConfig
}

func NewAppeal(service AppealUsecase, conf AppealConfig) *Appeal {
	return &Appeal{
		service: service,
		conf:    conf,
	}
}

// CreateAppeal godoc
// @Summary      Создать обращение
// @Description  Создает новое обращение (тикет) от лица авторизованного пользователя
// @Tags         appeals
// @Accept       json
// @Produce      json
// @Param        request body dto.EntityAppealRequest true "Данные обращения"
// @Success      200  {object} object{appeal_link=string} "Appeal link UUID"
// @Failure      400  {string} string "Bad Request"
// @Failure      401  {string} string "Unauthorized"
// @Failure      500  {string} string "Internal Server Error"
// @Security     BearerAuth
// @Router       /appeals [post]
func (h *Appeal) CreateAppeal(w http.ResponseWriter, r *http.Request) {
	logger := zerolog.Ctx(r.Context())

	value := r.Context().Value(middleware.UserContextLink{})

	userLink, ok := value.(uuid.UUID)
	if !ok {
		api.RespondError(w, http.StatusUnauthorized, handlerCommon.ErrUserNotAuthorized.Error())
		return
	}

	var request dto.CreateAppealRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		api.RespondError(w, http.StatusBadRequest, handlerCommon.ErrInvalidRequestSchema.Error())
		return
	}

	appealLink, err := h.service.CreateAppeal(r.Context(), domain.CreateAppealInfo{
		UserLink:    userLink,
		DisplayName: request.DisplayName,
		Email:       request.Email,
		Description: request.Description,
		Category:    request.Category,
	})
	if err != nil {
		switch {
		case errors.Is(err, common.ErrorExistingUser):
			api.RespondError(w, http.StatusBadRequest, common.ErrorExistingUser.Error())
			return
		case errors.Is(err, common.ErrorNotNullValue):
			api.RespondError(w, http.StatusBadRequest, common.ErrorNotNullValue.Error())
			return
		}

		logger.Error().Err(fmt.Errorf("srv.CreateAppeal: %w", err)).Msg("failed to create appeal")
		api.RespondError(w, http.StatusInternalServerError, ErrCannotCreateAppeal.Error())
		return
	}

	api.HandleError(api.RespondOk(w, dto.CreateAppealResponse{
		AppealLink: appealLink,
	}))
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
func (h *Appeal) GetAppeals(w http.ResponseWriter, r *http.Request) {
	logger := zerolog.Ctx(r.Context())

	value := r.Context().Value(middleware.UserContextLink{})

	userLink, ok := value.(uuid.UUID)
	if !ok {
		api.RespondError(w, http.StatusUnauthorized, handlerCommon.ErrUserNotAuthorized.Error())
		return
	}

	role, appeals, err := h.service.GetAppeal(r.Context(), userLink)
	if err != nil {
		logger.Error().Err(fmt.Errorf("srv.GetAppeal: %w", err)).Msg("failed to get user appeals")
		api.RespondError(w, http.StatusInternalServerError, ErrCannotGetAppeals.Error())
		return
	}

	response := dto.GetAppealsResponse{
		Role:    role,
		Appeals: make([]dto.AppealInfo, 0, len(appeals)),
	}

	for _, a := range appeals {
		response.Appeals = append(response.Appeals, dto.AppealInfo{
			AppealID:      a.AppealID,
			AppealLink:    a.AppealLink,
			Email:         a.Email,
			DisplayName:   a.DisplayName,
			Category:      a.Category,
			Status:        a.Status,
			Description:   a.Description,
			AttachmentURL: a.AttachmentURL,
			CreatedAt:     a.CreatedAt,
		})
	}

	api.HandleError(api.RespondOk(w, response))
}

// UploadAttachment godoc
// @Summary      Загрузить вложение к обращению
// @Description  Загружает изображение (multipart/form-data) и прикрепляет его к обращению
// @Tags         appeals
// @Accept       multipart/form-data
// @Produce      json
// @Param        link        path      string  true  "UUID обращения"  format(uuid)
// @Param        attachment  formData  file    true  "Файл вложения (PNG/JPEG)"
// @Success      200  {object} api.OkResponse[dto.UploadAttachmentResponse]
// @Failure      400  {string} string "Bad Request"
// @Failure      401  {string} string "Unauthorized"
// @Failure      500  {string} string "Internal Server Error"
// @Security     BearerAuth
// @Router       /appeals/{link}/attachment [put]
func (h *Appeal) UploadAttachment(w http.ResponseWriter, r *http.Request) {
	logger := zerolog.Ctx(r.Context())

	value := r.Context().Value(middleware.UserContextLink{})

	userLink, ok := value.(uuid.UUID)
	if !ok {
		api.RespondError(w, http.StatusUnauthorized, handlerCommon.ErrUserNotAuthorized.Error())
		return
	}

	vars := mux.Vars(r)
	rawAppealLink, ok := vars["link"]
	if !ok {
		api.RespondError(w, http.StatusBadRequest, ErrAppealLinkMissing.Error())
		return
	}

	appealLink, err := uuid.Parse(rawAppealLink)
	if err != nil {
		api.RespondError(w, http.StatusBadRequest, ErrInvalidAppealLink.Error())
		return
	}

	if err := r.ParseMultipartForm(h.conf.MaxAttachmentSize); err != nil {
		logger.Error().Err(err).Msg("parse multipart form")
		api.RespondError(w, http.StatusBadRequest, ErrParseMultipartForm.Error())
		return
	}

	file, header, err := r.FormFile(h.conf.MultipartAttachmentFileKey)
	if err != nil {
		logger.Error().Err(err).Msg("cannot find attachment key")
		api.RespondError(w, http.StatusBadRequest, ErrCannotFindAttachment.Error())
		return
	}
	defer func() {
		if err := file.Close(); err != nil {
			logger.Error().Err(err).Msg("Handler.UploadAttachment close file")
		}
	}()

	attachmentURL, err := h.service.UploadAttachment(r.Context(), domain.UploadAttachmentInfo{
		UserLink:   userLink,
		AppealLink: appealLink,
		Filename:   header.Filename,
	}, file)
	if err != nil {
		logger.Error().Err(fmt.Errorf("srv.UploadAttachment: %w", err)).Msg("failed to upload attachment")
		api.RespondError(w, http.StatusInternalServerError, ErrCannotUploadFile.Error())
		return
	}

	api.HandleError(api.RespondOk(w, dto.UploadAttachmentResponse{
		AttachmentURL: attachmentURL,
	}))
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
func (h *Appeal) DeleteAppeal(w http.ResponseWriter, r *http.Request) {
	logger := zerolog.Ctx(r.Context())

	value := r.Context().Value(middleware.UserContextLink{})

	userLink, ok := value.(uuid.UUID)
	if !ok {
		api.RespondError(w, http.StatusUnauthorized, handlerCommon.ErrUserNotAuthorized.Error())
		return
	}

	vars := mux.Vars(r)
	rawAppealLink, ok := vars["link"]
	if !ok {
		api.RespondError(w, http.StatusBadRequest, ErrAppealLinkMissing.Error())
		return
	}

	appealLink, err := uuid.Parse(rawAppealLink)
	if err != nil {
		api.RespondError(w, http.StatusBadRequest, ErrInvalidAppealLink.Error())
		return
	}

	err = h.service.DeleteAppeal(r.Context(), domain.DeleteInfo{
		UserLink:   userLink,
		AppealLink: appealLink,
	})
	if err != nil {
		logger.Error().Err(fmt.Errorf("srv.DeleteAppeal: %w", err)).Msg("failed to delete appeal")
		api.RespondError(w, http.StatusInternalServerError, ErrCannotDeleteAppeal.Error())
		return
	}

	api.HandleError(api.Respond(w, http.StatusOK, api.StatusOK))
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
// @Router       /appeals/stats [get]
func (h *Appeal) GetStats(w http.ResponseWriter, r *http.Request) {
	logger := zerolog.Ctx(r.Context())

	value := r.Context().Value(middleware.UserContextLink{})

	userLink, ok := value.(uuid.UUID)
	if !ok {
		api.RespondError(w, http.StatusUnauthorized, handlerCommon.ErrUserNotAuthorized.Error())
		return
	}

	stats, err := h.service.GetStats(r.Context(), userLink)
	if err != nil {
		logger.Error().Err(err).Msg("cannot get appeal stats")
		api.RespondError(w, http.StatusInternalServerError, ErrCannotGetStats.Error())
		return
	}

	api.HandleError(api.RespondOk(w, dto.AppealsStats{
		OpenAppeals:   stats.OpenAppeals,
		InWorkAppeals: stats.InWorkAppeals,
		CloseAppeals:  stats.CloseAppeals,
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
func (h *Appeal) ChangeAppealStatus(w http.ResponseWriter, r *http.Request) {
	logger := zerolog.Ctx(r.Context())

	value := r.Context().Value(middleware.UserContextLink{})

	userLink, ok := value.(uuid.UUID)
	if !ok {
		api.RespondError(w, http.StatusUnauthorized, handlerCommon.ErrUserNotAuthorized.Error())
		return
	}

	vars := mux.Vars(r)
	rawAppealLink, ok := vars["link"]
	if !ok {
		api.RespondError(w, http.StatusBadRequest, ErrAppealLinkMissing.Error())
		return
	}

	appealLink, err := uuid.Parse(rawAppealLink)
	if err != nil {
		api.RespondError(w, http.StatusBadRequest, ErrInvalidAppealLink.Error())
		return
	}

	var request dto.ChangeAppealStatusInfo
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		logger.Error().Err(err).Msg("can not decode status")
		api.RespondError(w, http.StatusBadRequest, handlerCommon.ErrInvalidRequestSchema.Error())
		return
	}

	err = h.service.ChangeAppealStatus(r.Context(), domain.ChangeAppealStatusInfo{
		UserLink:   userLink,
		AppealLink: appealLink,
		NewStatus:  request.NewStatus,
	})
	if err != nil {
		logger.Error().Err(err).Msg("cannot change appeal status")
		api.RespondError(w, http.StatusInternalServerError, ErrCannotChangeStatus.Error())
		return
	}

	api.HandleError(api.Respond(w, http.StatusOK, api.StatusOK))
}
