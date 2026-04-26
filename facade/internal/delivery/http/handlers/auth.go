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

// AuthUseCase covers auth-related operations delegated to the usecase layer.
type AuthUseCase interface {
	Login(ctx context.Context, cred domain.Credentials) (domain.UserInfo, string, error)
	Register(ctx context.Context, cred domain.NewCredentialsUser) (domain.UserInfo, string, error)
	LoginWithVK(ctx context.Context, code string) (domain.UserInfo, string, error)
	Logout(ctx context.Context, sessionID string) error
	SendRecoveryCode(ctx context.Context, email string) error
	CheckRecoveryCode(ctx context.Context, code string) error
	ResetPassword(ctx context.Context, tokenID, newPassword string) error
}

// CoolDownUseCase wraps rate-limiting / cooldown logic.
type CoolDownUseCase interface {
	CheckCoolDown(ctx context.Context, cooldown domain.Cooldown) (domain.CooldownResult, error)
}

// CSRFUseCase generates and verifies CSRF tokens.
type CSRFUseCase interface {
	GetExpireTime() time.Time
	Generate(ctx context.Context, sessionID string, expireAt int64) (string, error)
	Check(ctx context.Context, sessionID string, token string) error
}

// AuthConfig holds handler-level configuration.
type AuthConfig struct {
	MaxLenPassword    int
	MinLenPassword    int
	SessionLifetime   time.Duration
	VKOAuthRedirectTo string
}

// Auth handles authentication HTTP endpoints.
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

// MeHandler returns 200 when the request passes the auth middleware.
func (a *Auth) MeHandler(w http.ResponseWriter, r *http.Request) {
	value := r.Context().Value(middleware.UserContextLink{})
	_, ok := value.(uuid.UUID)
	if !ok {
		api.RespondError(w, http.StatusUnauthorized, handlerCommon.ErrUserNotAuthorized.Error())
		return
	}

	api.HandleError(api.Respond(w, http.StatusOK, api.StatusOK))
}

// LogInUser authenticates a user by email + password and sets a session cookie.
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

// RegisterUser creates a new account and returns a session cookie.
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

// LogOutUser deletes the session and clears the session cookie.
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

// SendRecoveryEmail sends a password-recovery code to the user's email.
func (a *Auth) SendRecoveryEmail(w http.ResponseWriter, r *http.Request) {
	logger := zerolog.Ctx(r.Context())

	var request dto.PasswordRecoveryRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		api.RespondError(w, http.StatusBadRequest, handlerCommon.ErrInvalidRequestSchema.Error())
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

// CheckRecoveryCode verifies the code sent to the user's email.
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

// ResetUserPassword sets a new password using a previously verified reset token.
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

// VkOAuthCallback handles the redirect from VK OAuth.
// It delegates all VK API interaction to the auth usecase.
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
		code := http.StatusInternalServerError
		if errors.Is(err, common.ErrorVKOAuthUnavailable) {
			code = http.StatusBadGateway
		}
		Redirect(w, r, a.cfg.VKOAuthRedirectTo, code, handlerCommon.ErrOAuthInternalServerError.Error())
		return
	}

	http.SetCookie(w, api.NewSessionCookie(
		middleware.SessiondIdKey,
		sessionID,
		time.Now().Add(a.cfg.SessionLifetime)))

	_ = userInfo
	Redirect(w, r, a.cfg.VKOAuthRedirectTo, http.StatusOK, oauthSuccessAuthMessage)
}

// SetCSRFCookieHandler generates a CSRF token and sets it as a cookie.
func (a *Auth) SetCSRFCookieHandler(w http.ResponseWriter, r *http.Request) {
	logger := zerolog.Ctx(r.Context())

	cookie, err := r.Cookie(middleware.SessiondIdKey)
	if err != nil {
		api.RespondError(w, http.StatusUnauthorized, handlerCommon.ErrUserNotAuthorized.Error())
		return
	}
	sessionID := cookie.Value

	expireTime := a.csrf.GetExpireTime()

	token, err := a.csrf.Generate(r.Context(), sessionID, expireTime.Unix())
	if err != nil {
		logger.Error().Err(err).Msg("generate csrf token")
		api.RespondError(w, http.StatusInternalServerError, handlerCommon.ErrCannotCreateCSRFToken.Error())
		return
	}

	http.SetCookie(w, api.NewCSRFCookie(csrfCookieKey, token, expireTime))
	api.Respond(w, http.StatusOK, api.StatusOK)
}
