package handlers

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"path/filepath"

	"github.com/go-park-mail-ru/2026_1_Clac_Clac/facade/internal/api"
	"github.com/go-park-mail-ru/2026_1_Clac_Clac/facade/internal/common"
	"github.com/go-park-mail-ru/2026_1_Clac_Clac/facade/internal/delivery/http/dto"
	handlerCommon "github.com/go-park-mail-ru/2026_1_Clac_Clac/facade/internal/delivery/http/handlers/common"
	"github.com/go-park-mail-ru/2026_1_Clac_Clac/facade/internal/domain"
	"github.com/go-park-mail-ru/2026_1_Clac_Clac/facade/internal/middleware"
	sentryLogger "github.com/go-park-mail-ru/2026_1_Clac_Clac/pkg/logger"
	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/mailru/easyjson"
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
	ResetPassword(ctx context.Context, updatedPassword domain.UpdatedPassword) error
}

type ProfileConfig struct {
	ValidExtensions       map[string]struct{}
	SignatureTypeBytes    int
	MaxLenNameUser        int
	MaxLenDescriptionUser int
	MaxReadBytes          int64
	MaxLenPassword        int
	MinLenPassword        int
}

type Profile struct {
	profile    ProfileUseCase
	mailSender MailSenderUsecase
	cfg        ProfileConfig
}

func NewProfileHandler(profile ProfileUseCase, mailSender MailSenderUsecase, cfg ProfileConfig) *Profile {
	return &Profile{
		profile:    profile,
		mailSender: mailSender,
		cfg:        cfg,
	}
}

// GetProfile возвращает профиль текущего пользователя
//
//	@Summary		Получить свой профиль
//	@Description	Возвращает полные данные профиля авторизованного пользователя: имя, описание, email, ссылка на аватар.
//	@Tags			Profiles
//	@Security		sessionCookie
//	@Produce		json
//	@Success		200	{object}	api.OkResponse[dto.ProfileResponse]	"Профиль пользователя"
//	@Failure		401	{object}	api.ErrorResponse					"Пользователь не авторизован"
//	@Failure		404	{object}	api.ErrorResponse					"Пользователь не найден"
//	@Failure		500	{object}	api.ErrorResponse					"Внутренняя ошибка сервера при получении профиля"
//	@Router			/profiles [get]
func (p *Profile) GetProfile(w http.ResponseWriter, r *http.Request) {
	logger := zerolog.Ctx(r.Context())

	value := r.Context().Value(middleware.UserContextLink{})
	userLink, ok := value.(uuid.UUID)
	if !ok {
		api.RespondError(w, http.StatusUnauthorized, msgUnauthorized)
		return
	}

	user, err := p.profile.GetProfile(r.Context(), userLink)
	if err != nil {
		if errors.Is(err, common.ErrorNonexistentUser) {
			api.RespondError(w, http.StatusNotFound, msgUserNotFound)
			return
		}
		errLog := fmt.Errorf("profile.GetProfile: %w", err)
		logger.Error().Err(errLog).Msg("profile.GetProfile failed")
		sentryLogger.CaptureFromContext(r.Context(), errLog, "GetProfile", map[string]interface{}{
			"user_link": userLink,
			"action":    "get_profile",
		})
		api.RespondError(w, http.StatusInternalServerError, msgFailGetProfile)
		return
	}

	api.HandleError(api.RespondOk(w, convertToProfileResponse(user)))
}

// GetProfileByLink возвращает профиль пользователя по UUID
//
//	@Summary		Получить профиль по ссылке
//	@Description	Возвращает публичный профиль любого пользователя по его UUID.
//	@Tags			Profiles
//	@Security		sessionCookie
//	@Produce		json
//	@Param			user_link	path		string								true	"UUID пользователя (формат: xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx)"
//	@Success		200			{object}	api.OkResponse[dto.ProfileResponse]	"Профиль пользователя"
//	@Failure		400			{object}	api.ErrorResponse					"Некорректный формат UUID пользователя"
//	@Failure		401			{object}	api.ErrorResponse					"Пользователь не авторизован"
//	@Failure		404			{object}	api.ErrorResponse					"Пользователь не найден"
//	@Failure		500			{object}	api.ErrorResponse					"Внутренняя ошибка сервера"
//	@Router			/profiles/{user_link} [get]
func (p *Profile) GetProfileByLink(w http.ResponseWriter, r *http.Request) {
	logger := zerolog.Ctx(r.Context())

	userLinkParam := mux.Vars(r)["user_link"]
	userLink, err := uuid.Parse(userLinkParam)
	if err != nil {
		api.RespondError(w, http.StatusBadRequest, "invalid user link format")
		return
	}

	user, err := p.profile.GetProfile(r.Context(), userLink)
	if err != nil {
		if errors.Is(err, common.ErrorNonexistentUser) {
			api.RespondError(w, http.StatusNotFound, msgUserNotFound)
			return
		}
		errLog := fmt.Errorf("profile.GetProfile: %w", err)
		logger.Error().Err(errLog).Msg("profile.GetProfileByLink failed")
		sentryLogger.CaptureFromContext(r.Context(), errLog, "GetProfileByLink", map[string]interface{}{
			"user_link": userLink,
			"action":    "get_profile_by_link",
		})
		api.RespondError(w, http.StatusInternalServerError, msgFailGetProfile)
		return
	}

	api.HandleError(api.RespondOk(w, convertToProfileResponse(user)))
}

// UpdateProfile обновляет текстовые данные профиля (имя, описание)
//
//	@Summary		Обновить профиль
//	@Description	Изменяет display_name и description_user. Требует валидный CSRF-токен.
//	@Tags			Profiles
//	@Security		sessionCookie
//	@Security		csrfToken
//	@Accept			json
//	@Produce		json
//	@Param			input	body		dto.UpdateProfileRequest	true	"Новые имя и описание"
//	@Success		200		{object}	api.Response				"Профиль успешно обновлён"
//	@Failure		400		{object}	api.ErrorResponse			"Некорректные данные: отсутствует обязательное поле или превышена длина"
//	@Failure		401		{object}	api.ErrorResponse			"Пользователь не авторизован"
//	@Failure		500		{object}	api.ErrorResponse			"Ошибка обновления профиля"
//	@Router			/profiles/info [post]
func (p *Profile) UpdateProfile(w http.ResponseWriter, r *http.Request) {
	logger := zerolog.Ctx(r.Context())

	value := r.Context().Value(middleware.UserContextLink{})
	userLink, ok := value.(uuid.UUID)
	if !ok {
		api.RespondError(w, http.StatusUnauthorized, msgUnauthorized)
		return
	}

	var req dto.UpdateProfileRequest
	if err := easyjson.UnmarshalFromReader(r.Body, &req); err != nil {
		api.RespondError(w, http.StatusBadRequest, handlerCommon.ErrInvalidRequestSchema.Error())
		return
	}

	req.Sanitize()

	if err := common.ValidateTextInfo(req.DisplayName, p.cfg.MaxLenNameUser); err != nil {
		api.RespondError(w, http.StatusBadRequest, fmt.Sprintf("incorrect name: %s", err.Error()))
		return
	}

	if err := common.ValidateTextInfo(req.Description, p.cfg.MaxLenDescriptionUser); err != nil {
		api.RespondError(w, http.StatusBadRequest, fmt.Sprintf("incorrect description: %s", err.Error()))
		return
	}

	err := p.profile.UpdateProfile(r.Context(), domain.UpdatedInfo{
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
		errLog := fmt.Errorf("profile.UpdateProfile: %w", err)
		logger.Error().Err(errLog).Msg("profile.UpdateProfile failed")
		sentryLogger.CaptureFromContext(r.Context(), errLog, "UpdateProfile", map[string]interface{}{
			"user_link": userLink,
			"action":    "update_profile",
		})
		api.RespondError(w, http.StatusInternalServerError, msgFailUpdateProfile)
		return
	}

	api.HandleError(api.RespondOk(w, api.StatusOK))
}

// UpdateAvatar загружает новый аватар
//
//	@Summary		Обновить аватар
//	@Description	Загружает новое изображение. Допустимые форматы определяются по magic bytes.
//	@Tags			Profiles
//	@Security		sessionCookie
//	@Security		csrfToken
//	@Accept			multipart/form-data
//	@Produce		json
//	@Param			avatar	formData	file								true	"Файл изображения (поле: avatar)"
//	@Success		200		{object}	api.OkResponse[dto.AvatarResponse]	"Аватар загружен, возвращается URL"
//	@Failure		400		{object}	api.ErrorResponse					"Файл слишком большой или отсутствует поле avatar"
//	@Failure		401		{object}	api.ErrorResponse					"Пользователь не авторизован"
//	@Failure		404		{object}	api.ErrorResponse					"Пользователь не найден"
//	@Failure		415		{object}	api.ErrorResponse					"Недопустимый тип файла"
//	@Failure		422		{object}	api.ErrorResponse					"Невозможно обработать/прочитать файл"
//	@Failure		500		{object}	api.ErrorResponse					"Ошибка на сервере при сохранении аватара"
//	@Router			/profiles/avatar [put]
func (p *Profile) UpdateAvatar(w http.ResponseWriter, r *http.Request) {
	logger := zerolog.Ctx(r.Context())

	value := r.Context().Value(middleware.UserContextLink{})
	userLink, ok := value.(uuid.UUID)
	if !ok {
		api.RespondError(w, http.StatusUnauthorized, msgUnauthorized)
		return
	}

	if err := r.ParseMultipartForm(p.cfg.MaxReadBytes); err != nil {
		logger.Error().Err(err).Msg("parse multipart form")
		var maxBytesErr *http.MaxBytesError
		if errors.As(err, &maxBytesErr) {
			api.RespondError(w, http.StatusRequestEntityTooLarge, msgTooLargeAvatar)
		} else {
			api.RespondError(w, http.StatusBadRequest, msgTooLargeAvatar)
		}
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
		errLog := fmt.Errorf("file.Read signature: %w", err)
		logger.Error().Err(errLog).Msg("failed to read signature bytes")
		sentryLogger.CaptureFromContext(r.Context(), errLog, "UpdateAvatar", map[string]interface{}{
			"user_link": userLink,
			"action":    "read_signature_bytes",
		})
		api.RespondError(w, http.StatusInternalServerError, msgInvalidFile)
		return
	}

	mimeType := http.DetectContentType(sigBuf[:n])
	if _, ok := p.cfg.ValidExtensions[mimeType]; !ok {
		api.RespondError(w, http.StatusUnsupportedMediaType, msgIncorrectType)
		return
	}

	if _, err := file.Seek(0, io.SeekStart); err != nil {
		logger.Error().Err(err).Msg("failed to seek file")
		api.RespondError(w, http.StatusUnprocessableEntity, msgFailProcessFile)
		return
	}

	ext := ""
	if header != nil {
		ext = filepath.Ext(header.Filename)
	}

	avatarURL, err := p.profile.UpdateAvatar(r.Context(), domain.AvatarInfo{
		UserLink:      userLink,
		FileData:      file,
		ContentType:   mimeType,
		FileExtension: ext,
	})
	if err != nil {
		if errors.Is(err, common.ErrorNonexistentUser) {
			api.RespondError(w, http.StatusNotFound, msgUserNotFound)
			return
		}
		errLog := fmt.Errorf("profile.UpdateAvatar: %w", err)
		logger.Error().Err(errLog).Msg(msgFailUpdateAvatar)
		sentryLogger.CaptureFromContext(r.Context(), errLog, "UpdateAvatar", map[string]interface{}{
			"user_link": userLink,
			"action":    "update_avatar",
		})
		api.RespondError(w, http.StatusInternalServerError, msgFailUpdateAvatar)
		return
	}

	api.HandleError(api.RespondOk(w, dto.AvatarResponse{AvatarURL: avatarURL}))
}

// DeleteAvatar удаляет аватар пользователя
//
//	@Summary		Удалить аватар
//	@Description	Удаляет текущий аватар пользователя из хранилища и сбрасывает поле avatar_url в профиле. Если аватара не было, всё равно возвращает 200.
//	@Tags			Profiles
//	@Security		sessionCookie
//	@Security		csrfToken
//	@Produce		json
//	@Success		200	{object}	api.Response		"Аватар удалён"
//	@Failure		401	{object}	api.ErrorResponse	"Пользователь не авторизован"
//	@Failure		404	{object}	api.ErrorResponse	"Пользователь не найден"
//	@Failure		500	{object}	api.ErrorResponse	"Ошибка удаления аватара"
//	@Router			/profiles/avatar [delete]
func (p *Profile) DeleteAvatar(w http.ResponseWriter, r *http.Request) {
	logger := zerolog.Ctx(r.Context())

	value := r.Context().Value(middleware.UserContextLink{})
	userLink, ok := value.(uuid.UUID)
	if !ok {
		api.RespondError(w, http.StatusUnauthorized, msgUnauthorized)
		return
	}

	if err := p.profile.DeleteAvatar(r.Context(), userLink); err != nil {
		if errors.Is(err, common.ErrorNonexistentUser) {
			api.RespondError(w, http.StatusNotFound, msgUserNotFound)
			return
		}
		errLog := fmt.Errorf("profile.DeleteAvatar: %w", err)
		logger.Error().Err(errLog).Msg(msgFailDeleteAvatar)
		sentryLogger.CaptureFromContext(r.Context(), errLog, "DeleteAvatar", map[string]interface{}{
			"user_link": userLink,
			"action":    "delete_avatar",
		})
		api.RespondError(w, http.StatusInternalServerError, msgFailDeleteAvatar)
		return
	}

	api.HandleError(api.RespondOk(w, api.StatusOK))
}

// ResetUserPassword устанавливает новый пароль
//
//	@Summary		Сброс пароля
//	@Description	Устанавливает новый пароль, используя одноразовый token_id (код из письма). Токен инвалидируется после использования.
//	@Tags			Auth
//	@Accept			json
//	@Produce		json
//	@Param			input	body		dto.NewPasswordRequest	true	"Новый пароль и одноразовый токен"
//	@Success		200		{object}	api.Response			"Пароль успешно изменён"
//	@Failure		400		{object}	api.ErrorResponse		"Пароли не совпадают, некорректная длина или токен не найден"
//	@Failure		404		{object}	api.ErrorResponse		"Токен не существует/истёк или пользователь не найден"
//	@Failure		500		{object}	api.ErrorResponse		"Внутренняя ошибка сервера при смене пароля"
//	@Router			/reset-password [post]
func (p *Profile) ResetUserPassword(w http.ResponseWriter, r *http.Request) {
	logger := zerolog.Ctx(r.Context())

	var request dto.NewPasswordRequest
	if err := easyjson.UnmarshalFromReader(r.Body, &request); err != nil {
		api.RespondError(w, http.StatusBadRequest, handlerCommon.ErrInvalidRequestSchema.Error())
		return
	}

	if err := ValidatorRequestNewPassword(request.Password, request.RepeatedPassword, p.cfg.MaxLenPassword, p.cfg.MinLenPassword); err != nil {
		api.RespondError(w, http.StatusBadRequest, handlerCommon.ErrInvalidEmailOrPassword.Error())
		return
	}

	if err := p.mailSender.CheckRecoveryCode(r.Context(), request.TokenID); err != nil {
		errLog := fmt.Errorf("mailSender.CheckRecoveryCode: %w", err)
		logger.Error().Err(errLog).Msg("mailSender.CheckRecoveryCode failed")

		if errors.Is(err, common.ErrorResetTokenNotFound) {
			api.RespondError(w, http.StatusBadRequest, common.ErrorResetTokenNotFound.Error())
			return
		}

		sentryLogger.CaptureFromContext(r.Context(), errLog, "ResetUserPassword", map[string]interface{}{
			"action": "check_recovery_code",
		})
		api.RespondError(w, http.StatusInternalServerError, handlerCommon.ErrCannotResetPassword.Error())
		return
	}

	userLink, err := p.mailSender.ExchangeTokenForUser(r.Context(), domain.ResetToken{
		Token: request.TokenID,
	})
	if err != nil {
		errLog := fmt.Errorf("mailSender.ExchangeTokenForUser: %w", err)
		logger.Error().Err(errLog).Msg("mailSender.ExchangeTokenForUser failed")

		if errors.Is(err, common.ErrorResetTokenNotFound) {
			api.RespondError(w, http.StatusNotFound, handlerCommon.ErrResetTokenNotExistOrExpired.Error())
			return
		}

		sentryLogger.CaptureFromContext(r.Context(), errLog, "ResetUserPassword", map[string]interface{}{
			"action": "exchange_token_for_user",
		})
		api.RespondError(w, http.StatusInternalServerError, handlerCommon.ErrInternalServerError.Error())
		return
	}

	if err := p.profile.ResetPassword(r.Context(), domain.UpdatedPassword{
		UserLink: userLink,
		Password: request.Password,
	}); err != nil {
		errLog := fmt.Errorf("profile.ResetPassword: %w", err)
		logger.Error().Err(errLog).Msg("profile.ResetPassword failed")

		if errors.Is(err, common.ErrorNotNullValue) {
			api.RespondError(w, http.StatusBadRequest, handlerCommon.ErrNullInNotNullField.Error())
			return
		}

		if errors.Is(err, common.ErrorNonexistentUser) {
			api.RespondError(w, http.StatusNotFound, common.ErrorNonexistentUser.Error())
			return
		}

		sentryLogger.CaptureFromContext(r.Context(), errLog, "ResetUserPassword", map[string]interface{}{
			"user_link": userLink,
			"action":    "reset_password",
		})
		api.RespondError(w, http.StatusInternalServerError, handlerCommon.ErrCannotResetPassword.Error())
		return
	}

	api.Respond(w, http.StatusOK, api.StatusOK)
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
