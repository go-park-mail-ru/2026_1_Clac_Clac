package handler

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"

	"github.com/go-park-mail-ru/2026_1_Clac_Clac/internal/api"
	"github.com/go-park-mail-ru/2026_1_Clac_Clac/internal/common"
	"github.com/go-park-mail-ru/2026_1_Clac_Clac/internal/middleware"
	"github.com/go-park-mail-ru/2026_1_Clac_Clac/internal/profile/handler/dto"
	serviceDto "github.com/go-park-mail-ru/2026_1_Clac_Clac/internal/profile/service/dto"
	"github.com/google/uuid"
	"github.com/rs/zerolog"
)

const (
	unauthorizedMessage = "unauthorized"
	failGetInfoUser     = "can not get info user"
	tooLargeAvatar      = "size avatar too large"
	invalidFile         = "file is invalid"
	failReadFile        = "can not read file"
	incorrectTypeAvatar = "avatar can be only jpeg/png/jpg/webp"
	fileProcessingError = "can not process file"
	failAvatarUrl       = "can not create correct avatar URL"
	failDeleteFile      = "can not delete avatar"
	incorrectContext    = "context has not correct element"
	incorrectCloseFile  = "fail close file"
	failFoundUser       = "user not found"
	failUpdateUserInfo  = "can not update name or description"

	nameAvatarBlock = "avatar"
)

var (
	ErrInvalidRequestSchema = errors.New("invalid schema")
)

// mockery --name=ProfileService --output=mock_profile_srv --outpkg=mockProfileSrv
type ProfileService interface {
	GetProfileUser(ctx context.Context, userLink uuid.UUID) (serviceDto.UserInfo, error)
	UpdateProfile(ctx context.Context, updatedInfo serviceDto.UpdatedUserInfo) error
	UpdateAvatar(ctx context.Context, avatar serviceDto.UpdatedAvatar) (string, error)
	DeleteAvatar(ctx context.Context, userLink uuid.UUID) error
}

type Deps struct {
	Srv                   ProfileService
	ValidExtensions       map[string]struct{}
	SiganatureTypeBytes   int
	MaxLenNameUser        int
	MaxLenDescriptionUser int
	MaxReadBytes          int64
}

func NewHandler(deps Deps) *Handler {
	return &Handler{
		deps: deps,
	}
}

type Handler struct {
	deps Deps
}

func (h *Handler) GetProfile(w http.ResponseWriter, r *http.Request) {
	value := r.Context().Value(middleware.UserContextLink{})

	userLink, ok := value.(uuid.UUID)
	if !ok {
		api.RespondError(w, http.StatusUnauthorized, unauthorizedMessage)
		return
	}

	serviceUser, err := h.deps.Srv.GetProfileUser(r.Context(), userLink)
	if err != nil {
		if errors.Is(err, common.ErrorNonexistentUser) {
			api.RespondError(w, http.StatusNotFound, failFoundUser)
			return
		}

		api.RespondError(w, http.StatusInternalServerError, failGetInfoUser)
		return
	}

	user := dto.UserInfoResponse{
		Link:        serviceUser.Link,
		DisplayName: serviceUser.DisplayName,
		Email:       serviceUser.Email,
		AvatarURL:   serviceUser.AvatarURL,
	}

	api.HandleError(api.RespondOk(w, user))
}

func (h *Handler) UpdateProfile(w http.ResponseWriter, r *http.Request) {
	logger := zerolog.Ctx(r.Context())

	value := r.Context().Value(middleware.UserContextLink{})
	userLink, ok := value.(uuid.UUID)
	if !ok {
		api.RespondError(w, http.StatusUnauthorized, unauthorizedMessage)
		return
	}

	var updatedInfo dto.UpdatedInfo
	err := json.NewDecoder(r.Body).Decode(&updatedInfo)
	if err != nil {
		api.HandleError(api.RespondError(w, http.StatusBadRequest, common.IncorrectFormatRequest))
		return
	}

	err = common.ValidateTextInfo(updatedInfo.DisplayName, h.deps.MaxLenNameUser)
	if err != nil {
		api.RespondError(w, http.StatusBadRequest, fmt.Sprintf("incorrect name: %s", err.Error()))
		return
	}

	err = common.ValidateTextInfo(updatedInfo.DescriptionUser, h.deps.MaxLenDescriptionUser)
	if err != nil {
		api.RespondError(w, http.StatusBadRequest, fmt.Sprintf("incorrect description: %s", err.Error()))
		return
	}

	userInfo := serviceDto.UpdatedUserInfo{
		Link:        userLink,
		DisplayName: updatedInfo.DisplayName,
		Description: updatedInfo.DescriptionUser,
	}

	err = h.deps.Srv.UpdateProfile(r.Context(), userInfo)
	if err != nil {
		logger.Err(fmt.Errorf("srv.UpdateProfile: %w", err))
		api.RespondError(w, http.StatusInternalServerError, failUpdateUserInfo)
		return
	}

	api.HandleError(api.RespondOk(w, api.StatusOK))
}

func (h *Handler) UpdateAvatar(w http.ResponseWriter, r *http.Request) {
	logger := zerolog.Ctx(r.Context())

	if err := r.ParseMultipartForm(h.deps.MaxReadBytes); err != nil {
		logger.Error().Err(err).Msg(tooLargeAvatar)
		api.RespondError(w, http.StatusBadRequest, tooLargeAvatar)
		return
	}

	file, _, err := r.FormFile(nameAvatarBlock)
	if err != nil {
		logger.Error().Err(err).Msg(invalidFile)
		api.RespondError(w, http.StatusBadRequest, invalidFile)
		return
	}
	defer func() {
		if err := file.Close(); err != nil {
			logger.Err(err).Msg(incorrectCloseFile)
		}
	}()

	signatureFile := make([]byte, h.deps.SiganatureTypeBytes)
	countSignificantBytes, err := file.Read(signatureFile)
	if err != nil && err != io.EOF {
		logger.Error().Err(err).Msg(failReadFile)
		api.RespondError(w, http.StatusInternalServerError, failReadFile)
		return
	}

	mimeType := http.DetectContentType(signatureFile[:countSignificantBytes])
	if _, ok := h.deps.ValidExtensions[mimeType]; !ok {
		logger.Error().Err(err).Msg(incorrectTypeAvatar)
		api.RespondError(w, http.StatusBadRequest, incorrectTypeAvatar)
		return
	}

	if _, err := file.Seek(0, io.SeekStart); err != nil {
		logger.Error().Err(err).Msg(fileProcessingError)
		api.RespondError(w, http.StatusUnprocessableEntity, fileProcessingError)
		return
	}

	value := r.Context().Value(middleware.UserContextLink{})

	userLink, ok := value.(uuid.UUID)
	if !ok {
		logger.Error().Msg(incorrectContext)
		api.RespondError(w, http.StatusUnauthorized, unauthorizedMessage)
		return
	}

	avatarUrl, err := h.deps.Srv.UpdateAvatar(r.Context(), serviceDto.UpdatedAvatar{
		UserLink: userLink,
		File:     file,
		MimeType: mimeType,
	})
	if err != nil {
		if errors.Is(err, common.ErrorNonexistentUser) {
			api.RespondError(w, http.StatusNotFound, failDeleteFile)
			return
		}

		logger.Error().Err(err).Msg(failAvatarUrl)
		api.RespondError(w, http.StatusInternalServerError, failAvatarUrl)
		return
	}

	avatarResponse := dto.AvatarResponse{
		AvatarURL: avatarUrl,
	}

	api.HandleError(api.RespondOk(w, avatarResponse))
}

func (h *Handler) DeleteAvatar(w http.ResponseWriter, r *http.Request) {
	logger := zerolog.Ctx(r.Context())

	value := r.Context().Value(middleware.UserContextLink{})

	userLink, ok := value.(uuid.UUID)
	if !ok {
		logger.Error().Msg(incorrectContext)
		api.RespondError(w, http.StatusUnauthorized, unauthorizedMessage)
		return
	}

	err := h.deps.Srv.DeleteAvatar(r.Context(), userLink)
	if err != nil {
		if errors.Is(err, common.ErrorNonexistentUser) {
			api.RespondError(w, http.StatusNotFound, failFoundUser)
			return
		}

		logger.Error().Err(err).Msg(failDeleteFile)
		api.RespondError(w, http.StatusInternalServerError, failDeleteFile)
		return
	}

	api.HandleError(api.RespondOk(w, api.StatusOK))
}
