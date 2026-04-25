package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"path/filepath"
	"strings"

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
	ErrInvalidActions = errors.New("this role can not do it")

	msgInternalError         = "server error internal"
	ErrInvalidRequestSchema  = errors.New("invalid schema")
	ErrInvalidEmailOrName    = errors.New("incorrect email or name")
	ErrParseMultipartForm    = errors.New("file too large or invalid form")
	ErrCannotFindAttachment  = errors.New("cannot find 'attachment' key")
	ErrCannotReadFile        = errors.New("cannot read file")
	ErrInvalidContentType    = errors.New("invalid content type")
	ErrCannotOperateWithFile = errors.New("cannot operate with file")
	ErrCannotUploadFile      = errors.New("cannot upload attachment")
)

type AppealService interface {
	CreateAppeal(ctx context.Context, appeal serviceDto.EntityAppeal) error
	GetAppeals(ctx context.Context, userLink uuid.UUID) (serviceDto.Appeals, error)
	DeleteAppeal(ctx context.Context, appealLink uuid.UUID) error
	GetStats(ctx context.Context, userLink uuid.UUID) (serviceDto.AppealStats, error)
	ChangeAppealStatus(ctx context.Context, info serviceDto.ChangeAppealStatusInfo) error
	UploadAttachment(ctx context.Context, file io.Reader, contentType, extension string, appealLink uuid.UUID) (string, error)
}

type Config struct {
	MultipartAttachmentFileKey string
	MaxAttachmentSize          int64
	AttachmentBaseURL          string
}

type Handler struct {
	srv  AppealService
	conf Config
}

func NewHandler(srv AppealService, conf Config) *Handler {
	return &Handler{
		srv:  srv,
		conf: conf,
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

		if errors.Is(err, common.ErrInvalidCategory) {
			api.RespondError(w, http.StatusBadRequest, common.ErrInvalidCategory.Error())
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
		attachmentKey := a.AttachmentKey
		if attachmentKey != "" {
			attachmentKey = fmt.Sprintf("%s/%s", h.conf.AttachmentBaseURL, attachmentKey)
		}

		response.Appeals = append(response.Appeals, dto.Appeal{
			AppealID:      a.AppealID,
			AppealLink:    a.AppealLink,
			Email:         a.Email,
			DisplayName:   a.DisplayName,
			Category:      a.Category,
			Status:        a.Status,
			Description:   a.Description,
			AttachmentKey: attachmentKey,
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
func (h *Handler) UploadAttachment(w http.ResponseWriter, r *http.Request) {
	logger := zerolog.Ctx(r.Context())

	value := r.Context().Value(middleware.UserContextLink{})
	_, ok := value.(uuid.UUID)
	if !ok {
		api.RespondError(w, http.StatusUnauthorized, "unauthorised")
		return
	}

	vars := mux.Vars(r)
	rawAppealLink, ok := vars["link"]
	if !ok {
		api.RespondError(w, http.StatusBadRequest, "appeal link missing")
		return
	}

	appealLink, err := uuid.Parse(rawAppealLink)
	if err != nil {
		api.RespondError(w, http.StatusBadRequest, "invalid appeal link format")
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

	data, err := io.ReadAll(file)
	if err != nil {
		api.RespondError(w, http.StatusInternalServerError, ErrCannotReadFile.Error())
		return
	}

	contentType := http.DetectContentType(data)
	if !strings.HasPrefix(contentType, "image/") {
		api.RespondError(w, http.StatusBadRequest, ErrInvalidContentType.Error())
		return
	}

	extension := filepath.Ext(header.Filename)

	key, err := h.srv.UploadAttachment(r.Context(), bytes.NewReader(data), contentType, extension, appealLink)
	if err != nil {
		logger.Error().Err(fmt.Errorf("srv.UploadAttachment: %w", err)).Msg("failed to upload attachment")

		if errors.Is(err, common.ErrorAppealNotFound) {
			api.RespondError(w, http.StatusNotFound, common.ErrorAppealNotFound.Error())
			return
		}

		api.RespondError(w, http.StatusInternalServerError, ErrCannotUploadFile.Error())
		return
	}

	api.HandleError(api.RespondOk(w, dto.UploadAttachmentResponse{
		AttachmentURL: fmt.Sprintf("%s/%s", h.conf.AttachmentBaseURL, key),
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
