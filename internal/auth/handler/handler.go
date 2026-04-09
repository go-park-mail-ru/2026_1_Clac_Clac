package handler

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/go-park-mail-ru/2026_1_Clac_Clac/internal/api"
	"github.com/go-park-mail-ru/2026_1_Clac_Clac/internal/auth/handler/dto"
	"github.com/go-park-mail-ru/2026_1_Clac_Clac/internal/auth/service"
	serviceDto "github.com/go-park-mail-ru/2026_1_Clac_Clac/internal/auth/service/dto"
	"github.com/go-park-mail-ru/2026_1_Clac_Clac/internal/common"
	"github.com/go-park-mail-ru/2026_1_Clac_Clac/internal/config"
	"github.com/go-park-mail-ru/2026_1_Clac_Clac/internal/middleware"
	"github.com/google/uuid"
	"github.com/rs/zerolog"
	"golang.org/x/oauth2"
)

// mockery --name=AuthService --output=mock_auth_srv --outpkg=mockAuthSrv
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
	GenerateRandomCSRFToken(ctx context.Context) (string, error)
}

// mockery --name=VkOAuth --output mock_vk_oauth
type VkOAuth interface {
	Exchange(ctx context.Context, code string, opts ...oauth2.AuthCodeOption) (*oauth2.Token, error)
	Client(ctx context.Context, t *oauth2.Token) *http.Client
}

type AuthHandler struct {
	Srv AuthService
}

func NewHandler(srv AuthService) *AuthHandler {
	return &AuthHandler{
		Srv: srv,
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

	ErrCannotCreateCSRFToken = errors.New("cannot create csrf token")
)

// MeHandler проверяет текущую сессию пользователя.
//
// @Summary      Проверка авторизации
// @Tags         auth
// @Produce      json
// @Success      200  {string}  string  "ok"
// @Failure      401  {object}  map[string]string "user not authorized"
// @Security     CookieAuth
// @Router       /me [get]
func (a *AuthHandler) MeHandler(w http.ResponseWriter, r *http.Request) {
	value := r.Context().Value(middleware.UserContextLink{})
	_, ok := value.(uuid.UUID)
	if !ok {
		api.RespondError(w, http.StatusUnauthorized, ErrUserNotAuthorized.Error())
		return
	}

	api.HandleError(api.Respond(w, http.StatusOK, api.StatusOK))
}

// LogInUser godoc
// @Summary      Вход в систему
// @Description  Аутентификация пользователя по email и паролю. Устанавливает HTTP-only cookie с сессией.
// @Tags         auth
// @Accept       json
// @Produce      json
// @Param        request body api.LogInRequest true "Учетные данные пользователя"
// @Success      200 {object} models.User "Успешная аутентификация"
// @Failure      400 {object} map[string]string "Некорректный запрос (невалидные данные)"
// @Failure      401 {object} map[string]string "Неверный email или пароль"
// @Failure      500 {object} map[string]string "Внутренняя ошибка сервера"
// @Router       /auth/login [post]
func (a *AuthHandler) LogInUser(w http.ResponseWriter, r *http.Request) {
	logger := zerolog.Ctx(r.Context())

	var request dto.LogInRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		api.RespondError(w, http.StatusBadRequest, ErrInvalidRequestSchema.Error())
		return
	}

	err := ValidatorRequestAuth(request.Email, request.Password)
	if err != nil {
		api.RespondError(w, http.StatusBadRequest, ErrInvalidEmailOrPassword.Error())
		return
	}

	serviceUser, sessionID, err := a.Srv.LogIn(r.Context(), serviceDto.LogInUser{
		Email:    request.Email,
		Password: request.Password,
	})
	if err != nil {
		if errors.Is(err, service.ErrorWrongPassword) {
			api.RespondError(w, http.StatusUnauthorized, ErrWrongEmailOrPassword.Error())
			return
		}

		logger.Err(fmt.Errorf("auth.Login: %w", err))
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
		time.Now().Add(service.SessionLifetime)))

	api.HandleError(api.RespondOk(w, handlerUser))
}

// RegisterUser godoc
// @Summary      Регистрация нового пользователя
// @Description  Создает новый аккаунт и сразу авторизует пользователя, выдавая cookie.
// @Tags         auth
// @Accept       json
// @Produce      json
// @Param        request body api.RegisterRequest true "Данные для регистрации"
// @Success      201 {object} models.User "Пользователь успешно создан"
// @Failure      400 {object} map[string]string "Ошибка валидации данных"
// @Failure      500 {object} map[string]string "Внутренняя ошибка сервера"
// @Router       /auth/register [post]
func (a *AuthHandler) RegisterUser(w http.ResponseWriter, r *http.Request) {
	logger := zerolog.Ctx(r.Context())

	var request dto.RegisterRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		api.RespondError(w, http.StatusBadRequest, ErrInvalidRequestSchema.Error())
		return
	}

	request.Sanitize()

	err := ValidatorWithCheckPassword(request.Email, request.Password, request.RepeatedPassword)
	if err != nil {
		api.RespondError(w, http.StatusBadRequest, ErrInvalidEmailOrPassword.Error())
		return
	}

	serviceUser, sessionID, err := a.Srv.Register(r.Context(), serviceDto.RegistrationUser{
		DisplayName: request.DisplayName,
		Email:       request.Email,
		Password:    request.Password,
	})
	if err != nil {
		logger.Err(fmt.Errorf("auth.Register: %w", err))
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
		time.Now().Add(service.SessionLifetime)))

	api.HandleError(api.RespondCreated(w, handlerUser))
}

// LogOutUser godoc
// @Summary      Выход из системы
// @Description  Удаляет сессию пользователя из хранилища и очищает cookie.
// @Tags         auth
// @Produce      json
// @Success      200 {object} map[string]string "Успешный выход"
// @Router       /auth/logout [post]
func (a *AuthHandler) LogOutUser(w http.ResponseWriter, r *http.Request) {
	logger := zerolog.Ctx(r.Context())

	cookie, err := r.Cookie(service.SessiondIdKey)
	if err == nil && cookie != nil {
		errLogOut := a.Srv.LogOut(r.Context(), cookie.Value)
		if errLogOut != nil {
			logger.Err(fmt.Errorf("Srv.LogOut: %w", errLogOut))
		}
	}

	http.SetCookie(w, api.NewExpiredCookie(service.SessiondIdKey))
	api.HandleError(api.Respond(w, http.StatusOK, api.StatusOK))
}

// SendRecoveryEmail godoc
// @Summary      Запрос восстановления пароля
// @Description  Генерирует код восстановления и отправляет его на указанный email.
// @Tags         auth
// @Accept       json
// @Produce      json
// @Param        request body api.PasswordRecoveryRequest true "Email пользователя"
// @Success      200 {object} map[string]string "Код успешно отправлен"
// @Failure      400 {object} map[string]string "Некорректный запрос"
// @Failure      500 {object} map[string]string "Ошибка отправки письма"
// @Router       /auth/recovery/send [post]
func (a *AuthHandler) SendRecoveryEmail(w http.ResponseWriter, r *http.Request) {
	logger := zerolog.Ctx(r.Context())

	var request dto.PasswordRecoveryRequest
	err := json.NewDecoder(r.Body).Decode(&request)
	if err != nil {
		api.RespondError(w, http.StatusBadRequest, ErrInvalidRequestSchema.Error())
		return
	}

	isAllowed, waitTime, err := a.Srv.CheckCoolDown(r.Context(), serviceDto.CoolDownConfig{
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

	err = a.Srv.SendRecoveryCode(r.Context(), request.Email)
	if err != nil {
		if errors.Is(err, common.ErrorNonexistentUser) {
			api.RespondError(w, http.StatusNotFound, ErrUserDoesNotExists.Error())
			return
		}

		logger.Err(fmt.Errorf("auth.SendRecoveryCode: %w", err))
		api.RespondError(w, http.StatusInternalServerError, ErrCannotSendRecoveryCode.Error())
		return
	}

	api.HandleError(api.Respond(w, http.StatusOK, api.StatusOK))
}

// CheckRecoveryCode godoc
// @Summary      Проверка кода восстановления
// @Description  Проверяет корректность 6-значного кода, отправленного на почту.
// @Tags         auth
// @Accept       json
// @Produce      json
// @Param        request body api.RecoveryCodeRequest true "Код из письма"
// @Success      200 {object} map[string]string "Код верен"
// @Failure      400 {object} map[string]string "Некорректный запрос"
// @Failure      500 {object} map[string]string "Неверный код или ошибка сервера"
// @Router       /auth/recovery/check [post]
func (a *AuthHandler) CheckRecoveryCode(w http.ResponseWriter, r *http.Request) {
	logger := zerolog.Ctx(r.Context())

	var request dto.RecoveryCodeRequest
	err := json.NewDecoder(r.Body).Decode(&request)
	if err != nil {
		api.RespondError(w, http.StatusBadRequest, ErrInvalidRequestSchema.Error())
		return
	}

	err = a.Srv.CheckRecoveryCode(r.Context(), request.Code)
	if err != nil {
		logger.Err(fmt.Errorf("auth.CheckRecoveryCode: %w", err))
		api.RespondError(w, http.StatusInternalServerError, ErrInternalServerError.Error())
		return
	}

	api.HandleError(api.Respond(w, http.StatusOK, api.StatusOK))
}

// ResetUserPassword godoc
//
//	@Summary		Сброс пароля
//	@Description	Устанавливает новый пароль пользователя с помощью проверенного токена.
//	@Tags			auth
//	@Accept			json
//	@Produce		json
//	@Param			request	body		api.NewPasswordRequest	true	"Новый пароль и токен"
//	@Success		200		{object}	map[string]string		"Пароль успешно изменен"
//	@Failure		400		{object}	map[string]string		"Некорректные данные"
//	@Failure		500		{object}	map[string]string		"Ошибка обновления пароля"
//	@Router			/reset-password [post]
func (a *AuthHandler) ResetUserPassword(w http.ResponseWriter, r *http.Request) {
	logger := zerolog.Ctx(r.Context())

	var request dto.NewPasswordRequest
	err := json.NewDecoder(r.Body).Decode(&request)
	if err != nil {
		api.RespondError(w, http.StatusBadRequest, ErrInvalidRequestSchema.Error())
		return
	}

	err = ValidatorRequestNewPassword(request.Password, request.RepeatedPassword)
	if err != nil {
		api.RespondError(w, http.StatusBadRequest, ErrInvalidEmailOrPassword.Error())
		return
	}

	err = a.Srv.ResetPassword(r.Context(), request.TokenID, request.Password)
	if err != nil {
		logger.Err(fmt.Errorf("auth.ResetPassword: %w", err))
		api.RespondError(w, http.StatusInternalServerError, ErrCannotResetPassword.Error())
		return
	}

	api.HandleError(api.Respond(w, http.StatusOK, api.StatusOK))
}

// VkOAuthCallback godoc
//
//	@Summary		VK OAuth Callback
//	@Description	Делает redirect 302, передает code и message через query параметры.
//	@Tags			auth
//	@Param			code	query	string	true	"Код авторизации, полученный от ВКонтакте"
//	@Success		302		"Успешная авторизация. Устанавливается Cookie и происходит редирект"
//	@Header			302		{string}	Location	"URL редиректа на клиент (например, /?code=200&message=success)"
//	@Header			302		{string}	Set-Cookie	"Сессионная кука"
//	@Failure		400		"Отсутствует code, email или email некорректный"
//	@Failure		500		"Внутренняя ошибка сервера"
//	@Failure		502		"Ошибка при обращении к API ВКонтакте"
//	@Router			/oauth/vk [get]
func (a *AuthHandler) VkOAuthCallback(conf *config.VkOAuth, redirectTo string, vkOAuth VkOAuth) func(http.ResponseWriter, *http.Request) {
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
		user, err := a.Srv.EnsureUserByEmail(r.Context(), registrationUserInfo)
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

		err = a.Srv.SaveRefreshTokenFroUser(r.Context(), userInfo, token.RefreshToken)
		if err != nil {
			logger.Err(ErrOAuthCannotSaveRefreshToken).Msg("authService.SaveRefreshToken")
			Redirect(w, r, redirectTo, http.StatusInternalServerError, ErrOAuthInternalServerError.Error())
			return
		}

		sessionID, err := a.Srv.CreateSessionForUser(r.Context(), user.Link)
		if err != nil {
			logger.Err(err).Msg("authService.CreateSessionForUser")
			Redirect(w, r, redirectTo, http.StatusInternalServerError, ErrOAuthInternalServerError.Error())
			return
		}

		http.SetCookie(w, api.NewSessionCookie(
			service.SessiondIdKey,
			sessionID,
			time.Now().Add(service.SessionLifetime)))

		Redirect(w, r, redirectTo, http.StatusOK, oauthSuccessAuthMessage)
	}
}

// SetCSRFCookieHandler устанавливает CSRF куку
//
//	@Summary		Установка CSRF куки
//	@Description	Генерирует новый CSRF токен и записывает его в Cookie.
//	@Description	Вместе с кукой также надо отправлять X-CSRF-Token
//	@Tags			csrf
//	@Produce		json
//	@Success		200	{object}	api.Response		"ok"	"Успешная установка куки"
//	@Header			200	{string}	Set-Cookie			"csrf_token=...; Path=/; Secure; SameSite=Lax"
//	@Failure		500	{object}	api.ErrorResponse	"internal server error - cannot create token"
//	@Router			/csrf [get]
func (a *AuthHandler) SetCSRFCookieHandler(w http.ResponseWriter, r *http.Request) {
	logger := zerolog.Ctx(r.Context())

	token, err := a.Srv.GenerateRandomCSRFToken(r.Context())
	if err != nil {
		logger.Error().Err(ErrCannotCreateCSRFToken).Msg("generate token")
		api.RespondError(w, http.StatusInternalServerError, ErrCannotCreateCSRFToken.Error())
		return
	}

	http.SetCookie(w, api.NewCSRFCookie(csrfCookieKey, token))

	api.HandleError(api.Respond(w, http.StatusOK, api.StatusOK))
}
