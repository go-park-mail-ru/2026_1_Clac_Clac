package auth

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

	"golang.org/x/oauth2"

	"github.com/go-park-mail-ru/2026_1_Clac_Clac/internal/api"
	"github.com/go-park-mail-ru/2026_1_Clac_Clac/internal/common"
	"github.com/go-park-mail-ru/2026_1_Clac_Clac/internal/config"
	"github.com/go-park-mail-ru/2026_1_Clac_Clac/internal/middleware"
	"github.com/go-park-mail-ru/2026_1_Clac_Clac/internal/models"
	service "github.com/go-park-mail-ru/2026_1_Clac_Clac/internal/service/auth"
	"github.com/google/uuid"
	"github.com/rs/zerolog"
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
		srv: srv,
	}
}

type AuthHandler struct {
	srv AuthService
}

const (
	invalidDataMessage     = "invalid data"
	invalidEmailOrPassword = "invalid email or password"
	wrongEmailOrPassword   = "wrong email or password"
	cannotSendEmail        = "cannot send email"
	cannotResetPassword    = "cannot reset password"
	somethingWentWrong     = "something went wrong"
	userNotAuthorized      = "user not authorized"
	userDoesNotExists      = "user does not exists"
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
		api.RespondError(w, http.StatusUnauthorized, userNotAuthorized)
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
		api.RespondError(w, http.StatusBadRequest, invalidDataMessage)
		return
	}

	err := ValidatorRequestAuth(request.Email, request.Password)
	if err != nil {
		api.RespondError(w, http.StatusBadRequest, invalidEmailOrPassword)
		return
	}

	user, sessionID, err := a.srv.LogIn(r.Context(), request.Email, request.Password)
	if err != nil {
		if errors.Is(err, service.ErrorWrongPassword) {
			api.RespondError(w, http.StatusUnauthorized, wrongEmailOrPassword)
			return
		}

		logger.Err(fmt.Errorf("auth.Login: %w", err))
		api.RespondError(w, http.StatusInternalServerError, somethingWentWrong)
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
		api.RespondError(w, http.StatusBadRequest, invalidDataMessage)
		return
	}

	err := ValidatorWithCheckPassword(request.Email, request.Password, request.RepeatedPassword)
	if err != nil {
		api.RespondError(w, http.StatusBadRequest, invalidEmailOrPassword)
		return
	}

	user, sessionID, err := a.srv.Register(r.Context(), request.DisplayName, request.Password, request.Email)
	if err != nil {
		logger.Err(fmt.Errorf("auth.Register: %w", err))
		api.RespondError(w, http.StatusInternalServerError, somethingWentWrong)
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
		errLogOut := a.srv.LogOut(r.Context(), cookie.Value)
		if errLogOut != nil {
			logger.Err(fmt.Errorf("srv.LogOut: %w", errLogOut))
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
		api.RespondError(w, http.StatusBadRequest, invalidDataMessage)
		return
	}

	err = a.srv.SendRecoveryCode(r.Context(), request.Email)
	if err != nil {
		if errors.Is(err, common.ErrorNonexistentUser) {
			api.RespondError(w, http.StatusBadRequest, userDoesNotExists)
			return
		}

		logger.Err(fmt.Errorf("auth.SendRecoveryCode: %w", err))
		api.RespondError(w, http.StatusInternalServerError, cannotSendEmail)
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
		api.RespondError(w, http.StatusBadRequest, invalidDataMessage)
		return
	}

	err = a.srv.CheckRecoveryCode(r.Context(), request.Code)
	if err != nil {
		logger.Err(fmt.Errorf("auth.CheckRecoveryCode: %w", err))
		api.RespondError(w, http.StatusInternalServerError, somethingWentWrong)
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
		api.RespondError(w, http.StatusBadRequest, invalidDataMessage)
		return
	}

	err = ValidatorRequestNewPassword(request.Password, request.RepeatedPassword)
	if err != nil {
		api.RespondError(w, http.StatusBadRequest, invalidEmailOrPassword)
		return
	}

	err = a.srv.ResetPassword(r.Context(), request.TokenID, request.Password)
	if err != nil {
		logger.Err(fmt.Errorf("auth.ResetPassword: %w", err))
		api.RespondError(w, http.StatusInternalServerError, cannotResetPassword)
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
		user, err := a.srv.GetUserByEmail(ctx, userEmail)
		if err != nil {
			if errors.Is(err, common.ErrorNonexistentUser) {
				b := make([]byte, randomPasswordLength)
				if _, err := rand.Read(b); err != nil {
					logger.Err(err).Msg("vk oauth generate user password")
					redirectWithMessage(w, r, http.StatusInternalServerError, "something_went_wrong")
					return
				}

				password := base64.URLEncoding.EncodeToString(b)

				user, sessionID, err = a.srv.Register(r.Context(), userData.FirstName, password, userEmail)
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
			sessionID, err = a.srv.CreateSessionForUser(ctx, user)
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
