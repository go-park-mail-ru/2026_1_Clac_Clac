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
	"github.com/go-park-mail-ru/2026_1_Clac_Clac/facade/internal/delivery/http/middleware"
	"github.com/go-park-mail-ru/2026_1_Clac_Clac/facade/internal/domain"
	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/rs/zerolog"
)

const (
	nameAvatarBlock = "avatar"

	msgUnauthorized      = "unauthorized"
	msgUserNotFound      = "user not found"
	msgFailGetProfile    = "cannot get user profile"
	msgFailUpdateProfile = "cannot update profile"
	msgTooLargeAvatar    = "avatar file too large"
	msgInvalidFile       = "file is invalid"
	msgIncorrectType     = "avatar must be jpeg/png/jpg/webp"
	msgFailProcessFile   = "cannot process file"
	msgFailUpdateAvatar  = "cannot update avatar"
	msgFailDeleteAvatar  = "cannot delete avatar"
)

type ProfileUseCase interface {
	GetProfile(ctx context.Context, userLink uuid.UUID) (domain.FullInfoUser, error)
	UpdateProfile(ctx context.Context, info domain.UpdatedInfo) error
	UpdateAvatar(ctx context.Context, info domain.AvatarInfo) (string, error)
	DeleteAvatar(ctx context.Context, userLink uuid.UUID) error
}

type ProfileConfig struct {
	ValidExtensions       map[string]struct{}
	SignatureTypeBytes    int
	MaxLenNameUser        int
	MaxLenDescriptionUser int
	MaxReadBytes          int64
}

type Profile struct {
	usecase ProfileUseCase
	cfg     ProfileConfig
}

func NewProfileHandler(usecase ProfileUseCase, cfg ProfileConfig) *Profile {
	return &Profile{usecase: usecase, cfg: cfg}
}

// GetProfile возвращает профиль текущего пользователя
//
//	@Summary		Получить свой профиль
//	@Tags			Profile
//	@Security		sessionCookie
//	@Security		csrfToken
//	@Produce		json
//	@Success		200	{object}	dto.ProfileResponse
//	@Failure		401	{object}	api.ErrorResponse	"unauthorized"
//	@Failure		404	{object}	api.ErrorResponse	"user not found"
//	@Failure		500	{object}	api.ErrorResponse	"cannot get user profile"
//	@Router			/api/profiles [get]
func (p *Profile) GetProfile(w http.ResponseWriter, r *http.Request) {
	value := r.Context().Value(middleware.UserContextLink{})
	userLink, ok := value.(uuid.UUID)
	if !ok {
		api.RespondError(w, http.StatusUnauthorized, msgUnauthorized)
		return
	}

	user, err := p.usecase.GetProfile(r.Context(), userLink)
	if err != nil {
		if errors.Is(err, common.ErrorNonexistentUser) {
			api.RespondError(w, http.StatusNotFound, msgUserNotFound)
			return
		}
		api.RespondError(w, http.StatusInternalServerError, msgFailGetProfile)
		return
	}

	api.HandleError(api.RespondOk(w, convertToProfileResponse(user)))
}

// GetProfileByLink возвращает профиль пользователя по UUID
//
//	@Summary		Получить профиль по ссылке
//	@Tags			Profile
//	@Security		sessionCookie
//	@Security		csrfToken
//	@Produce		json
//	@Param			user_link	path		string	true	"UUID пользователя"
//	@Success		200			{object}	dto.ProfileResponse
//	@Failure		400			{object}	api.ErrorResponse	"invalid user link"
//	@Failure		404			{object}	api.ErrorResponse	"user not found"
//	@Failure		500			{object}	api.ErrorResponse	"cannot get user profile"
//	@Router			/api/profiles/{user_link} [get]
func (p *Profile) GetProfileByLink(w http.ResponseWriter, r *http.Request) {
	logger := zerolog.Ctx(r.Context())

	userLinkParam := mux.Vars(r)["user_link"]
	userLink, err := uuid.Parse(userLinkParam)
	if err != nil {
		api.RespondError(w, http.StatusBadRequest, "invalid user link")
		return
	}

	user, err := p.usecase.GetProfile(r.Context(), userLink)
	if err != nil {
		if errors.Is(err, common.ErrorNonexistentUser) {
			api.RespondError(w, http.StatusNotFound, msgUserNotFound)
			return
		}
		logger.Error().Err(err).Msg("profile.GetProfileByLink")
		api.RespondError(w, http.StatusInternalServerError, msgFailGetProfile)
		return
	}

	api.HandleError(api.RespondOk(w, convertToProfileResponse(user)))
}

// UpdateProfile обновляет текстовые данные профиля (имя, описание)
//
//	@Summary		Обновить профиль
//	@Tags			Profile
//	@Security		sessionCookie
//	@Security		csrfToken
//	@Accept			json
//	@Produce		json
//	@Param			input	body		dto.UpdateProfileRequest	true	"Новые данные"
//	@Success		200		{string}	string						"OK"
//	@Failure		400		{object}	api.ErrorResponse			"incorrect name/description"
//	@Failure		401		{object}	api.ErrorResponse			"unauthorized"
//	@Failure		500		{object}	api.ErrorResponse			"cannot update profile"
//	@Router			/api/profiles/info [post]
func (p *Profile) UpdateProfile(w http.ResponseWriter, r *http.Request) {
	value := r.Context().Value(middleware.UserContextLink{})
	userLink, ok := value.(uuid.UUID)
	if !ok {
		api.RespondError(w, http.StatusUnauthorized, msgUnauthorized)
		return
	}

	var req dto.UpdateProfileRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		api.RespondError(w, http.StatusBadRequest, handlerCommon.ErrInvalidRequestSchema.Error())
		return
	}

	if err := common.ValidateTextInfo(req.DisplayName, p.cfg.MaxLenNameUser); err != nil {
		api.RespondError(w, http.StatusBadRequest, fmt.Sprintf("incorrect name: %s", err.Error()))
		return
	}

	if err := common.ValidateTextInfo(req.Description, p.cfg.MaxLenDescriptionUser); err != nil {
		api.RespondError(w, http.StatusBadRequest, fmt.Sprintf("incorrect description: %s", err.Error()))
		return
	}

	err := p.usecase.UpdateProfile(r.Context(), domain.UpdatedInfo{
		UserLink:    userLink,
		DisplayName: req.DisplayName,
		Description: req.Description,
	})
	if err != nil {
		if errors.Is(err, common.ErrorMissingRequiredField) {
			api.RespondError(w, http.StatusBadRequest, common.ErrorMissingRequiredField.Error())
			return
		}
		if errors.Is(err, common.ErrorInvalidProfileData) {
			api.RespondError(w, http.StatusBadRequest, common.ErrorInvalidProfileData.Error())
			return
		}
		api.RespondError(w, http.StatusInternalServerError, msgFailUpdateProfile)
		return
	}

	api.HandleError(api.RespondOk(w, api.StatusOK))
}

// UpdateAvatar загружает новый аватар
//
//	@Summary		Обновить аватар
//	@Tags			Profile
//	@Security		sessionCookie
//	@Security		csrfToken
//	@Accept			multipart/form-data
//	@Produce		json
//	@Param			avatar	formData	file	true	"Файл изображения (jpg/jpeg/png/webp)"
//	@Success		200		{object}	dto.AvatarResponse
//	@Failure		400		{object}	api.ErrorResponse	"avatar file too large or invalid"
//	@Failure		401		{object}	api.ErrorResponse	"unauthorized"
//	@Failure		415		{object}	api.ErrorResponse	"avatar must be jpeg/png/jpg/webp"
//	@Failure		500		{object}	api.ErrorResponse	"cannot update avatar"
//	@Router			/api/profiles/avatar [put]
func (p *Profile) UpdateAvatar(w http.ResponseWriter, r *http.Request) {
	logger := zerolog.Ctx(r.Context())

	value := r.Context().Value(middleware.UserContextLink{})
	userLink, ok := value.(uuid.UUID)
	if !ok {
		api.RespondError(w, http.StatusUnauthorized, msgUnauthorized)
		return
	}

	if err := r.ParseMultipartForm(p.cfg.MaxReadBytes); err != nil {
		logger.Error().Err(err).Msg(msgTooLargeAvatar)
		api.RespondError(w, http.StatusBadRequest, msgTooLargeAvatar)
		return
	}

	defer r.MultipartForm.RemoveAll()

	file, header, err := r.FormFile(nameAvatarBlock)
	if err != nil {
		api.RespondError(w, http.StatusBadRequest, msgInvalidFile)
		return
	}
	defer func() {
		if err := file.Close(); err != nil {
			logger.Err(err).Msg("close avatar file")
		}
	}()

	sigBuf := make([]byte, p.cfg.SignatureTypeBytes)
	n, err := file.Read(sigBuf)
	if err != nil && !errors.Is(err, io.EOF) {
		api.RespondError(w, http.StatusInternalServerError, msgInvalidFile)
		return
	}

	mimeType := http.DetectContentType(sigBuf[:n])
	if _, ok := p.cfg.ValidExtensions[mimeType]; !ok {
		api.RespondError(w, http.StatusUnsupportedMediaType, msgIncorrectType)
		return
	}

	if _, err := file.Seek(0, io.SeekStart); err != nil {
		api.RespondError(w, http.StatusUnprocessableEntity, msgFailProcessFile)
		return
	}

	fileData, err := io.ReadAll(file)
	if err != nil {
		api.RespondError(w, http.StatusInternalServerError, msgFailProcessFile)
		return
	}

	ext := ""
	if header != nil {
		ext = header.Filename
	}

	avatarURL, err := p.usecase.UpdateAvatar(r.Context(), domain.AvatarInfo{
		UserLink:      userLink,
		FileData:      fileData,
		ContentType:   mimeType,
		FileExtension: ext,
	})
	if err != nil {
		if errors.Is(err, common.ErrorNonexistentUser) {
			api.RespondError(w, http.StatusNotFound, msgUserNotFound)
			return
		}
		logger.Error().Err(err).Msg(msgFailUpdateAvatar)
		api.RespondError(w, http.StatusInternalServerError, msgFailUpdateAvatar)
		return
	}

	api.HandleError(api.RespondOk(w, dto.AvatarResponse{AvatarURL: avatarURL}))
}

// DeleteAvatar удаляет аватар пользователя
//
//	@Summary		Удалить аватар
//	@Tags			Profile
//	@Security		sessionCookie
//	@Security		csrfToken
//	@Produce		json
//	@Success		200	{string}	string				"OK"
//	@Failure		401	{object}	api.ErrorResponse	"unauthorized"
//	@Failure		404	{object}	api.ErrorResponse	"user not found"
//	@Failure		500	{object}	api.ErrorResponse	"cannot delete avatar"
//	@Router			/api/profiles/avatar [delete]
func (p *Profile) DeleteAvatar(w http.ResponseWriter, r *http.Request) {
	logger := zerolog.Ctx(r.Context())

	value := r.Context().Value(middleware.UserContextLink{})
	userLink, ok := value.(uuid.UUID)
	if !ok {
		api.RespondError(w, http.StatusUnauthorized, msgUnauthorized)
		return
	}

	if err := p.usecase.DeleteAvatar(r.Context(), userLink); err != nil {
		if errors.Is(err, common.ErrorNonexistentUser) {
			api.RespondError(w, http.StatusNotFound, msgUserNotFound)
			return
		}
		logger.Error().Err(err).Msg(msgFailDeleteAvatar)
		api.RespondError(w, http.StatusInternalServerError, msgFailDeleteAvatar)
		return
	}

	api.HandleError(api.RespondOk(w, api.StatusOK))
}

func convertToProfileResponse(u domain.FullInfoUser) dto.ProfileResponse {
	return dto.ProfileResponse{
		Link:        u.UserLink,
		DisplayName: u.DisplayName,
		Description: u.Description,
		Email:       u.Email,
		AvatarURL:   u.AvatarURL,
	}
}
