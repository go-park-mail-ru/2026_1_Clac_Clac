package handler

import (
	"context"
	"errors"
	"io"
	"net/http"

	"github.com/go-park-mail-ru/2026_1_Clac_Clac/internal/api"
	"github.com/go-park-mail-ru/2026_1_Clac_Clac/internal/middleware"
	"github.com/go-park-mail-ru/2026_1_Clac_Clac/internal/profile/handler/dto"
	serviceDto "github.com/go-park-mail-ru/2026_1_Clac_Clac/internal/profile/service/dto"
	"github.com/google/uuid"
	"github.com/rs/zerolog"
)

const (
	unauthorizedMessage = "unauthorized"
	somethingWentWrong  = "something went wrong"
	tooLargeAvatar      = "size avatar too large"
	invalidFile         = "file is invalid"
	failReadFile        = "can not read file"
	incorrectTypeAvatar = "avatar can be only jpeg/png"
	fileProcessingError = "can not process file"
	failAvatarUrl       = "can not create coorect avatar URL"
	failDeleteFile      = "can not delete avatar"
	incorrectContext    = "context has not correct element"
)

var (
	ErrInvalidRequestSchema = errors.New("invalid schema")
)

type ProfileService interface {
	GetProfileUser(ctx context.Context, userLink uuid.UUID) (serviceDto.UserInfo, error)
	UpdateAvatar(ctx context.Context, userLink uuid.UUID, file io.Reader, mimeType string) (string, error)
	DeleteAvatar(ctx context.Context, userLink uuid.UUID) error
}

func NewHandler(srv ProfileService) *ProfileHandler {
	return &ProfileHandler{
		srv: srv,
	}
}

type ProfileHandler struct {
	srv ProfileService
}

func (ps *ProfileHandler) GetProfile(w http.ResponseWriter, r *http.Request) {
	value := r.Context().Value(middleware.UserContextLink{})

	userLink, ok := value.(uuid.UUID)
	if !ok {
		api.RespondError(w, http.StatusUnauthorized, unauthorizedMessage)
		return
	}

	serviceUser, err := ps.srv.GetProfileUser(r.Context(), userLink)
	if err != nil {
		api.RespondError(w, http.StatusInternalServerError, somethingWentWrong)
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

func (ps *ProfileHandler) UpdateAvatar(w http.ResponseWriter, r *http.Request) {
	logger := zerolog.Ctx(r.Context())

	r.Body = http.MaxBytesReader(w, r.Body, 5<<20)

	if err := r.ParseMultipartForm(5 << 20); err != nil {
		logger.Error().Err(err).Msg(tooLargeAvatar)
		api.RespondError(w, http.StatusBadRequest, tooLargeAvatar)
		return
	}

	file, _, err := r.FormFile("avatar")
	if err != nil {
		logger.Error().Err(err).Msg(invalidFile)
		api.RespondError(w, http.StatusBadRequest, invalidFile)
		return
	}
	defer file.Close()

	signatureFile := make([]byte, 512)
	countSignificantBytes, err := file.Read(signatureFile)
	if err != nil && err != io.EOF {
		logger.Error().Err(err).Msg(failReadFile)
		api.RespondError(w, http.StatusInternalServerError, failReadFile)
		return
	}

	mimeType := http.DetectContentType(signatureFile[:countSignificantBytes])
	if mimeType != "image/png" && mimeType != "image/jpeg" {
		logger.Error().Err(err).Msg(incorrectTypeAvatar)
		api.RespondError(w, http.StatusBadRequest, incorrectTypeAvatar)
		return
	}

	if _, err := file.Seek(0, io.SeekStart); err != nil {
		logger.Error().Err(err).Msg(fileProcessingError)
		api.RespondError(w, http.StatusInternalServerError, fileProcessingError)
		return
	}

	value := r.Context().Value(middleware.UserContextLink{})

	userLink, ok := value.(uuid.UUID)
	if !ok {
		logger.Error().Msg(incorrectContext)
		api.RespondError(w, http.StatusUnauthorized, unauthorizedMessage)
		return
	}

	avatarUrl, err := ps.srv.UpdateAvatar(r.Context(), userLink, file, mimeType)
	if err != nil {
		logger.Error().Err(err).Msg(failAvatarUrl)
		api.RespondError(w, http.StatusInternalServerError, failAvatarUrl)
		return
	}

	avatarResponse := dto.AvatarResponse{
		AvatarURL: avatarUrl,
	}

	api.HandleError(api.RespondOk(w, avatarResponse))
}

func (ps *ProfileHandler) DeleteAvatar(w http.ResponseWriter, r *http.Request) {
	logger := zerolog.Ctx(r.Context())

	value := r.Context().Value(middleware.UserContextLink{})

	userLink, ok := value.(uuid.UUID)
	if !ok {
		logger.Error().Msg(incorrectContext)
		api.RespondError(w, http.StatusUnauthorized, unauthorizedMessage)
		return
	}

	err := ps.srv.DeleteAvatar(r.Context(), userLink)
	if err != nil {
		logger.Error().Err(err).Msg(failDeleteFile)
		api.RespondError(w, http.StatusInternalServerError, failDeleteFile)
		return
	}

	avatarResponse := dto.AvatarResponse{
		AvatarURL: "",
	}

	api.HandleError(api.RespondOk(w, avatarResponse))
}
