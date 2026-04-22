package handler

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/go-park-mail-ru/2026_1_Clac_Clac/monolith/internal/api"
	"github.com/go-park-mail-ru/2026_1_Clac_Clac/monolith/internal/auth/handler/dto"
	"github.com/go-park-mail-ru/2026_1_Clac_Clac/monolith/internal/auth/service"
	serviceDto "github.com/go-park-mail-ru/2026_1_Clac_Clac/monolith/internal/auth/service/dto"
	"github.com/go-park-mail-ru/2026_1_Clac_Clac/monolith/internal/common"
	"github.com/go-park-mail-ru/2026_1_Clac_Clac/monolith/internal/config"
	"github.com/go-park-mail-ru/2026_1_Clac_Clac/monolith/internal/middleware"
	"github.com/google/uuid"
	"github.com/rs/zerolog"
	"golang.org/x/oauth2"
)

//go:generate mockery --name=AuthService --output=mock_auth_srv --outpkg=mockAuthSrv
type AuthService interface {
	Register(ctx context.Context, requestUser serviceDto.RegistrationUser) (serviceDto.UserInfo, string, error)
	LogIn(ctx context.Context, requestUser serviceDto.LogInUser) (serviceDto.UserInfo, string, error)
	CreateSessionForUser(ctx context.Context, link uuid.UUID) (string, error)
	RefreshSession(ctx context.Context, sessionID string) error
	UpdateCountRequests(ctx context.Context, config serviceDto.RateLimiterConfig) (bool, error)
	CheckCoolDown(ctx context.Context, config serviceDto.CoolDownConfig) (bool, time.Duration, error)
	LogOut(ctx context.Context, sessionID string) error
	GetUserLink(ctx context.Context, sessionID string) (uuid.UUID, error)
	GetUserByEmail(ctx context.Context, email string) (serviceDto.UserInfo, error)
	SendRecoveryCode(ctx context.Context, email string) error
	CheckRecoveryCode(ctx context.Context, tokenID string) error
	ResetPassword(ctx context.Context, tokenID, newPassword string) error
	EnsureUserByEmail(ctx context.Context, info serviceDto.RegistrationUser) (serviceDto.UserInfo, error)
	SaveRefreshTokenFroUser(ctx context.Context, info serviceDto.UserInfo, token string) error
	GetCSRFTokenExpireTime(ctx context.Context) (time.Time, error)
	GenerateCSRFToken(ctx context.Context, sessionId string, expireTime int64) (string, error)
	CheckCSRFToken(ctx context.Context, sessionId string, token string) error
}

//go:generate mockery --name=VkOAuth --output mock_vk_oauth
type VkOAuth interface {
	Exchange(ctx context.Context, code string, opts ...oauth2.AuthCodeOption) (*oauth2.Token, error)
	Client(ctx context.Context, t *oauth2.Token) *http.Client
}

type Config struct {
	MaxLenPassword  int
	MinLenPassword  int
	SessionLifetime time.Duration
}

type Handler struct {
	srv AuthService
	cfg Config
}

func NewHandler(srv AuthService, cfg Config) *Handler {
	return &Handler{
		srv: srv,
		cfg: cfg,
	}
}

const (
	oauthCodeKey            = "code"
	oauthEmailKey           = "email"
	oauthSuccessAuthMessage = "success"

	csrfCookieKey = "csrf_token"

	nameCoolDown = "recovery_email"
)

var (
	ErrInvalidRequestSchema   = errors.New("invalid schema")
	ErrInvalidEmailOrPassword = errors.New("invalid email or password")
	ErrWrongEmailOrPassword   = errors.New("wrong email or password")
	ErrCannotSendRecoveryCode = errors.New("cannot send recovery code")
	ErrCannotResetPassword    = errors.New("cannot reset password")
	ErrInternalServerError    = errors.New("something went wrong")
	ErrUserNotAuthorized      = errors.New("user not authorized")
	ErrUserDoesNotExists      = errors.New("user does not exists")

	ErrOAuthCodeEmpty              = errors.New("oauth_code_empty")
	ErrOAuthExchangeFailed         = errors.New("oauth_error")
	ErrOAuthNoEmailProvided        = errors.New("oauth_no_email")
	ErrOAuthInvalidEmail           = errors.New("oauth_invalid_email")
	ErrOAuthCannotRequestUserData  = errors.New("oauth_cannot_request_user_data")
	ErrOAuthEmptyUserData          = errors.New("oauth_no_user_data")
	ErrOAuthInternalServerError    = errors.New("oauth_something_went_wrong")
	ErrOAuthCannotSaveRefreshToken = errors.New("oauth cannot save refresh token")

	ErrCannotCreateCSRFToken        = errors.New("cannot create csrf token")
	ErrCannotGetCSRFTokenExpireTime = errors.New("cannot get csrf token expire time")
)

// @Summary      Проверка авторизации
// @Description  Проверяет валидность текущей сессии пользователя (извлекается через middleware).
// @Tags         auth
// @Produce      json
// @Success      200 {object} api.Response "Успешная авторизация (ok)"
// @Failure      401 {object} api.ErrorResponse "user not authorized"
// @Security     CookieAuth
// @Router       /me [get]
func (h *Handler) MeHandler(w http.ResponseWriter, r *http.Request) {
	value := r.Context().Value(middleware.UserContextLink{})
	_, ok := value.(uuid.UUID)
	if !ok {
		api.RespondError(w, http.StatusUnauthorized, ErrUserNotAuthorized.Error())
		return
	}

	api.HandleError(api.Respond(w, http.StatusOK, api.StatusOK))
}

// @Summary      Вход в систему
// @Description  Аутентификация пользователя по email и паролю. Устанавливает HTTP-only cookie с сессией.
// @Tags         auth
// @Accept       json
// @Produce      json
// @Param        request body     dto.LogInRequest true "Учетные данные пользователя"
// @Success      200     {object} api.OkResponse[dto.UserInfoResponse] "Успешная аутентификация"
// @Failure      400     {object} api.ErrorResponse "invalid schema / invalid email or password"
// @Failure      401     {object} api.ErrorResponse "wrong email or password"
// @Failure      500     {object} api.ErrorResponse "internal server error"
// @Router       /login [post]
func (h *Handler) LogInUser(w http.ResponseWriter, r *http.Request) {
	logger := zerolog.Ctx(r.Context())

	var request dto.LogInRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		api.RespondError(w, http.StatusBadRequest, ErrInvalidRequestSchema.Error())
		return
	}

	err := ValidatorRequestAuth(request.Email, request.Password, h.cfg.MaxLenPassword, h.cfg.MinLenPassword)
	if err != nil {
		api.RespondError(w, http.StatusBadRequest, ErrInvalidEmailOrPassword.Error())
		return
	}

	serviceUser, sessionID, err := h.srv.LogIn(r.Context(), serviceDto.LogInUser{
		Email:    request.Email,
		Password: request.Password,
	})
	if err != nil {
		if errors.Is(err, service.ErrorWrongPassword) {
			api.RespondError(w, http.StatusUnauthorized, ErrWrongEmailOrPassword.Error())
			return
		}

		logger.Err(fmt.Errorf("auth.Login: %w", err)).Msg("auth handler log in user")
		api.RespondError(w, http.StatusInternalServerError, ErrInternalServerError.Error())
		return
	}

	handlerUser := dto.UserInfoResponse{
		Link:        serviceUser.Link,
		DisplayName: serviceUser.DisplayName,
		Email:       serviceUser.Email,
		Avatar:      serviceUser.Avatar,
	}

	http.SetCookie(w, api.NewSessionCookie(
		service.SessiondIdKey,
		sessionID,
		time.Now().Add(h.cfg.SessionLifetime)))

	api.HandleError(api.RespondOk(w, handlerUser))
}

// @Summary      Регистрация нового пользователя
// @Description  Создает новый аккаунт и сразу авторизует пользователя, выдавая сессионную cookie.
// @Tags         auth
// @Accept       json
// @Produce      json
// @Param        request body     dto.RegisterRequest true "Данные для регистрации"
// @Success      201     {object} api.OkResponse[dto.UserInfoResponse] "Пользователь успешно создан"
// @Failure      400     {object} api.ErrorResponse "invalid schema / invalid email or password / user exists / empty fields"
// @Failure      500     {object} api.ErrorResponse "internal server error"
// @Router       /register [post]
func (h *Handler) RegisterUser(w http.ResponseWriter, r *http.Request) {
	logger := zerolog.Ctx(r.Context())

	var request dto.RegisterRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		api.RespondError(w, http.StatusBadRequest, ErrInvalidRequestSchema.Error())
		return
	}

	request.Sanitize()

	err := ValidatorWithCheckPassword(request.Email, request.Password, request.RepeatedPassword, h.cfg.MaxLenPassword, h.cfg.MinLenPassword)
	if err != nil {
		api.RespondError(w, http.StatusBadRequest, ErrInvalidEmailOrPassword.Error())
		return
	}

	serviceUser, sessionID, err := h.srv.Register(r.Context(), serviceDto.RegistrationUser{
		DisplayName: request.DisplayName,
		Email:       request.Email,
		Password:    request.Password,
	})
	if err != nil {
		logger.Err(fmt.Errorf("auth.Register: %w", err))
		if errors.Is(err, common.ErrorExistingUser) {
			api.RespondError(w, http.StatusBadRequest, common.ErrorExistingUser.Error())
			return
		}

		if errors.Is(err, common.ErrorNotNullValue) {
			api.RespondError(w, http.StatusBadRequest, common.ErrorNotNullValue.Error())
			return
		}

		api.RespondError(w, http.StatusInternalServerError, ErrInternalServerError.Error())
		return
	}

	handlerUser := dto.UserInfoResponse{
		Link:        serviceUser.Link,
		DisplayName: serviceUser.DisplayName,
		Email:       serviceUser.Email,
		Avatar:      serviceUser.Avatar,
	}

	http.SetCookie(w, api.NewSessionCookie(
		service.SessiondIdKey,
		sessionID,
		time.Now().Add(h.cfg.SessionLifetime)))

	api.HandleError(api.RespondCreated(w, handlerUser))
}

// @Summary      Выход из системы
// @Description  Удаляет сессию пользователя из хранилища и очищает cookie.
// @Tags         auth
// @Produce      json
// @Success      200 {object} api.Response "Успешный выход (ok)"
// @Router       /logout [post]
func (h *Handler) LogOutUser(w http.ResponseWriter, r *http.Request) {
	logger := zerolog.Ctx(r.Context())

	cookie, err := r.Cookie(service.SessiondIdKey)
	if err == nil && cookie != nil {
		errLogOut := h.srv.LogOut(r.Context(), cookie.Value)
		if errLogOut != nil {
			logger.Err(fmt.Errorf("srv.LogOut: %w", errLogOut))
		}
	}

	http.SetCookie(w, api.NewExpiredCookie(service.SessiondIdKey))
	api.Respond(w, http.StatusOK, api.StatusOK)
}

// @Summary      Запрос восстановления пароля
// @Description  Генерирует код восстановления и отправляет его на указанный email. Поддерживает rate-limiting.
// @Tags         auth
// @Accept       json
// @Produce      json
// @Param        request body     dto.PasswordRecoveryRequest true "Email пользователя"
// @Success      200     {object} api.Response "Код успешно отправлен"
// @Failure      400     {object} api.ErrorResponse "invalid schema"
// @Failure      404     {object} api.ErrorResponse "user does not exists"
// @Failure      429     {object} api.ErrorResponse "Too many requests. Wait X seconds"
// @Failure      500     {object} api.ErrorResponse "cannot send recovery code / internal server error"
// @Header       429     {string} Retry-After "Время до следующей попытки в секундах"
// @Router       /forgot-password [post]
func (h *Handler) SendRecoveryEmail(w http.ResponseWriter, r *http.Request) {
	logger := zerolog.Ctx(r.Context())

	var request dto.PasswordRecoveryRequest
	err := json.NewDecoder(r.Body).Decode(&request)
	if err != nil {
		api.RespondError(w, http.StatusBadRequest, ErrInvalidRequestSchema.Error())
		return
	}

	isAllowed, waitTime, err := h.srv.CheckCoolDown(r.Context(), serviceDto.CoolDownConfig{
		Name:       nameCoolDown,
		Email:      request.Email,
		Expiration: 1 * time.Minute,
	})

	if err != nil {
		api.RespondError(w, http.StatusInternalServerError, "internal server error")
		return
	}

	if !isAllowed {
		w.Header().Set("Retry-After", fmt.Sprintf("%.0f", waitTime.Seconds()))
		errMsg := fmt.Sprintf("Too many requests. Wait %d seconds", int(waitTime.Seconds()))

		api.RespondError(w, http.StatusTooManyRequests, errMsg)
		return
	}

	err = h.srv.SendRecoveryCode(r.Context(), request.Email)
	if err != nil {
		if errors.Is(err, common.ErrorNonexistentUser) {
			api.RespondError(w, http.StatusNotFound, ErrUserDoesNotExists.Error())
			return
		}

		logger.Err(fmt.Errorf("auth.SendRecoveryCode: %w", err))
		api.RespondError(w, http.StatusInternalServerError, ErrCannotSendRecoveryCode.Error())
		return
	}

	api.Respond(w, http.StatusOK, api.StatusOK)
}

// @Summary      Проверка кода восстановления
// @Description  Проверяет корректность 6-значного кода, отправленного на почту.
// @Tags         auth
// @Accept       json
// @Produce      json
// @Param        request body     dto.RecoveryCodeRequest true "Код из письма"
// @Success      200     {object} api.Response "Код верен"
// @Failure      400     {object} api.ErrorResponse "invalid schema"
// @Failure      500     {object} api.ErrorResponse "internal server error"
// @Router       /check-code [post]
func (h *Handler) CheckRecoveryCode(w http.ResponseWriter, r *http.Request) {
	logger := zerolog.Ctx(r.Context())

	var request dto.RecoveryCodeRequest
	err := json.NewDecoder(r.Body).Decode(&request)
	if err != nil {
		api.RespondError(w, http.StatusBadRequest, ErrInvalidRequestSchema.Error())
		return
	}

	err = h.srv.CheckRecoveryCode(r.Context(), request.Code)
	if err != nil {
		logger.Error().Err(err).Msg("auth.CheckRecoveryCode failed")
		api.RespondError(w, http.StatusInternalServerError, ErrInternalServerError.Error())
		return
	}

	api.Respond(w, http.StatusOK, api.StatusOK)
}

// @Summary      Сброс пароля
// @Description  Устанавливает новый пароль пользователя с помощью проверенного токена.
// @Tags         auth
// @Accept       json
// @Produce      json
// @Param        request body     dto.NewPasswordRequest true "Новый пароль и токен"
// @Success      200     {object} api.Response "Пароль успешно изменен"
// @Failure      400     {object} api.ErrorResponse "invalid schema / invalid email or password / missing field / token expired"
// @Failure      404     {object} api.ErrorResponse "user not found"
// @Failure      500     {object} api.ErrorResponse "cannot reset password"
// @Router       /reset-password [post]
func (h *Handler) ResetUserPassword(w http.ResponseWriter, r *http.Request) {
	logger := zerolog.Ctx(r.Context())

	var request dto.NewPasswordRequest
	err := json.NewDecoder(r.Body).Decode(&request)
	if err != nil {
		api.RespondError(w, http.StatusBadRequest, ErrInvalidRequestSchema.Error())
		return
	}

	err = ValidatorRequestNewPassword(request.Password, request.RepeatedPassword, h.cfg.MaxLenPassword, h.cfg.MinLenPassword)
	if err != nil {
		api.RespondError(w, http.StatusBadRequest, ErrInvalidEmailOrPassword.Error())
		return
	}

	err = h.srv.ResetPassword(r.Context(), request.TokenID, request.Password)
	if err != nil {
		logger.Error().Err(err).Msg("auth.ResetPassword failed")

		if errors.Is(err, common.ErrorNotNullValue) {
			api.RespondError(w, http.StatusBadRequest, common.ErrorNotNullValue.Error())
			return
		}

		if errors.Is(err, common.ErrorNotExistingResetToken) {
			api.RespondError(w, http.StatusBadRequest, common.ErrorNotExistingResetToken.Error())
			return
		}

		if errors.Is(err, common.ErrorNonexistentUser) {
			api.RespondError(w, http.StatusNotFound, common.ErrorNonexistentUser.Error())
			return
		}

		api.RespondError(w, http.StatusInternalServerError, ErrCannotResetPassword.Error())
		return
	}

	api.Respond(w, http.StatusOK, api.StatusOK)
}

// @Summary      VK OAuth Callback
// @Description  Коллбэк для авторизации через ВК. Производит редирект на клиент с результатом и устанавливает cookie сессии.
// @Tags         auth
// @Param        code query string true "Код авторизации, полученный от ВКонтакте"
// @Success      302 "Успешная авторизация. Устанавливается Cookie и происходит редирект"
// @Header       302 {string} Location "URL редиректа на клиент"
// @Header       302 {string} Set-Cookie "Сессионная кука"
// @Failure      400 "Отсутствует код (ошибка: oauth_code_empty)"
// @Failure      500 "Внутренняя ошибка сервера"
// @Failure      502 "Ошибка обмена токена, API ВКонтакте или отсутствие email"
// @Router       /oauth/vk [get]
func (h *Handler) VkOAuthCallback(conf *config.VkOAuth, redirectTo string, vkOAuth VkOAuth) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		logger := zerolog.Ctx(r.Context())

		code := r.FormValue(oauthCodeKey)
		if code == "" {
			logger.Err(ErrOAuthCodeEmpty).Msg("vk oauth callback")
			Redirect(w, r, redirectTo, http.StatusBadRequest, ErrOAuthCodeEmpty.Error())
			return
		}

		token, err := vkOAuth.Exchange(r.Context(), code)
		if err != nil {
			logger.Err(err).Msg("vk oauth exchange")
			Redirect(w, r, redirectTo, http.StatusBadGateway, ErrOAuthExchangeFailed.Error())
			return
		}

		rawEmail := token.Extra(oauthEmailKey)
		if rawEmail == nil {
			Redirect(w, r, redirectTo, http.StatusBadGateway, ErrOAuthNoEmailProvided.Error())
			return
		}

		var ok bool
		var userEmail string
		if userEmail, ok = rawEmail.(string); !ok {
			Redirect(w, r, redirectTo, http.StatusBadGateway, ErrOAuthNoEmailProvided.Error())
			return
		}

		if ok := ValidateEmail(userEmail); !ok {
			Redirect(w, r, redirectTo, http.StatusBadGateway, ErrOAuthInvalidEmail.Error())
			return
		}

		client := vkOAuth.Client(r.Context(), token)
		res, err := client.Get(fmt.Sprintf(conf.APIMethod, token.AccessToken))
		if err != nil {
			logger.Err(err).Msg("vk api cannot request data")
			Redirect(w, r, redirectTo, http.StatusBadGateway, ErrOAuthCannotRequestUserData.Error())
			return
		}

		defer func() {
			if err := res.Body.Close(); err != nil {
				logger.Err(err).Msg("close response body")
			}
		}()

		usersData := &api.VkAPIUsersData{}
		if err := json.NewDecoder(res.Body).Decode(usersData); err != nil {
			logger.Err(err).Msg("vk api cannot read response body")
			Redirect(w, r, redirectTo, http.StatusInternalServerError, ErrOAuthInternalServerError.Error())
			return
		}

		if len(usersData.Response) < 1 {
			logger.Err(ErrOAuthEmptyUserData).Msg("vk api: empty user data")
			Redirect(w, r, redirectTo, http.StatusInternalServerError, ErrOAuthEmptyUserData.Error())
			return
		}

		userData := usersData.Response[0]

		registrationUserInfo := serviceDto.RegistrationUser{
			DisplayName: userData.FirstName,
			Email:       userEmail,
		}

		user, err := h.srv.EnsureUserByEmail(r.Context(), registrationUserInfo)
		if err != nil {
			logger.Err(err).Msg("authService.EnsureUserByEmail")
			Redirect(w, r, redirectTo, http.StatusInternalServerError, ErrOAuthInternalServerError.Error())
			return
		}

		userInfo := serviceDto.UserInfo{
			Link:        user.Link,
			DisplayName: user.DisplayName,
			Email:       user.Email,
			Avatar:      user.Avatar,
		}

		err = h.srv.SaveRefreshTokenFroUser(r.Context(), userInfo, token.RefreshToken)
		if err != nil {
			logger.Err(ErrOAuthCannotSaveRefreshToken).Msg("authService.SaveRefreshToken")
			Redirect(w, r, redirectTo, http.StatusInternalServerError, ErrOAuthInternalServerError.Error())
			return
		}

		sessionID, err := h.srv.CreateSessionForUser(r.Context(), user.Link)
		if err != nil {
			logger.Err(err).Msg("authService.CreateSessionForUser")
			Redirect(w, r, redirectTo, http.StatusInternalServerError, ErrOAuthInternalServerError.Error())
			return
		}

		http.SetCookie(w, api.NewSessionCookie(
			service.SessiondIdKey,
			sessionID,
			time.Now().Add(h.cfg.SessionLifetime)))

		Redirect(w, r, redirectTo, http.StatusOK, oauthSuccessAuthMessage)
	}
}

// @Summary      Установка CSRF куки
// @Description  Генерирует новый CSRF токен и записывает его в Cookie.
// @Tags         csrf
// @Produce      json
// @Success      200 {object} api.Response "Успешная установка куки (ok)"
// @Header       200 {string} Set-Cookie "csrf_token=...; Path=/; Secure; SameSite=Lax"
// @Failure      500 {object} api.ErrorResponse "cannot create csrf token"
// @Router       /csrf [get]
func (h *Handler) SetCSRFCookieHandler(w http.ResponseWriter, r *http.Request) {
	logger := zerolog.Ctx(r.Context())

	cookie, err := r.Cookie(service.SessiondIdKey)
	if err != nil {
		api.RespondError(w, http.StatusUnauthorized, ErrUserNotAuthorized.Error())
		return
	}
	sessionId := cookie.Value

	expireTime, err := h.srv.GetCSRFTokenExpireTime(r.Context())
	if err != nil {
		logger.Error().Err(ErrCannotGetCSRFTokenExpireTime).Msg("get csrf token expire time")
		api.RespondError(w, http.StatusInternalServerError, ErrCannotCreateCSRFToken.Error())
		return
	}

	token, err := h.srv.GenerateCSRFToken(r.Context(), sessionId, expireTime.Unix())
	if err != nil {
		logger.Error().Err(ErrCannotCreateCSRFToken).Msg("generate csrf token")
		api.RespondError(w, http.StatusInternalServerError, ErrCannotCreateCSRFToken.Error())
		return
	}

	http.SetCookie(w, api.NewCSRFCookie(csrfCookieKey, token, expireTime))

	api.Respond(w, http.StatusOK, api.StatusOK)
}
