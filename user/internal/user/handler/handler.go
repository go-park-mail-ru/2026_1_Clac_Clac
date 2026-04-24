package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"

	pb "github.com/go-park-mail-ru/2026_1_Clac_Clac/pkg/contracts/user"
	"github.com/go-park-mail-ru/2026_1_Clac_Clac/user/internal/api"
	"github.com/go-park-mail-ru/2026_1_Clac_Clac/user/internal/common"
	"github.com/go-park-mail-ru/2026_1_Clac_Clac/user/internal/user/service"
	serviceDto "github.com/go-park-mail-ru/2026_1_Clac_Clac/user/internal/user/service/dto"
	"github.com/google/uuid"
	"github.com/microcosm-cc/bluemonday"
	"github.com/rs/zerolog"
	"golang.org/x/oauth2"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

const (
	oauthEmailKey = "email"

	msgInternalError = "something went wrong"
	msgInvalidInput  = "invalid input parameters"

	msgInvalidEmailOrPassword  = "invalid email or password"
	msgWrongEmailOrPassword    = "wrong email or password"
	msgInvalidNewPassword      = "invalid password or repeated password"
	msgResetTokenDoesNotExists = "reset token does not exist"
	msgUserDoesNotExists       = "user does not exist"
	msgEmailDoesNotExists      = "user with this email does not exist"
	msgUserAlreadyExists       = "user already exists"
	msgNullInNotNullField      = "put null value in not null field"

	msgOAuthNoEmailProvided       = "oauth_no_email"
	msgOAuthInvalidEmail          = "oauth_invalid_email"
	msgOAuthCannotRequestUserData = "oauth_cannot_request_user_data"
	msgOAuthEmptyUserData         = "oauth_no_user_data"

	msgFailParseUserLink   = "user link can not convert to uuid"
	msgTooLargeAvatar      = "size avatar too large"
	msgEmptyFile           = "empty file provided"
	msgIncorrectTypeAvatar = "avatar can be only jpeg/png/jpg/webp"
	msgFailFoundUser       = "user not found"
	msgFailNullValue       = "get null, but wait not null"
	msgInvalidProfileData  = "invalid profile data"
)

type AuthService interface {
	Register(ctx context.Context, requestUser serviceDto.RegistrationUser) (serviceDto.UserInfo, error)
	LogIn(ctx context.Context, requestUser serviceDto.LogInUser) (serviceDto.UserInfo, error)
	GetUserLink(ctx context.Context, email string) (string, error)
	EnsureUserByEmail(ctx context.Context, info serviceDto.RegistrationUser) (string, error)
	ResetPassword(ctx context.Context, passwordInfo serviceDto.ResetPasswordInfo) error

	GetProfile(ctx context.Context, userLink uuid.UUID) (serviceDto.UserInfo, error)
	UpdateProfile(ctx context.Context, updatedInfo serviceDto.UpdatedUserInfo) error
	UpdateAvatar(ctx context.Context, avatar serviceDto.UpdatedAvatar) (string, error)
	DeleteAvatar(ctx context.Context, userLink uuid.UUID) error
}

type VkOAuth interface {
	Exchange(ctx context.Context, code string, opts ...oauth2.AuthCodeOption) (*oauth2.Token, error)
	Client(ctx context.Context, t *oauth2.Token) *http.Client
}

type Config struct {
	MaxLenPassword int
	MinLenPassword int

	ValidExtensions       map[string]struct{}
	SignatureTypeBytes    int
	MaxLenNameUser        int
	MaxLenDescriptionUser int
	MaxReadBytes          int64

	APIMethod string
}

type Handler struct {
	srv       AuthService
	cfg       Config
	vkOAuth   VkOAuth
	sanitizer *bluemonday.Policy
	pb.UnimplementedUserServiceServer
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

	serviceUser, err := h.srv.LogIn(ctx, serviceDto.LogInUser{
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
		Avatar:      serviceUser.AvatarURL,
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

	err := ValidatorWithCheckPassword(sanitizedUser.Email, req.Password, req.RepeatedPassword, h.cfg.MaxLenPassword, h.cfg.MinLenPassword)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, msgInvalidEmailOrPassword)
	}

	serviceUser, err := h.srv.Register(ctx, serviceDto.RegistrationUser{
		DisplayName: sanitizedUser.DisplayName,
		Email:       sanitizedUser.Email,
		Password:    req.Password,
	})
	if err != nil {
		logger.Err(fmt.Errorf("auth.Register: %w", err)).Msg("register user")

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
		Avatar:      serviceUser.AvatarURL,
	}, nil
}

func (h *Handler) GetUserLink(ctx context.Context, req *pb.GetUserLinkRequest) (*pb.GetUserLinkResponse, error) {
	userLink, err := h.srv.GetUserLink(ctx, req.Email)
	if err != nil {
		if errors.Is(err, common.ErrorNonexistentEmail) {
			return nil, status.Error(codes.NotFound, msgEmailDoesNotExists)
		}
		return nil, status.Error(codes.Internal, msgInternalError)
	}

	return &pb.GetUserLinkResponse{
		UserLink: userLink,
	}, nil
}

func (h *Handler) ResetPassword(ctx context.Context, req *pb.ResetPasswordRequest) (*pb.ResetPasswordResponse, error) {
	logger := zerolog.Ctx(ctx)

	err := ValidatorRequestNewPassword(req.Password, req.RepeatedPassword, h.cfg.MaxLenPassword, h.cfg.MinLenPassword)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, msgInvalidNewPassword)
	}

	err = h.srv.ResetPassword(ctx, serviceDto.ResetPasswordInfo{
		UserLink:    req.UserLink,
		NewPassword: req.Password,
	})
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

func (h *Handler) LoginWithVK(ctx context.Context, req *pb.VKLoginRequest) (*pb.VKLoginResponse, error) {
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

	userEmail, ok := rawEmail.(string)
	if !ok {
		return nil, status.Error(codes.Unavailable, msgOAuthNoEmailProvided)
	}

	if !ValidateEmail(userEmail) {
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
		return nil, status.Error(codes.Internal, msgInternalError)
	}

	if len(usersData.Response) < 1 {
		logger.Error().Msg("vk api: empty user data")
		return nil, status.Error(codes.Internal, msgOAuthEmptyUserData)
	}

	userData := usersData.Response[0]

	userLink, err := h.srv.EnsureUserByEmail(ctx, serviceDto.RegistrationUser{
		DisplayName: userData.FirstName,
		Email:       userEmail,
	})
	if err != nil {
		logger.Err(err).Msg("authService.EnsureUserByEmail")
		return nil, status.Error(codes.Internal, msgInternalError)
	}

	return &pb.VKLoginResponse{
		UserLink: userLink,
	}, nil
}

func (h *Handler) GetProfile(ctx context.Context, req *pb.UserLinkRequest) (*pb.ProfileResponse, error) {
	parseUserLink, err := uuid.Parse(req.UserLink)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, msgFailParseUserLink)
	}

	serviceUser, err := h.srv.GetProfile(ctx, parseUserLink)
	if err != nil {
		if errors.Is(err, common.ErrorNonexistentUser) {
			return nil, status.Error(codes.NotFound, msgFailFoundUser)
		}

		return nil, status.Error(codes.Internal, msgInternalError)
	}

	return &pb.ProfileResponse{
		UserLink:    serviceUser.Link.String(),
		Email:       serviceUser.Email,
		DisplayName: serviceUser.DisplayName,
		Description: serviceUser.Description,
		AvatarUrl:   serviceUser.AvatarURL,
	}, nil
}

func (h *Handler) UpdateProfile(ctx context.Context, req *pb.UpdateProfileRequest) (*pb.UpdateProfileResponse, error) {
	parseUserLink, err := uuid.Parse(req.UserLink)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, msgFailParseUserLink)
	}

	cleanDisplayName := h.sanitizer.Sanitize(strings.TrimSpace(req.DisplayName))
	cleanDescription := h.sanitizer.Sanitize(strings.TrimSpace(req.Description))

	if err = common.ValidateTextInfo(cleanDisplayName, h.cfg.MaxLenNameUser); err != nil {
		return nil, status.Error(codes.InvalidArgument, msgInvalidInput)
	}

	if err = common.ValidateTextInfo(cleanDescription, h.cfg.MaxLenDescriptionUser); err != nil {
		return nil, status.Error(codes.InvalidArgument, msgInvalidInput)
	}

	err = h.srv.UpdateProfile(ctx, serviceDto.UpdatedUserInfo{
		Link:        parseUserLink,
		DisplayName: cleanDisplayName,
		Description: cleanDescription,
	})
	if err != nil {
		if errors.Is(err, common.ErrorMissingRequiredField) {
			return nil, status.Error(codes.InvalidArgument, msgFailNullValue)
		}
		if errors.Is(err, common.ErrorInvalidProfileData) {
			return nil, status.Error(codes.InvalidArgument, msgInvalidProfileData)
		}

		return nil, status.Error(codes.Internal, msgInternalError)
	}

	return &pb.UpdateProfileResponse{}, nil
}

func (h *Handler) UpdateAvatar(ctx context.Context, req *pb.UpdateAvatarRequest) (*pb.AvatarResponse, error) {
	logger := zerolog.Ctx(ctx)

	parseUserLink, err := uuid.Parse(req.UserLink)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, msgFailParseUserLink)
	}

	if len(req.FileData) == 0 {
		return nil, status.Error(codes.InvalidArgument, msgEmptyFile)
	}

	if int64(len(req.FileData)) > h.cfg.MaxReadBytes {
		logger.Error().
			Int("size", len(req.FileData)).
			Int64("max_size", h.cfg.MaxReadBytes).
			Msg("avatar file is too large")
		return nil, status.Error(codes.InvalidArgument, msgTooLargeAvatar)
	}

	signatureSize := h.cfg.SignatureTypeBytes
	if len(req.FileData) < signatureSize {
		signatureSize = len(req.FileData)
	}

	mimeType := http.DetectContentType(req.FileData[:signatureSize])

	if _, ok := h.cfg.ValidExtensions[mimeType]; !ok {
		logger.Error().Str("mime_type", mimeType).Msg("incorrect avatar type")
		return nil, status.Error(codes.InvalidArgument, msgIncorrectTypeAvatar)
	}

	avatarUrl, err := h.srv.UpdateAvatar(ctx, serviceDto.UpdatedAvatar{
		UserLink: parseUserLink,
		MimeType: mimeType,
		File:     bytes.NewReader(req.FileData),
	})
	if err != nil {
		if errors.Is(err, common.ErrorNonexistentUser) {
			return nil, status.Error(codes.NotFound, msgFailFoundUser)
		}

		logger.Error().Err(err).Msg("failed to update avatar in service")
		return nil, status.Error(codes.Internal, msgInternalError)
	}

	return &pb.AvatarResponse{
		AvatarUrl: avatarUrl,
	}, nil
}

func (h *Handler) DeleteAvatar(ctx context.Context, req *pb.UserLinkRequest) (*pb.DeleteAvatarResponse, error) {
	logger := zerolog.Ctx(ctx)

	parseUserLink, err := uuid.Parse(req.UserLink)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, msgFailParseUserLink)
	}

	err = h.srv.DeleteAvatar(ctx, parseUserLink)
	if err != nil {
		if errors.Is(err, common.ErrorNonexistentUser) {
			return nil, status.Error(codes.NotFound, msgFailFoundUser)
		}

		logger.Error().Err(err).Msg("DeleteAvatar failed")
		return nil, status.Error(codes.Internal, msgInternalError)
	}

	return &pb.DeleteAvatarResponse{}, nil
}
