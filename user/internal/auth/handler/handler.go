package handler

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"time"

	pb "github.com/go-park-mail-ru/2026_1_Clac_Clac/pkg/auth"
	"github.com/go-park-mail-ru/2026_1_Clac_Clac/user/internal/api"
	"github.com/go-park-mail-ru/2026_1_Clac_Clac/user/internal/auth/service"
	serviceDto "github.com/go-park-mail-ru/2026_1_Clac_Clac/user/internal/auth/service/dto"
	"github.com/go-park-mail-ru/2026_1_Clac_Clac/user/internal/common"
	"github.com/google/uuid"
	"github.com/microcosm-cc/bluemonday"
	"github.com/rs/zerolog"
	"golang.org/x/oauth2"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

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
	// SendRecoveryCode(ctx context.Context, email string) error
	CheckRecoveryCode(ctx context.Context, tokenID string) error
	ResetPassword(ctx context.Context, tokenID, newPassword string) error
	EnsureUserByEmail(ctx context.Context, info serviceDto.RegistrationUser) (serviceDto.UserInfo, error)
	SaveRefreshTokenFroUser(ctx context.Context, info serviceDto.UserInfo, token string) error
	GetCSRFTokenExpireTime(ctx context.Context) (time.Time, error)
	GenerateCSRFToken(ctx context.Context, sessionId string, expireTime int64) (string, error)
	CheckCSRFToken(ctx context.Context, sessionId string, token string) error
}

type VkOAuth interface {
	Exchange(ctx context.Context, code string, opts ...oauth2.AuthCodeOption) (*oauth2.Token, error)
	Client(ctx context.Context, t *oauth2.Token) *http.Client
}

const (
	oauthCodeKey            = "code"
	oauthEmailKey           = "email"
	oauthSuccessAuthMessage = "success"
	csrfCookieKey           = "csrf_token"
	nameCoolDown            = "recovery_email"

	msgInternalError = "something went wrong"
	msgInvalidInput  = "invalid input parameters"

	msgInvalidRequestSchema    = "invalid schema"
	msgInvalidEmailOrPassword  = "invalid email or password"
	msgWrongEmailOrPassword    = "wrong email or password"
	msgInvalidNewPassword      = "invalid password or repeated password"
	msgCannotSendRecoveryCode  = "cannot send recovery code"
	msgCannotResetPassword     = "cannot reset password"
	msgResetTokenDoesNotExists = "reset token does not exist"
	msgInternalServerError     = "something went wrong"
	msgUserNotAuthorized       = "user not authorized"
	msgUserDoesNotExists       = "user does not exist"
	msgUserAlreadyExists       = "user already exists"
	msgNullInNotNullField      = "put null value in not null field"

	msgOAuthCodeEmpty              = "oauth_code_empty"
	msgOAuthExchangeFailed         = "oauth_error"
	msgOAuthNoEmailProvided        = "oauth_no_email"
	msgOAuthInvalidEmail           = "oauth_invalid_email"
	msgOAuthCannotRequestUserData  = "oauth_cannot_request_user_data"
	msgOAuthEmptyUserData          = "oauth_no_user_data"
	msgOAuthInternalServerError    = "oauth_something_went_wrong"
	msgOAuthCannotSaveRefreshToken = "oauth cannot save refresh token"

	msgCannotCreateCSRFToken        = "cannot create csrf token"
	msgCannotGetCSRFTokenExpireTime = "cannot get csrf token expire time"
)

type Config struct {
	MaxLenPassword  int
	MinLenPassword  int
	SessionLifetime time.Duration

	APIMethod string
}

type Handler struct {
	srv       AuthService
	cfg       Config
	vkOAuth   VkOAuth
	sanitizer *bluemonday.Policy
	pb.UnimplementedAuthServiceServer
}

func NewHandler(srv AuthService, cfg Config, vkOAuth VkOAuth) *Handler {
	return &Handler{
		srv:       srv,
		cfg:       cfg,
		sanitizer: bluemonday.StrictPolicy(),
		vkOAuth:   vkOAuth,
	}
}

func (h *Handler) LogInUser(ctx context.Context, req *pb.LogInRequest) (*pb.UserResponse, error) {
	logger := zerolog.Ctx(ctx)

	err := ValidatorRequestAuth(req.Email, req.Password, h.cfg.MaxLenPassword, h.cfg.MinLenPassword)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, msgInvalidEmailOrPassword)
	}

	serviceUser, sessionID, err := h.srv.LogIn(ctx, serviceDto.LogInUser{
		Email:    req.Email,
		Password: req.Password,
	})
	if err != nil {
		if errors.Is(err, service.ErrorWrongPassword) {
			return nil, status.Error(codes.InvalidArgument, msgWrongEmailOrPassword)
		}

		logger.Err(fmt.Errorf("auth.Login: %w", err)).Msg("auth handler log in user")
		return nil, status.Error(codes.Internal, msgInternalError)
	}

	return &pb.UserResponse{
		UserLink:    serviceUser.Link.String(),
		DisplayName: serviceUser.DisplayName,
		Email:       serviceUser.Email,
		Avatar:      serviceUser.Avatar,
		SessionId:   sessionID,
	}, nil
}

func (h *Handler) RegisterUser(ctx context.Context, req *pb.RegisterRequest) (*pb.UserResponse, error) {
	logger := zerolog.Ctx(ctx)

	sanitizedUser := serviceDto.RegistrationUser{
		DisplayName: req.DisplayName,
		Password:    req.Password,
		Email:       req.Email,
	}

	sanitizedUser.Sanitize(h.sanitizer)

	err := ValidatorWithCheckPassword(sanitizedUser.Email, req.Password, req.RepeatPassword, h.cfg.MaxLenPassword, h.cfg.MinLenPassword)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, msgInvalidEmailOrPassword)
	}

	serviceUser, sessionID, err := h.srv.Register(ctx, serviceDto.RegistrationUser{
		DisplayName: sanitizedUser.DisplayName,
		Email:       sanitizedUser.Email,
		Password:    req.Password,
	})

	if err != nil {
		logger.Err(fmt.Errorf("auth.Register: %w", err))

		if errors.Is(err, common.ErrorExistingUser) {
			return nil, status.Error(codes.AlreadyExists, msgUserAlreadyExists)
		}
		if errors.Is(err, common.ErrorNotNullValue) {
			return nil, status.Error(codes.InvalidArgument, msgNullInNotNullField)
		}

		return nil, status.Error(codes.Internal, msgInternalError)
	}

	return &pb.UserResponse{
		UserLink:    serviceUser.Link.String(),
		DisplayName: serviceUser.DisplayName,
		Email:       serviceUser.Email,
		Avatar:      serviceUser.Avatar,
		SessionId:   sessionID,
	}, nil
}

func (h *Handler) LogOutUser(ctx context.Context, req *pb.SessionRequest) (*pb.LogOutResponse, error) {
	logger := zerolog.Ctx(ctx)

	errLogOut := h.srv.LogOut(ctx, req.SessionId)
	if errLogOut != nil {
		logger.Err(fmt.Errorf("srv.LogOut: %w", errLogOut))
	}

	return &pb.LogOutResponse{}, nil
}

func (h *Handler) CheckRecoveryCode(ctx context.Context, req *pb.TokenRequest) (*pb.CheckRecoveryCodeResponse, error) {
	logger := zerolog.Ctx(ctx)

	err := h.srv.CheckRecoveryCode(ctx, req.TokenId)
	if err != nil {
		if errors.Is(err, common.ErrorNotExistingResetToken) {
			return nil, status.Error(codes.NotFound, msgResetTokenDoesNotExists)
		}

		logger.Error().Err(err).Msg("auth.CheckRecoveryCode failed")
		return nil, status.Error(codes.Internal, msgInternalError)
	}

	return &pb.CheckRecoveryCodeResponse{}, nil
}

func (h *Handler) ResetPassword(ctx context.Context, req *pb.ResetPasswordRequest) (*pb.ResetPasswordResponse, error) {
	logger := zerolog.Ctx(ctx)

	err := ValidatorRequestNewPassword(req.Password, req.RepeatedPassword, h.cfg.MaxLenPassword, h.cfg.MinLenPassword)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, msgInvalidNewPassword)
	}

	err = h.srv.ResetPassword(ctx, req.TokenId, req.Password)
	if err != nil {
		logger.Error().Err(err).Msg("auth.ResetPassword failed")

		if errors.Is(err, common.ErrorNotNullValue) {
			return nil, status.Error(codes.InvalidArgument, msgNullInNotNullField)
		}

		if errors.Is(err, common.ErrorNotExistingResetToken) {
			return nil, status.Error(codes.NotFound, msgResetTokenDoesNotExists)
		}

		if errors.Is(err, common.ErrorNonexistentUser) {
			return nil, status.Error(codes.NotFound, msgUserDoesNotExists)
		}

		return nil, status.Error(codes.Internal, msgInternalError)
	}

	return &pb.ResetPasswordResponse{}, nil
}

func (h *Handler) LoginWithVK(ctx context.Context, req *pb.VKLoginRequest) (*pb.LoginVKResponse, error) {
	logger := zerolog.Ctx(ctx)

	token, err := h.vkOAuth.Exchange(ctx, req.Code)
	if err != nil {
		logger.Err(err).Msg("vk oauth exchange failed")
		return nil, status.Error(codes.Unavailable, "vk oauth exchange")
	}

	rawEmail := token.Extra(oauthEmailKey)
	if rawEmail == nil {
		return nil, status.Error(codes.Unavailable, msgOAuthNoEmailProvided)
	}

	var ok bool
	var userEmail string
	if userEmail, ok = rawEmail.(string); !ok {
		return nil, status.Error(codes.Unavailable, msgOAuthNoEmailProvided)
	}

	if ok := ValidateEmail(userEmail); !ok {
		return nil, status.Error(codes.Unavailable, msgOAuthInvalidEmail)
	}

	client := h.vkOAuth.Client(ctx, token)
	res, err := client.Get(fmt.Sprintf(h.cfg.APIMethod, token.AccessToken))
	if err != nil {
		logger.Err(err).Msg("vk api cannot request data")
		return nil, status.Error(codes.Unavailable, msgOAuthCannotRequestUserData)
	}

	defer func() {
		if err := res.Body.Close(); err != nil {
			logger.Err(err).Msg("close response body")
		}
	}()

	usersData := &api.VkAPIUsersData{}
	if err := json.NewDecoder(res.Body).Decode(usersData); err != nil {
		logger.Err(err).Msg("vk api cannot read response body")
		return nil, status.Error(codes.Internal, msgInternalServerError)
	}

	if len(usersData.Response) < 1 {
		logger.Error().Msg("vk api: empty user data")
		return nil, status.Error(codes.Internal, msgOAuthEmptyUserData)
	}

	userData := usersData.Response[0]

	registrationUserInfo := serviceDto.RegistrationUser{
		DisplayName: userData.FirstName,
		Email:       userEmail,
	}

	user, err := h.srv.EnsureUserByEmail(ctx, registrationUserInfo)
	if err != nil {
		logger.Err(err).Msg("authService.EnsureUserByEmail")
		return nil, status.Error(codes.Internal, msgInternalServerError)
	}

	userInfo := serviceDto.UserInfo{
		Link:        user.Link,
		DisplayName: user.DisplayName,
		Email:       user.Email,
		Avatar:      user.Avatar,
	}

	err = h.srv.SaveRefreshTokenFroUser(ctx, userInfo, token.RefreshToken)
	if err != nil {
		logger.Err(err).Msg("authService.SaveRefreshToken")
		return nil, status.Error(codes.Internal, msgOAuthCannotSaveRefreshToken)
	}

	sessionID, err := h.srv.CreateSessionForUser(ctx, user.Link)
	if err != nil {
		logger.Err(err).Msg("authService.CreateSessionForUser")
		return nil, status.Error(codes.Internal, msgInternalServerError)
	}

	return &pb.LoginVKResponse{
		SessionId: sessionID,
	}, nil
}

func (h *Handler) SetCSRFCookieHandler(ctx context.Context, req *pb.CSRFRequest) (*pb.CSRFResponse, error) {
	logger := zerolog.Ctx(ctx)

	expireTime, err := h.srv.GetCSRFTokenExpireTime(ctx)
	if err != nil {
		logger.Error().Err(err).Msg(msgCannotGetCSRFTokenExpireTime)
		return nil, status.Error(codes.Internal, msgCannotCreateCSRFToken)
	}

	token, err := h.srv.GenerateCSRFToken(ctx, req.SessionId, expireTime.Unix())
	if err != nil {
		logger.Error().Err(err).Msg(msgCannotCreateCSRFToken)
		return nil, status.Error(codes.Internal, msgCannotCreateCSRFToken)
	}

	return &pb.CSRFResponse{
		Token: token,
	}, nil
}
