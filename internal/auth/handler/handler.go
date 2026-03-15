package handler

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"time"

	"github.com/go-park-mail-ru/2026_1_Clac_Clac/internal/api"
	"github.com/go-park-mail-ru/2026_1_Clac_Clac/internal/auth/models"
	"github.com/go-park-mail-ru/2026_1_Clac_Clac/internal/auth/service"
	"github.com/go-park-mail-ru/2026_1_Clac_Clac/internal/common"
	"github.com/go-park-mail-ru/2026_1_Clac_Clac/internal/config"
	"github.com/go-park-mail-ru/2026_1_Clac_Clac/internal/middleware"
	"github.com/google/uuid"
	"github.com/rs/zerolog"
	"golang.org/x/oauth2"
)

type AuthService interface {
	Register(ctx context.Context, name, password, email string) (models.User, string, error)
	LogIn(ctx context.Context, email, userID string) (models.User, string, error)
	CreateSessionForUser(ctx context.Context, user models.User) (string, error)
	LogOut(ctx context.Context, sessionID string) error
	GetUserID(ctx context.Context, sessionID string) (uuid.UUID, error)
	GetUserByEmail(ctx context.Context, email string) (models.User, error)
	SendRecoveryCode(ctx context.Context, email string) error
	CheckRecoveryCode(ctx context.Context, tokenID string) error
	ResetPassword(ctx context.Context, tokenID, newPassword string) error
}

type VkOAuth interface {
	Exchange(ctx context.Context, code string, opts ...oauth2.AuthCodeOption) (*oauth2.Token, error)
	Client(ctx context.Context, t *oauth2.Token) *http.Client
}

func NewAuthHandler(srv AuthService) *AuthHandler {
	return &AuthHandler{
		Srv: srv,
	}
}

type AuthHandler struct {
	Srv AuthService
}

const (
	InvalidDataMessage     = "invalid data"
	InvalidEmailOrPassword = "invalid email or password"
	WrongEmailOrPassword   = "wrong email or password"
	CannotSendEmail        = "cannot send email"
	CannotResetPassword    = "cannot reset password"
	SomethingWentWrong     = "something went wrong"
	UserNotAuthorized      = "user not authorized"
	UserDoesNotExists      = "user does not exists"
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
	value := r.Context().Value(middleware.UserIDKey{})
	_, ok := value.(uuid.UUID)
	if !ok {
		api.RespondError(w, http.StatusUnauthorized, UserNotAuthorized)
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

	var request api.LogInRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		api.RespondError(w, http.StatusBadRequest, InvalidDataMessage)
		return
	}

	err := ValidatorRequestAuth(request.Email, request.Password)
	if err != nil {
		api.RespondError(w, http.StatusBadRequest, InvalidEmailOrPassword)
		return
	}

	user, sessionID, err := a.Srv.LogIn(r.Context(), request.Email, request.Password)
	if err != nil {
		if errors.Is(err, service.ErrorWrongPassword) {
			api.RespondError(w, http.StatusUnauthorized, WrongEmailOrPassword)
			return
		}

		logger.Err(fmt.Errorf("auth.Login: %w", err))
		api.RespondError(w, http.StatusInternalServerError, SomethingWentWrong)
		return
	}

	http.SetCookie(w, api.NewCookie(
		service.SessiondIdKey,
		sessionID,
		time.Now().Add(service.SessionLifetime)))

	api.HandleError(api.RespondOk(w, user))
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

	var request api.RegisterRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		api.RespondError(w, http.StatusBadRequest, InvalidDataMessage)
		return
	}

	err := ValidatorWithCheckPassword(request.Email, request.Password, request.RepeatedPassword)
	if err != nil {
		api.RespondError(w, http.StatusBadRequest, InvalidEmailOrPassword)
		return
	}

	user, sessionID, err := a.Srv.Register(r.Context(), request.DisplayName, request.Password, request.Email)
	if err != nil {
		logger.Err(fmt.Errorf("auth.Register: %w", err))
		api.RespondError(w, http.StatusInternalServerError, SomethingWentWrong)
		return
	}

	http.SetCookie(w, api.NewCookie(
		service.SessiondIdKey,
		sessionID,
		time.Now().Add(service.SessionLifetime)))

	api.HandleError(api.RespondCreated(w, user))
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

	var request api.PasswordRecoveryRequest
	err := json.NewDecoder(r.Body).Decode(&request)
	if err != nil {
		api.RespondError(w, http.StatusBadRequest, InvalidDataMessage)
		return
	}

	err = a.Srv.SendRecoveryCode(r.Context(), request.Email)
	if err != nil {
		if errors.Is(err, common.ErrorNonexistentUser) {
			api.RespondError(w, http.StatusBadRequest, UserDoesNotExists)
			return
		}

		logger.Err(fmt.Errorf("auth.SendRecoveryCode: %w", err))
		api.RespondError(w, http.StatusInternalServerError, CannotSendEmail)
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

	var request api.RecoveryCodeRequest
	err := json.NewDecoder(r.Body).Decode(&request)
	if err != nil {
		api.RespondError(w, http.StatusBadRequest, InvalidDataMessage)
		return
	}

	err = a.Srv.CheckRecoveryCode(r.Context(), request.Code)
	if err != nil {
		logger.Err(fmt.Errorf("auth.CheckRecoveryCode: %w", err))
		api.RespondError(w, http.StatusInternalServerError, SomethingWentWrong)
		return
	}

	api.HandleError(api.Respond(w, http.StatusOK, api.StatusOK))
}

// ResetUserPassword godoc
// @Summary      Сброс пароля
// @Description  Устанавливает новый пароль пользователя с помощью проверенного токена.
// @Tags         auth
// @Accept       json
// @Produce      json
// @Param        request body api.NewPasswordRequest true "Новый пароль и токен"
// @Success      200 {object} map[string]string "Пароль успешно изменен"
// @Failure      400 {object} map[string]string "Некорректные данные"
// @Failure      500 {object} map[string]string "Ошибка обновления пароля"
// @Router       /auth/password/reset [post]
func (a *AuthHandler) ResetUserPassword(w http.ResponseWriter, r *http.Request) {
	logger := zerolog.Ctx(r.Context())

	var request api.NewPasswordRequest
	err := json.NewDecoder(r.Body).Decode(&request)
	if err != nil {
		api.RespondError(w, http.StatusBadRequest, InvalidDataMessage)
		return
	}

	err = ValidatorRequestNewPassword(request.Password, request.RepeatedPassword)
	if err != nil {
		api.RespondError(w, http.StatusBadRequest, InvalidEmailOrPassword)
		return
	}

	err = a.Srv.ResetPassword(r.Context(), request.TokenID, request.Password)
	if err != nil {
		logger.Err(fmt.Errorf("auth.ResetPassword: %w", err))
		api.RespondError(w, http.StatusInternalServerError, CannotResetPassword)
		return
	}

	api.HandleError(api.Respond(w, http.StatusOK, api.StatusOK))
}

func (a *AuthHandler) VkOAuthCallback(conf *config.VkOAuth, redirectTo string, vkOAuth VkOAuth) func(http.ResponseWriter, *http.Request) {
	redirectWithMessage := func(w http.ResponseWriter, r *http.Request, statusCode int, errorMessage string) {
		params := url.Values{}
		params.Add("message", errorMessage)
		params.Add("code", strconv.Itoa(statusCode))

		targetURL := fmt.Sprintf("%s?%s", redirectTo, params.Encode())

		http.Redirect(w, r, targetURL, http.StatusFound)
	}

	return func(w http.ResponseWriter, r *http.Request) {
		const vkOAuthCodeKey = "code"
		const emailKey = "email"
		const randomPasswordLength = 32

		ctx := r.Context()
		logger := zerolog.Ctx(ctx)

		code := r.FormValue(vkOAuthCodeKey)

		token, err := vkOAuth.Exchange(ctx, code)
		if err != nil {
			logger.Err(err).Msg("vk oauth exchange")
			redirectWithMessage(w, r, http.StatusBadRequest, "vk_oauth_error")
			return
		}

		var userEmail string
		var ok bool
		if emailRaw := token.Extra(emailKey); emailRaw != nil {
			if userEmail, ok = emailRaw.(string); !ok {
				redirectWithMessage(w, r, http.StatusBadRequest, "no_valid_email")
				return
			}
		}

		if ok := ValidateEmail(userEmail); !ok {
			redirectWithMessage(w, r, http.StatusBadRequest, "no_valid_email")
			return
		}

		client := vkOAuth.Client(ctx, token)
		res, err := client.Get(fmt.Sprintf(conf.APIMethod, token.AccessToken))
		if err != nil {
			logger.Err(err).Msg("vk api cannot request data")
			redirectWithMessage(w, r, http.StatusBadRequest, "cannot_request_data")
			return
		}

		defer func() {
			if err := res.Body.Close(); err != nil {
				logger.Err(err).Msg("close response body")
			}
		}()

		usersData := &api.VkAPIUsersData{}
		if err := json.NewDecoder(res.Body).Decode(usersData); err != nil {
			logger.Err(err).Msg("cannot read response body")
			redirectWithMessage(w, r, http.StatusInternalServerError, "something_went_wrong")
			return
		}

		if len(usersData.Response) < 1 {
			logger.Err(errors.New("cannot find user")).Msg("read user data from vk api")
			redirectWithMessage(w, r, http.StatusInternalServerError, "something_went_wrong")
			return
		}

		userData := usersData.Response[0]

		var sessionID string
		user, err := a.Srv.GetUserByEmail(ctx, userEmail)
		if err != nil {
			if errors.Is(err, common.ErrorNonexistentUser) {
				b := make([]byte, randomPasswordLength)
				if _, err := rand.Read(b); err != nil {
					logger.Err(err).Msg("vk oauth generate user password")
					redirectWithMessage(w, r, http.StatusInternalServerError, "something_went_wrong")
					return
				}

				password := base64.URLEncoding.EncodeToString(b)

				user, sessionID, err = a.Srv.Register(r.Context(), userData.FirstName, password, userEmail)
				if err != nil {
					logger.Err(fmt.Errorf("auth.Register: %w", err))
					redirectWithMessage(w, r, http.StatusInternalServerError, "something_went_wrong")
					return
				}
			} else {
				logger.Err(err).Msg("service.GetUser")
				redirectWithMessage(w, r, http.StatusInternalServerError, "something_went_wrong")
				return
			}
		} else {
			sessionID, err = a.Srv.CreateSessionForUser(ctx, user)
			if err != nil {
				logger.Err(err).Msg("service.CreateSessionForUser")
				redirectWithMessage(w, r, http.StatusInternalServerError, "something_went_wrong")
				return
			}
		}

		http.SetCookie(w, api.NewCookie(
			service.SessiondIdKey,
			sessionID,
			time.Now().Add(service.SessionLifetime)))

		redirectWithMessage(w, r, http.StatusOK, "success")
	}
}
