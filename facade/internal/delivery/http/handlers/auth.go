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
	"github.com/go-park-mail-ru/2026_1_Clac_Clac/facade/internal/domain"
	"github.com/go-park-mail-ru/2026_1_Clac_Clac/facade/internal/middleware"
	sentryLogger "github.com/go-park-mail-ru/2026_1_Clac_Clac/pkg/logger"
	"github.com/google/uuid"
	"github.com/rs/zerolog"
)

const (
	oauthCodeKey            = "code"
	oauthSuccessAuthMessage = "success"
)

type AuthUsecase interface {
	CreateSession(ctx context.Context, userLink uuid.UUID) (string, error)
	DeleteSession(ctx context.Context, sessionID string) error
	ExchangeVKCode(ctx context.Context, code string) (string, string, error)
}

type UserUsecase interface {
	GetUser(ctx context.Context, entryUser domain.Credentials) (domain.FullInfoUser, error)
	CreateUser(ctx context.Context, infoUser domain.NewCredentialsUser) (domain.FullInfoUser, error)
	ProcessUserWithVK(ctx context.Context, accessToken, email string) (uuid.UUID, error)
	GetUserLink(ctx context.Context, email string) (uuid.UUID, error)
}

type AuthConfig struct {
	MaxLenPassword    int
	MinLenPassword    int
	SessionLifetime   time.Duration
	VKOAuthRedirectTo string
}

type Auth struct {
	auth AuthUsecase
	user UserUsecase
	cfg  AuthConfig
}

func NewAuthHandler(auth AuthUsecase, user UserUsecase, cfg AuthConfig) *Auth {
	return &Auth{
		auth: auth,
		user: user,
		cfg:  cfg,
	}
}

// MeHandler проверяет текущую сессию пользователя
//
//	@Summary		Проверка авторизации (Me)
//	@Description	Возвращает 200 если сессия активна (пользователь авторизован). Используется для проверки состояния авторизации на клиенте.
//	@Tags			Auth
//	@Security		sessionCookie
//	@Produce		json
//	@Success		200	{object}	api.Response		"Пользователь авторизован"
//	@Failure		401	{object}	api.ErrorResponse	"Сессия отсутствует или истекла"
//	@Router			/me [get]
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
//	@Description	Аутентифицирует пользователя по email и паролю. При успехе устанавливает session_id cookie. Rate limit: 5 попыток в минуту.
//	@Tags			Auth
//	@Accept			json
//	@Produce		json
//	@Param			input	body		dto.LogInRequest						true	"Email и пароль"
//	@Success		200		{object}	api.OkResponse[dto.UserInfoResponse]	"Успешный вход, cookie session_id установлен"
//	@Failure		400		{object}	api.ErrorResponse						"Некорректный формат запроса, email или пароль"
//	@Failure		404		{object}	api.ErrorResponse						"Пользователь не найден или неверные учётные данные"
//	@Failure		429		{object}	api.ErrorResponse						"Слишком много попыток входа"
//	@Failure		500		{object}	api.ErrorResponse						"Внутренняя ошибка сервера"
//	@Router			/login [post]
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

	user, err := a.user.GetUser(r.Context(), domain.Credentials{
		Email:    request.Email,
		Password: request.Password,
	})
	if err != nil {
		if errors.Is(err, common.ErrorNonexistentEmail) ||
			errors.Is(err, common.ErrorNonexistentUser) ||
			errors.Is(err, common.ErrorWrongCredentials) {
			logger.Err(fmt.Errorf("user.GetUser: %w", err)).Msg("get info user")
			api.RespondError(w, http.StatusNotFound, handlerCommon.ErrWrongEmailOrPassword.Error())
			return
		}

		errLog := fmt.Errorf("user.GetUser: %w", err)
		logger.Err(errLog).Msg("get info user")

		sentryLogger.CaptureFromContext(r.Context(), errLog, "LoginUser", map[string]interface{}{
			"email":  request.Email,
			"action": "get_db_user",
		})

		api.RespondError(w, http.StatusInternalServerError, handlerCommon.ErrInternalServerError.Error())
		return
	}

	sessionID, err := a.auth.CreateSession(r.Context(), user.UserLink)
	if err != nil {
		if errors.Is(err, common.ErrorNonexistentUser) {
			errLog := fmt.Errorf("auth.CreateSession user not found just after creation: %w", err)
			logger.Err(errLog).Msg("create session")

			sentryLogger.CaptureFromContext(r.Context(), errLog, "LoginUser", map[string]interface{}{
				"user_link": user.UserLink,
				"action":    "create_session_anomaly",
			})

			api.RespondError(w, http.StatusNotFound, handlerCommon.ErrUserDoesNotExists.Error())
			return
		}

		errLog := fmt.Errorf("auth.CreateSession: %w", err)
		logger.Err(errLog).Msg("create session")

		sentryLogger.CaptureFromContext(r.Context(), errLog, "LoginUser", map[string]interface{}{
			"user_link": user.UserLink,
			"action":    "create_session_anomaly",
		})

		api.RespondError(w, http.StatusInternalServerError, handlerCommon.ErrInternalServerError.Error())
		return
	}

	http.SetCookie(w, api.NewSessionCookie(
		middleware.SessiondIdKey,
		sessionID,
		time.Now().Add(a.cfg.SessionLifetime)))

	api.HandleError(api.RespondOk(w, dto.UserInfoResponse{
		Link:        user.UserLink,
		DisplayName: user.DisplayName,
		Email:       user.Email,
		Avatar:      user.AvatarURL,
	}))
}

// RegisterUser регистрирует нового пользователя
//
//	@Summary		Регистрация
//	@Description	Создаёт новый аккаунт. Пароли должны совпадать, email должен быть уникальным. При успехе устанавливает session_id cookie. Rate limit: 5 попыток в час.
//	@Tags			Auth
//	@Accept			json
//	@Produce		json
//	@Param			input	body		dto.RegisterRequest						true	"Данные для регистрации"
//	@Success		201		{object}	api.OkResponse[dto.UserInfoResponse]	"Аккаунт создан, cookie session_id установлен"
//	@Failure		400		{object}	api.ErrorResponse						"Некорректный формат запроса, email/пароль; пароли не совпадают"
//	@Failure		404		{object}	api.ErrorResponse						"Пользователь не найден (ошибка при создании сессии)"
//	@Failure		409		{object}	api.ErrorResponse						"Пользователь с таким email уже существует"
//	@Failure		429		{object}	api.ErrorResponse						"Слишком много попыток регистрации"
//	@Failure		500		{object}	api.ErrorResponse						"Внутренняя ошибка сервера"
//	@Router			/register [post]
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

	user, err := a.user.CreateUser(r.Context(), domain.NewCredentialsUser{
		DisplayName: request.DisplayName,
		Password:    request.Password,
		Email:       request.Email,
	})
	if err != nil {
		if errors.Is(err, common.ErrorExistingUser) {
			api.RespondError(w, http.StatusConflict, handlerCommon.ErrUserAlreadyExists.Error())
			return
		}

		if errors.Is(err, common.ErrorNotNullValue) {
			api.RespondError(w, http.StatusBadRequest, handlerCommon.ErrNullInNotNullField.Error())
			return
		}

		errLog := fmt.Errorf("user.CreateUser: %w", err)
		logger.Err(errLog).Str("email", request.Email).Msg("failed to create user")

		sentryLogger.CaptureFromContext(r.Context(), errLog, "RegisterUser", map[string]interface{}{
			"email":        request.Email,
			"display_name": request.DisplayName,
			"action":       "create_db_record",
		})
		api.RespondError(w, http.StatusInternalServerError, handlerCommon.ErrInternalServerError.Error())
		return
	}

	sessionID, err := a.auth.CreateSession(r.Context(), user.UserLink)
	if err != nil {
		if errors.Is(err, common.ErrorNonexistentUser) {
			errLog := fmt.Errorf("auth.CreateSession user not found just after creation: %w", err)
			logger.Err(errLog).Msg("create session")

			sentryLogger.CaptureFromContext(r.Context(), errLog, "RegisterUser", map[string]interface{}{
				"user_link": user.UserLink,
				"action":    "create_session_anomaly",
			})

			api.RespondError(w, http.StatusNotFound, handlerCommon.ErrUserDoesNotExists.Error())
			return
		}

		errLog := fmt.Errorf("auth.CreateSession: %w", err)
		logger.Err(errLog).Msg("create session")

		sentryLogger.CaptureFromContext(r.Context(), errLog, "RegisterUser", map[string]interface{}{
			"user_link": user.UserLink,
			"action":    "create_session_redis",
		})

		api.RespondError(w, http.StatusInternalServerError, handlerCommon.ErrInternalServerError.Error())
		return
	}

	http.SetCookie(w, api.NewSessionCookie(
		middleware.SessiondIdKey,
		sessionID,
		time.Now().Add(a.cfg.SessionLifetime)))

	api.HandleError(api.RespondCreated(w, dto.UserInfoResponse{
		Link:        user.UserLink,
		DisplayName: user.DisplayName,
		Email:       user.Email,
		Avatar:      user.AvatarURL,
	}))
}

// LogOutUser удаляет сессию пользователя
//
//	@Summary		Выход (Logout)
//	@Description	Инвалидирует текущую сессию и очищает cookie session_id. Если cookie отсутствует, всё равно возвращает 200 — endpoint идемпотентен.
//	@Tags			Auth
//	@Produce		json
//	@Success		200	{object}	api.Response	"Сессия завершена"
//	@Router			/logout [post]
func (a *Auth) LogOutUser(w http.ResponseWriter, r *http.Request) {
	logger := zerolog.Ctx(r.Context())

	cookie, err := r.Cookie(middleware.SessiondIdKey)
	if err == nil && cookie != nil {
		if errDeleteSession := a.auth.DeleteSession(r.Context(), cookie.Value); errDeleteSession != nil {
			errLog := fmt.Errorf("usecase.Logout: %w", errDeleteSession)
			logger.Err(errLog).Msg("logout user")

			sentryLogger.CaptureFromContext(r.Context(), errLog, "LogOutUser Context", map[string]interface{}{
				"action": "delete_session_redis",
			})
		}
	}

	http.SetCookie(w, api.NewExpiredCookie(middleware.SessiondIdKey))
	api.Respond(w, http.StatusOK, api.StatusOK)
}

// VkOAuthCallback обрабатывает коллбэк от VK
//
//	@Summary		VK OAuth Коллбэк
//	@Description	Обменивает временный code от VK на access_token, находит или создаёт пользователя, создаёт сессию. Всегда отвечает HTTP 302 редиректом на фронтенд — при успехе и при ошибке.
//	@Tags			Auth
//	@Param			code	query	string	true	"Временный OAuth-код от VK"
//	@Success		302		"Редирект: ?code=200&message=success (cookie session_id установлен)"
//	@Failure		302		"Редирект с ошибкой: ?code=400 (code пустой), ?code=502 (VK недоступен), ?code=500 (ошибка обработки или создания сессии)"
//	@Router			/oauth/vk [get]
func (a *Auth) VkOAuthCallback(w http.ResponseWriter, r *http.Request) {
	logger := zerolog.Ctx(r.Context())

	code := r.FormValue(oauthCodeKey)
	if code == "" {
		logger.Err(handlerCommon.ErrOAuthCodeEmpty).Msg("vk oauth callback: missing code")
		Redirect(w, r, a.cfg.VKOAuthRedirectTo, http.StatusBadRequest, handlerCommon.ErrOAuthCodeEmpty.Error())
		return
	}

	accessToken, email, err := a.auth.ExchangeVKCode(r.Context(), code)
	if err != nil {
		errLog := fmt.Errorf("auth.ExchangeVKCode: %w", err)
		logger.Err(errLog).Msg("vk oauth callback: exchange code failed")

		outErr := handlerCommon.ErrOAuthExchangeFailed
		if errors.Is(err, common.ErrorVKOAuthUnavailable) {
			outErr = handlerCommon.ErrOAuthUnavailable
		}

		sentryLogger.CaptureFromContext(r.Context(), errLog, "VkOAuth Context", map[string]interface{}{
			"action":            "exchange_code",
			"is_vk_unavailable": errors.Is(err, common.ErrorVKOAuthUnavailable),
		})

		Redirect(w, r, a.cfg.VKOAuthRedirectTo, http.StatusBadGateway, outErr.Error())
		return
	}

	userLink, err := a.user.ProcessUserWithVK(r.Context(), accessToken, email)
	if err != nil {
		errLog := fmt.Errorf("user.ProcessUserWithVK: %w", err)
		logger.Err(errLog).Msg("vk oauth callback: user processing failed")

		outErr := handlerCommon.ErrOAuthInternalServerError
		if errors.Is(err, common.ErrorVKOAuthUnavailable) {
			outErr = handlerCommon.ErrOAuthCannotRequestUserData
		}

		sentryLogger.CaptureFromContext(r.Context(), errLog, "VkOAuth Context", map[string]interface{}{
			"email":  email,
			"action": "process_vk_user",
		})

		Redirect(w, r, a.cfg.VKOAuthRedirectTo, http.StatusInternalServerError, outErr.Error())
		return
	}

	sessionID, err := a.auth.CreateSession(r.Context(), userLink)
	if err != nil {
		errLog := fmt.Errorf("auth.CreateSession: %w", err)
		logger.Err(errLog).Msg("create session in vk oauth")

		outErr := handlerCommon.ErrInternalServerError
		if errors.Is(err, common.ErrorNonexistentUser) {
			outErr = handlerCommon.ErrUserDoesNotExists

			sentryLogger.CaptureFromContext(r.Context(), errLog, "VkOAuth Context", map[string]interface{}{
				"user_link": userLink,
				"action":    "create_session_anomaly",
			})
		} else {
			sentryLogger.CaptureFromContext(r.Context(), errLog, "VkOAuth Context", map[string]interface{}{
				"user_link": userLink,
				"action":    "create_session_redis",
			})
		}

		Redirect(w, r, a.cfg.VKOAuthRedirectTo, http.StatusInternalServerError, outErr.Error())
		return
	}

	http.SetCookie(w, api.NewSessionCookie(
		middleware.SessiondIdKey,
		sessionID,
		time.Now().Add(a.cfg.SessionLifetime)))

	Redirect(w, r, a.cfg.VKOAuthRedirectTo, http.StatusOK, oauthSuccessAuthMessage)
}
