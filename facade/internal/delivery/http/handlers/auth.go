package handlers

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/go-park-mail-ru/2026_1_Clac_Clac/facade/internal/api"
	"github.com/go-park-mail-ru/2026_1_Clac_Clac/facade/internal/common"
	"github.com/go-park-mail-ru/2026_1_Clac_Clac/facade/internal/delivery/http/dto"
	handlerCommon "github.com/go-park-mail-ru/2026_1_Clac_Clac/facade/internal/delivery/http/handlers/common"
	"github.com/go-park-mail-ru/2026_1_Clac_Clac/facade/internal/delivery/http/middleware"
	"github.com/go-park-mail-ru/2026_1_Clac_Clac/facade/internal/domain"
	"github.com/google/uuid"
	"github.com/rs/zerolog"
)

const (
	oauthCodeKey            = "code"
	oauthSuccessAuthMessage = "success"
	csrfCookieKey           = "csrf_token"
	nameCoolDown            = "recovery_email"
	coolDownExpirationSec   = 60
)

type AuthUseCase interface {
	Login(ctx context.Context, cred domain.Credentials) (domain.UserInfo, string, error)
	Register(ctx context.Context, cred domain.NewCredentialsUser) (domain.UserInfo, string, error)
	LoginWithVK(ctx context.Context, code string) (domain.UserInfo, string, error)
	Logout(ctx context.Context, sessionID string) error
	SendRecoveryCode(ctx context.Context, email string) error
	CheckRecoveryCode(ctx context.Context, code string) error
	ResetPassword(ctx context.Context, tokenID, newPassword string) error
}

type CoolDownUseCase interface {
	CheckCoolDown(ctx context.Context, cooldown domain.Cooldown) (domain.CooldownResult, error)
}

type CSRFUseCase interface {
	GetExpireTime(ctx context.Context) time.Time
	Generate(ctx context.Context, sessionID string, expireAt int64) (string, error)
	Check(ctx context.Context, sessionID string, token string) error
}

type AuthConfig struct {
	MaxLenPassword    int
	MinLenPassword    int
	SessionLifetime   time.Duration
	VKOAuthRedirectTo string
}
type Auth struct {
	usecase  AuthUseCase
	cooldown CoolDownUseCase
	csrf     CSRFUseCase
	cfg      AuthConfig
}

func NewAuthHandler(usecase AuthUseCase, cooldown CoolDownUseCase, csrf CSRFUseCase, cfg AuthConfig) *Auth {
	return &Auth{
		usecase:  usecase,
		cooldown: cooldown,
		csrf:     csrf,
		cfg:      cfg,
	}
}

// MeHandler проверяет текущую сессию пользователя
//
//	@Summary		Проверка авторизации (Me)
//	@Tags			Auth
//	@Security		sessionCookie
//	@Security		csrfToken
//	@Produce		json
//	@Success		200	{string}	string				"OK"
//	@Failure		401	{object}	api.ErrorResponse	"User not authorized"
//	@Router			/api/me [get]
func (a *Auth) MeHandler(w http.ResponseWriter, r *http.Request) {
	value := r.Context().Value(middleware.UserContextLink{})
	_, ok := value.(uuid.UUID)
	if !ok {
		api.RespondError(w, http.StatusUnauthorized, handlerCommon.ErrUserNotAuthorized.Error())
		return
	}

	api.HandleError(api.Respond(w, http.StatusOK, api.StatusOK))
}

// LogInUser выполняет вход пользователя
//
//	@Summary		Вход (Login)
//	@Tags			Auth
//	@Accept			json
//	@Produce		json
//	@Param			input	body		dto.LogInRequest	true	"Данные для входа"
//	@Success		200		{object}	dto.UserInfoResponse
//	@Failure		400		{object}	api.ErrorResponse	"Invalid email or password"
//	@Failure		401		{object}	api.ErrorResponse	"Wrong credentials"
//	@Failure		429		{object}	api.ErrorResponse	"Too many requests"
//	@Failure		500		{object}	api.ErrorResponse	"Internal server error"
//	@Router			/api/login [post]
func (a *Auth) LogInUser(w http.ResponseWriter, r *http.Request) {
	logger := zerolog.Ctx(r.Context())

	var request dto.LogInRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		api.RespondError(w, http.StatusBadRequest, handlerCommon.ErrInvalidRequestSchema.Error())
		return
	}

	if err := ValidatorRequestAuth(request.Email, request.Password, a.cfg.MaxLenPassword, a.cfg.MinLenPassword); err != nil {
		api.RespondError(w, http.StatusBadRequest, handlerCommon.ErrInvalidEmailOrPassword.Error())
		return
	}

	serviceUser, sessionID, err := a.usecase.Login(r.Context(), domain.Credentials{
		Email:    request.Email,
		Password: request.Password,
	})
	if err != nil {
		if errors.Is(err, common.ErrorWrongCredentials) {
			api.RespondError(w, http.StatusUnauthorized, handlerCommon.ErrWrongEmailOrPassword.Error())
			return
		}
		logger.Err(fmt.Errorf("auth.Login: %w", err)).Msg("login user")
		api.RespondError(w, http.StatusInternalServerError, handlerCommon.ErrInternalServerError.Error())
		return
	}

	handlerUser := dto.UserInfoResponse{
		Link:        serviceUser.Link,
		DisplayName: serviceUser.DisplayName,
		Email:       serviceUser.Email,
		Avatar:      serviceUser.Avatar,
	}

	http.SetCookie(w, api.NewSessionCookie(
		middleware.SessiondIdKey,
		sessionID,
		time.Now().Add(a.cfg.SessionLifetime)))

	api.HandleError(api.RespondOk(w, handlerUser))
}

// RegisterUser регистрирует нового пользователя
//
//	@Summary		Регистрация
//	@Tags			Auth
//	@Accept			json
//	@Produce		json
//	@Param			input	body		dto.RegisterRequest	true	"Данные для регистрации"
//	@Success		201		{object}	dto.UserInfoResponse
//	@Failure		400		{object}	api.ErrorResponse	"Invalid email or password"
//	@Failure		409		{object}	api.ErrorResponse	"User already exists"
//	@Failure		429		{object}	api.ErrorResponse	"Too many requests"
//	@Failure		500		{object}	api.ErrorResponse	"Internal server error"
//	@Router			/api/register [post]
func (a *Auth) RegisterUser(w http.ResponseWriter, r *http.Request) {
	logger := zerolog.Ctx(r.Context())

	var request dto.RegisterRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		api.RespondError(w, http.StatusBadRequest, handlerCommon.ErrInvalidRequestSchema.Error())
		return
	}

	request.Sanitize()

	if err := ValidatorWithCheckPassword(request.Email, request.Password, request.RepeatedPassword, a.cfg.MaxLenPassword, a.cfg.MinLenPassword); err != nil {
		api.RespondError(w, http.StatusBadRequest, handlerCommon.ErrInvalidEmailOrPassword.Error())
		return
	}

	serviceUser, sessionID, err := a.usecase.Register(r.Context(), domain.NewCredentialsUser{
		DisplayName: request.DisplayName,
		Email:       request.Email,
		Password:    request.Password,
	})
	if err != nil {
		logger.Err(fmt.Errorf("auth.Register: %w", err)).Msg("register user")
		if errors.Is(err, common.ErrorExistingUser) {
			api.RespondError(w, http.StatusConflict, common.ErrorExistingUser.Error())
			return
		}
		if errors.Is(err, common.ErrorNotNullValue) {
			api.RespondError(w, http.StatusBadRequest, common.ErrorNotNullValue.Error())
			return
		}
		api.RespondError(w, http.StatusInternalServerError, handlerCommon.ErrInternalServerError.Error())
		return
	}

	handlerUser := dto.UserInfoResponse{
		Link:        serviceUser.Link,
		DisplayName: serviceUser.DisplayName,
		Email:       serviceUser.Email,
		Avatar:      serviceUser.Avatar,
	}

	http.SetCookie(w, api.NewSessionCookie(
		middleware.SessiondIdKey,
		sessionID,
		time.Now().Add(a.cfg.SessionLifetime)))

	api.HandleError(api.RespondCreated(w, handlerUser))
}

// LogOutUser удаляет сессию пользователя
//
//	@Summary		Выход (Logout)
//	@Tags			Auth
//	@Produce		json
//	@Success		200	{string}	string	"OK"
//	@Router			/api/logout [post]
func (a *Auth) LogOutUser(w http.ResponseWriter, r *http.Request) {
	logger := zerolog.Ctx(r.Context())

	cookie, err := r.Cookie(middleware.SessiondIdKey)
	if err == nil && cookie != nil {
		if errLogOut := a.usecase.Logout(r.Context(), cookie.Value); errLogOut != nil {
			logger.Err(fmt.Errorf("usecase.Logout: %w", errLogOut)).Msg("logout user")
		}
	}

	http.SetCookie(w, api.NewExpiredCookie(middleware.SessiondIdKey))
	api.Respond(w, http.StatusOK, api.StatusOK)
}

// SendRecoveryEmail отправляет код восстановления пароля
//
//	@Summary		Отправить код восстановления
//	@Tags			Auth
//	@Accept			json
//	@Produce		json
//	@Param			input	body		dto.PasswordRecoveryRequest	true	"Email пользователя"
//	@Success		200		{string}	string						"OK"
//	@Failure		400		{object}	api.ErrorResponse			"Invalid email"
//	@Failure		404		{object}	api.ErrorResponse			"User does not exist"
//	@Failure		429		{object}	api.ErrorResponse			"Too many requests"
//	@Failure		500		{object}	api.ErrorResponse			"Cannot send recovery code"
//	@Router			/api/forgot-password [post]
func (a *Auth) SendRecoveryEmail(w http.ResponseWriter, r *http.Request) {
	logger := zerolog.Ctx(r.Context())

	var request dto.PasswordRecoveryRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		api.RespondError(w, http.StatusBadRequest, handlerCommon.ErrInvalidRequestSchema.Error())
		return
	}

	ok := ValidateEmail(request.Email)
	if !ok {
		api.RespondError(w, http.StatusBadRequest, handlerCommon.ErrInvalidEmailOrPassword.Error())
		return
	}

	result, err := a.cooldown.CheckCoolDown(r.Context(), domain.Cooldown{
		Name:         nameCoolDown,
		Email:        request.Email,
		ExpirationMs: coolDownExpirationSec,
	})
	if err != nil {
		api.RespondError(w, http.StatusInternalServerError, handlerCommon.ErrInternalServerError.Error())
		return
	}

	if !result.Allowed {
		w.Header().Set("Retry-After", fmt.Sprintf("%d", result.WaitS))
		api.RespondError(w, http.StatusTooManyRequests,
			fmt.Sprintf("Too many requests. Wait %d seconds", result.WaitS))
		return
	}

	if err := a.usecase.SendRecoveryCode(r.Context(), request.Email); err != nil {
		if errors.Is(err, common.ErrorNonexistentEmail) || errors.Is(err, common.ErrorNonexistentUser) {
			api.RespondError(w, http.StatusNotFound, handlerCommon.ErrUserDoesNotExists.Error())
			return
		}
		logger.Err(fmt.Errorf("auth.SendRecoveryCode: %w", err)).Msg("send recovery code")
		api.RespondError(w, http.StatusInternalServerError, handlerCommon.ErrCannotSendRecoveryCode.Error())
		return
	}

	api.Respond(w, http.StatusOK, api.StatusOK)
}

// CheckRecoveryCode проверяет отправленный на почту код
//
//	@Summary		Проверить код восстановления
//	@Tags			Auth
//	@Accept			json
//	@Produce		json
//	@Param			input	body		dto.RecoveryCodeRequest	true	"Код из письма"
//	@Success		200		{string}	string					"OK"
//	@Failure		400		{object}	api.ErrorResponse		"Invalid request schema"
//	@Failure		500		{object}	api.ErrorResponse		"Internal server error"
//	@Router			/api/check-code [post]
func (a *Auth) CheckRecoveryCode(w http.ResponseWriter, r *http.Request) {
	logger := zerolog.Ctx(r.Context())

	var request dto.RecoveryCodeRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		api.RespondError(w, http.StatusBadRequest, handlerCommon.ErrInvalidRequestSchema.Error())
		return
	}

	if err := a.usecase.CheckRecoveryCode(r.Context(), request.Code); err != nil {
		logger.Error().Err(err).Msg("auth.CheckRecoveryCode failed")
		api.RespondError(w, http.StatusInternalServerError, handlerCommon.ErrInternalServerError.Error())
		return
	}

	api.Respond(w, http.StatusOK, api.StatusOK)
}

// ResetUserPassword устанавливает новый пароль
//
//	@Summary		Сброс пароля
//	@Tags			Auth
//	@Accept			json
//	@Produce		json
//	@Param			input	body		dto.NewPasswordRequest	true	"Новый пароль и токен"
//	@Success		200		{string}	string					"OK"
//	@Failure		400		{object}	api.ErrorResponse		"Invalid password or token not found"
//	@Failure		404		{object}	api.ErrorResponse		"User not found"
//	@Failure		500		{object}	api.ErrorResponse		"Cannot reset password"
//	@Router			/api/reset-password [post]
func (a *Auth) ResetUserPassword(w http.ResponseWriter, r *http.Request) {
	logger := zerolog.Ctx(r.Context())

	var request dto.NewPasswordRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		api.RespondError(w, http.StatusBadRequest, handlerCommon.ErrInvalidRequestSchema.Error())
		return
	}

	if err := ValidatorRequestNewPassword(request.Password, request.RepeatedPassword, a.cfg.MaxLenPassword, a.cfg.MinLenPassword); err != nil {
		api.RespondError(w, http.StatusBadRequest, handlerCommon.ErrInvalidEmailOrPassword.Error())
		return
	}

	if err := a.usecase.ResetPassword(r.Context(), request.TokenID, request.Password); err != nil {
		logger.Error().Err(err).Msg("auth.ResetPassword failed")

		if errors.Is(err, common.ErrorNotNullValue) {
			api.RespondError(w, http.StatusBadRequest, common.ErrorNotNullValue.Error())
			return
		}
		if errors.Is(err, common.ErrorResetTokenNotFound) {
			api.RespondError(w, http.StatusBadRequest, common.ErrorResetTokenNotFound.Error())
			return
		}
		if errors.Is(err, common.ErrorNonexistentUser) {
			api.RespondError(w, http.StatusNotFound, common.ErrorNonexistentUser.Error())
			return
		}

		api.RespondError(w, http.StatusInternalServerError, handlerCommon.ErrCannotResetPassword.Error())
		return
	}

	api.Respond(w, http.StatusOK, api.StatusOK)
}

// VkOAuthCallback обрабатывает коллбэк от VK
//
//	@Summary		VK OAuth Коллбэк
//	@Tags			Auth
//	@Param			code	query	string	true	"Временный код от VK"
//	@Success		302		"Редирект на фронтенд с успешной авторизацией"
//	@Failure		400		"Редирект с ошибкой: code пустой"
//	@Failure		502		"Редирект с ошибкой: VK недоступен"
//	@Router			/api/oauth/vk [get]
func (a *Auth) VkOAuthCallback(w http.ResponseWriter, r *http.Request) {
	logger := zerolog.Ctx(r.Context())

	code := r.FormValue(oauthCodeKey)
	if code == "" {
		logger.Err(handlerCommon.ErrOAuthCodeEmpty).Msg("vk oauth callback: missing code")
		Redirect(w, r, a.cfg.VKOAuthRedirectTo, http.StatusBadRequest, handlerCommon.ErrOAuthCodeEmpty.Error())
		return
	}

	userInfo, sessionID, err := a.usecase.LoginWithVK(r.Context(), code)
	if err != nil {
		logger.Err(err).Msg("usecase.LoginWithVK failed")
		statusCode := http.StatusInternalServerError
		if errors.Is(err, common.ErrorVKOAuthUnavailable) {
			statusCode = http.StatusBadGateway
		}
		Redirect(w, r, a.cfg.VKOAuthRedirectTo, statusCode, handlerCommon.ErrOAuthInternalServerError.Error())
		return
	}

	http.SetCookie(w, api.NewSessionCookie(
		middleware.SessiondIdKey,
		sessionID,
		time.Now().Add(a.cfg.SessionLifetime)))

	_ = userInfo
	Redirect(w, r, a.cfg.VKOAuthRedirectTo, http.StatusOK, oauthSuccessAuthMessage)
}

// SetCSRFCookieHandler генерирует и устанавливает CSRF токен
//
//	@Summary		Получить CSRF токен
//	@Tags			Auth
//	@Security		sessionCookie
//	@Produce		json
//	@Success		200	{string}	string				"OK"
//	@Failure		401	{object}	api.ErrorResponse	"User not authorized"
//	@Failure		500	{object}	api.ErrorResponse	"Cannot create CSRF token"
//	@Router			/api/csrf [get]
func (a *Auth) SetCSRFCookieHandler(w http.ResponseWriter, r *http.Request) {
	logger := zerolog.Ctx(r.Context())

	cookie, err := r.Cookie(middleware.SessiondIdKey)
	if err != nil {
		api.RespondError(w, http.StatusUnauthorized, handlerCommon.ErrUserNotAuthorized.Error())
		return
	}
	sessionID := cookie.Value

	expireTime := a.csrf.GetExpireTime(r.Context())

	token, err := a.csrf.Generate(r.Context(), sessionID, expireTime.Unix())
	if err != nil {
		logger.Error().Err(err).Msg("generate csrf token")
		api.RespondError(w, http.StatusInternalServerError, handlerCommon.ErrCannotCreateCSRFToken.Error())
		return
	}

	http.SetCookie(w, api.NewCSRFCookie(csrfCookieKey, token, expireTime))
	api.Respond(w, http.StatusOK, api.StatusOK)
}
