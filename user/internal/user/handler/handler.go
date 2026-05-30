package handler

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"

	sentryLogger "github.com/go-park-mail-ru/2026_1_Clac_Clac/pkg/logger"
	pb "github.com/go-park-mail-ru/2026_1_Clac_Clac/pkg/proto/user/v1"
	"github.com/go-park-mail-ru/2026_1_Clac_Clac/user/internal/api"
	"github.com/go-park-mail-ru/2026_1_Clac_Clac/user/internal/common"
	"github.com/go-park-mail-ru/2026_1_Clac_Clac/user/internal/user/service"
	serviceDto "github.com/go-park-mail-ru/2026_1_Clac_Clac/user/internal/user/service/dto"
	"github.com/google/uuid"
	"github.com/rs/zerolog"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

const (
	msgInternalError = "something went wrong"

	msgWrongEmailOrPassword = "wrong email or password"
	msgUserDoesNotExists    = "user does not exist"
	msgEmailDoesNotExists   = "user with this email does not exist"
	msgUserAlreadyExists    = "user already exists"
	msgNullInNotNullField   = "put null value in not null field"

	msgOAuthCannotRequestUserData = "oauth_cannot_request_user_data"
	msgOAuthEmptyUserData         = "oauth_no_user_data"
	msgOAuthInvalidParams         = "oauth_invalid_params"

	msgFailParseUserLink  = "user link can not convert to uuid"
	msgEmptyFile          = "empty file provided"
	msgFailFoundUser      = "user not found"
	msgFailNullValue      = "get null, but wait not null"
	msgInvalidProfileData = "invalid profile data"
	msgInvalidMetadata    = "invalid metadata for avatar"
	msgCannotCreateFile   = "cannot create temp file for download avatar"
	msgWriteInFile        = "cannot write chunk to file"
	msgCursorOffset       = "cannot set cursor offset"

	vkIDAuthEndpoint     = "https://id.vk.ru/oauth2/auth"
	vkIDUserInfoEndpoint = "https://id.vk.ru/oauth2/user_info"
)

type AuthService interface {
	CreateUser(ctx context.Context, requestUser serviceDto.EntityUser) (serviceDto.UserInfo, error)
	GetUser(ctx context.Context, requestUser serviceDto.GetUserInfo) (serviceDto.UserInfo, error)
	GetUserLink(ctx context.Context, email string) (string, error)
	EnsureUserByEmail(ctx context.Context, info serviceDto.EntityUser) (string, error)
	ResetPassword(ctx context.Context, passwordInfo serviceDto.ResetPasswordInfo) error

	GetProfile(ctx context.Context, userLink uuid.UUID) (serviceDto.UserInfo, error)
	GetProfiles(ctx context.Context, userLinks []uuid.UUID) ([]serviceDto.UserInfo, error)
	UpdateProfile(ctx context.Context, updatedInfo serviceDto.UpdatedUserInfo) error
	UpdateAvatar(ctx context.Context, avatar serviceDto.UpdatedAvatar) (string, error)
	DeleteAvatar(ctx context.Context, userLink uuid.UUID) error
}

type HTTPClient interface {
	Get(url string) (*http.Response, error)
	PostForm(url string, data url.Values) (*http.Response, error)
}

type Config struct {
	ClientID     string
	ClientSecret string
	RedirectURI  string
}

type Handler struct {
	srv        AuthService
	cfg        Config
	httpClient HTTPClient
	pb.UnimplementedUserServiceServer
}

func NewHandler(srv AuthService, cfg Config, httpClient HTTPClient) *Handler {
	return &Handler{
		srv:        srv,
		cfg:        cfg,
		httpClient: httpClient,
	}
}

func (h *Handler) GetUser(ctx context.Context, req *pb.GetUserRequest) (*pb.UserResponse, error) {
	logger := zerolog.Ctx(ctx)
	serviceUser, err := h.srv.GetUser(ctx, serviceDto.GetUserInfo{
		Email:    req.Email,
		Password: req.Password,
	})

	if err != nil {
		if errors.Is(err, service.ErrorWrongPassword) {
			return nil, status.Error(codes.InvalidArgument, msgWrongEmailOrPassword)
		}

		if errors.Is(err, common.ErrorNonexistentEmail) {
			return nil, status.Error(codes.NotFound, msgUserDoesNotExists)
		}

		errLog := fmt.Errorf("GetUser: %w", err)
		logger.Err(errLog).Msg("login user")

		sentryLogger.CaptureFromContext(ctx, errLog, "GetUser", map[string]interface{}{
			"email":  req.Email,
			"action": "get_user",
		})

		return nil, status.Error(codes.Internal, msgInternalError)
	}

	return &pb.UserResponse{
		UserLink:    serviceUser.Link.String(),
		DisplayName: serviceUser.DisplayName,
		Email:       serviceUser.Email,
		Avatar:      serviceUser.AvatarURL,
	}, nil
}

func (h *Handler) CreateUser(ctx context.Context, req *pb.CreateRequest) (*pb.UserResponse, error) {
	logger := zerolog.Ctx(ctx)

	serviceUser, err := h.srv.CreateUser(ctx, serviceDto.EntityUser{
		DisplayName: req.DisplayName,
		Email:       req.Email,
		Password:    req.Password,
	})
	if err != nil {
		errLog := fmt.Errorf("CreateUser: %w", err)
		logger.Err(errLog).Msg("register user")

		if errors.Is(err, common.ErrorExistingUser) {
			return nil, status.Error(codes.AlreadyExists, msgUserAlreadyExists)
		}
		if errors.Is(err, common.ErrorNotNullValue) {
			return nil, status.Error(codes.InvalidArgument, msgNullInNotNullField)
		}

		sentryLogger.CaptureFromContext(ctx, errLog, "UserService.CreateUser", map[string]interface{}{
			"email":  req.Email,
			"action": "create_user",
		})

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
	logger := zerolog.Ctx(ctx)

	userLink, err := h.srv.GetUserLink(ctx, req.Email)
	if err != nil {
		if errors.Is(err, common.ErrorNonexistentEmail) {
			return nil, status.Error(codes.NotFound, msgEmailDoesNotExists)
		}

		errLog := fmt.Errorf("UserService.GetUserLink: %w", err)
		logger.Error().Err(errLog).Msg("UserService.GetUserLink failed")

		sentryLogger.CaptureFromContext(ctx, errLog, "GetUserLink", map[string]interface{}{
			"email":  req.Email,
			"action": "get_user_link",
		})

		return nil, status.Error(codes.Internal, msgInternalError)
	}

	return &pb.GetUserLinkResponse{UserLink: userLink}, nil
}

func (h *Handler) ResetPassword(ctx context.Context, req *pb.ResetPasswordRequest) (*pb.ResetPasswordResponse, error) {
	logger := zerolog.Ctx(ctx)

	err := h.srv.ResetPassword(ctx, serviceDto.ResetPasswordInfo{
		UserLink:    req.UserLink,
		NewPassword: req.Password,
	})
	if err != nil {
		errLog := fmt.Errorf("UserService.ResetPassword: %w", err)
		logger.Error().Err(errLog).Msg("ResetPassword failed")

		if errors.Is(err, common.ErrorNotNullValue) {
			return nil, status.Error(codes.InvalidArgument, msgNullInNotNullField)
		}
		if errors.Is(err, common.ErrorNonexistentUser) {
			return nil, status.Error(codes.NotFound, msgUserDoesNotExists)
		}

		sentryLogger.CaptureFromContext(ctx, errLog, "ResetPassword", map[string]interface{}{
			"user_link": req.UserLink,
			"action":    "reset_password",
		})

		return nil, status.Error(codes.Internal, msgInternalError)
	}

	return &pb.ResetPasswordResponse{}, nil
}

func (h *Handler) ProcessUserWithVK(ctx context.Context, req *pb.ProcessUserVKRequest) (*pb.ProcessUserVKResponse, error) {
	logger := zerolog.Ctx(ctx)

	if req.Code == "" || req.CodeVerifier == "" || req.State == "" {
		return nil, status.Error(codes.InvalidArgument, msgOAuthInvalidParams)
	}

	tokenData := url.Values{
		"grant_type":    {"authorization_code"},
		"client_id":     {h.cfg.ClientID},
		"code":          {req.Code},
		"code_verifier": {req.CodeVerifier},
		"state":         {req.State},
		"redirect_uri":  {h.cfg.RedirectURI},
	}
	if req.DeviceId != "" {
		tokenData.Set("device_id", req.DeviceId)
	}

	tokenResp, err := h.httpClient.PostForm(vkIDAuthEndpoint, tokenData)
	if err != nil {
		logger.Err(err).Msg("vk id auth request failed")

		sentryLogger.CaptureFromContext(ctx, err, "ProcessUserWithVK", map[string]interface{}{
			"action": "vk_id_auth_request",
		})

		return nil, status.Error(codes.Unavailable, msgOAuthCannotRequestUserData)
	}
	defer func() {
		if err := tokenResp.Body.Close(); err != nil {
			logger.Err(err).Msg("close vk id auth response body")
		}
	}()

	var token api.VkIDTokenResponse
	if err := json.NewDecoder(tokenResp.Body).Decode(&token); err != nil {
		logger.Err(err).Msg("vk id auth decode response")

		sentryLogger.CaptureFromContext(ctx, err, "ProcessUserWithVK", map[string]interface{}{
			"action": "vk_id_auth_decode",
		})

		return nil, status.Error(codes.Internal, msgInternalError)
	}

	if token.Error != "" {
		logger.Error().
			Int("vk_status_code", tokenResp.StatusCode).
			Str("vk_error", token.Error).
			Str("vk_error_description", token.ErrorDescription).
			Msg("vk id auth: vk returned error")

		sentryLogger.CaptureFromContext(ctx, fmt.Errorf("vk oauth error: %s — %s", token.Error, token.ErrorDescription), "ProcessUserWithVK", map[string]interface{}{
			"action":      "vk_id_error_response",
			"vk_error":    token.Error,
			"status_code": tokenResp.StatusCode,
		})

		return nil, status.Error(codes.Unavailable, msgOAuthCannotRequestUserData)
	}

	if token.AccessToken == "" {
		logger.Error().Msg("vk id auth: empty access token")

		sentryLogger.CaptureFromContext(ctx, errors.New("vk id returned empty access token"), "ProcessUserWithVK", map[string]interface{}{
			"action": "vk_id_empty_token",
		})

		return nil, status.Error(codes.Unavailable, msgOAuthCannotRequestUserData)
	}

	userData := url.Values{
		"client_id":    {h.cfg.ClientID},
		"access_token": {token.AccessToken},
	}

	userResp, err := h.httpClient.PostForm(vkIDUserInfoEndpoint, userData)
	if err != nil {
		logger.Err(err).Msg("vk id user info request failed")

		sentryLogger.CaptureFromContext(ctx, err, "ProcessUserWithVK", map[string]interface{}{
			"action": "vk_id_user_info_request",
		})

		return nil, status.Error(codes.Unavailable, msgOAuthCannotRequestUserData)
	}
	defer func() {
		if err := userResp.Body.Close(); err != nil {
			logger.Err(err).Msg("close vk id user info response body")
		}
	}()

	var userInfo api.VkIDUserInfoResponse
	if err := json.NewDecoder(userResp.Body).Decode(&userInfo); err != nil {
		logger.Err(err).Msg("vk id user info decode response")

		sentryLogger.CaptureFromContext(ctx, err, "ProcessUserWithVK", map[string]interface{}{
			"action": "vk_id_user_info_decode",
		})

		return nil, status.Error(codes.Internal, msgInternalError)
	}

	userInfo.User.Email = fmt.Sprintf("%s@nexus.internal", userInfo.User.UserID)

	if userInfo.User.FirstName == "" {
		logger.Error().Msg("vk id: empty user data")

		sentryLogger.CaptureFromContext(ctx, errors.New("vk id returned empty user data"), "ProcessUserWithVK", map[string]interface{}{
			"action": "vk_id_empty_user_data",
		})

		return nil, status.Error(codes.Internal, msgOAuthEmptyUserData)
	}

	userLink, err := h.srv.EnsureUserByEmail(ctx, serviceDto.EntityUser{
		DisplayName: userInfo.User.FirstName,
		Email:       userInfo.User.Email,
	})
	if err != nil {
		logger.Err(err).Msg("EnsureUserByEmail failed")

		sentryLogger.CaptureFromContext(ctx, err, "ProcessUserWithVK", map[string]interface{}{
			"email":  userInfo.User.Email,
			"action": "ensure_user_by_email",
		})

		return nil, status.Error(codes.Internal, msgInternalError)
	}

	return &pb.ProcessUserVKResponse{UserLink: userLink}, nil
}

func (h *Handler) GetProfile(ctx context.Context, req *pb.UserLinkRequest) (*pb.ProfileResponse, error) {
	logger := zerolog.Ctx(ctx)

	parseUserLink, err := uuid.Parse(req.UserLink)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, msgFailParseUserLink)
	}

	serviceUser, err := h.srv.GetProfile(ctx, parseUserLink)
	if err != nil {
		if errors.Is(err, common.ErrorNonexistentUser) {
			return nil, status.Error(codes.NotFound, msgFailFoundUser)
		}
		logger.Error().Err(err).Msg("srv.GetProfile failed")

		sentryLogger.CaptureFromContext(ctx, err, "GetProfile", map[string]interface{}{
			"user_link": req.UserLink,
			"action":    "get_profile",
		})

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

func (h *Handler) GetProfiles(ctx context.Context, req *pb.GetProfilesRequest) (*pb.GetProfilesResponse, error) {
	logger := zerolog.Ctx(ctx)

	rawLinks := req.GetUserLinks()
	links := make([]uuid.UUID, 0, len(rawLinks))
	for _, rawLink := range rawLinks {
		parsed, err := uuid.Parse(rawLink)
		if err != nil {
			return nil, status.Error(codes.InvalidArgument, msgFailParseUserLink)
		}
		links = append(links, parsed)
	}

	serviceUsers, err := h.srv.GetProfiles(ctx, links)
	if err != nil {
		logger.Error().Err(err).Msg("srv.GetProfiles failed")

		sentryLogger.CaptureFromContext(ctx, err, "GetProfiles", map[string]interface{}{
			"action": "get_profiles",
		})

		return nil, status.Error(codes.Internal, msgInternalError)
	}

	profiles := make([]*pb.ProfileResponse, 0, len(serviceUsers))
	for _, su := range serviceUsers {
		profiles = append(profiles, &pb.ProfileResponse{
			UserLink:    su.Link.String(),
			Email:       su.Email,
			DisplayName: su.DisplayName,
			Description: su.Description,
			AvatarUrl:   su.AvatarURL,
		})
	}

	return &pb.GetProfilesResponse{
		Profiles: profiles,
	}, nil
}

func (h *Handler) UpdateProfile(ctx context.Context, req *pb.UpdateProfileRequest) (*pb.UpdateProfileResponse, error) {
	logger := zerolog.Ctx(ctx)

	parseUserLink, err := uuid.Parse(req.UserLink)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, msgFailParseUserLink)
	}

	err = h.srv.UpdateProfile(ctx, serviceDto.UpdatedUserInfo{
		Link:        parseUserLink,
		DisplayName: req.DisplayName,
		Description: req.Description,
	})
	if err != nil {
		if errors.Is(err, common.ErrorMissingRequiredField) {
			return nil, status.Error(codes.InvalidArgument, msgFailNullValue)
		}
		if errors.Is(err, common.ErrorInvalidProfileData) {
			return nil, status.Error(codes.InvalidArgument, msgInvalidProfileData)
		}
		logger.Error().Err(err).Msg("srv.UpdateProfile failed")

		sentryLogger.CaptureFromContext(ctx, err, "UpdateProfile", map[string]interface{}{
			"user_link": req.UserLink,
			"action":    "update_profile",
		})

		return nil, status.Error(codes.Internal, msgInternalError)
	}

	return &pb.UpdateProfileResponse{}, nil
}

func (h *Handler) UpdateAvatar(stream grpc.ClientStreamingServer[pb.UpdateAvatarRequest, pb.AvatarResponse]) error {
	ctx := stream.Context()
	logger := zerolog.Ctx(ctx)

	req, err := stream.Recv()
	if err != nil {
		return status.Error(codes.Aborted, "connection is failed")
	}

	metadata := req.GetMetadata()
	if metadata == nil {
		return status.Error(codes.InvalidArgument, msgInvalidMetadata)
	}

	parseUserLink, err := uuid.Parse(metadata.UserLink)
	if err != nil {
		return status.Error(codes.InvalidArgument, msgFailParseUserLink)
	}

	uniqueFileName := fmt.Sprintf("%s_avatar", uuid.New().String())
	tempFilePath := filepath.Join(os.TempDir(), uniqueFileName)

	file, err := os.Create(tempFilePath)
	if err != nil {
		return status.Error(codes.Internal, msgCannotCreateFile)
	}

	defer func() {
		_ = file.Close()
		_ = os.Remove(tempFilePath)
	}()

	for {
		req, err := stream.Recv()
		if err == io.EOF {
			break
		}
		if err != nil {
			return status.Error(codes.Aborted, "connection is failed")
		}

		chunk := req.GetFileData()
		if len(chunk) == 0 && err != io.EOF {
			continue
		}
		if _, err := file.Write(chunk); err != nil {
			return status.Error(codes.Internal, msgWriteInFile)
		}
	}

	if _, err := file.Seek(0, io.SeekStart); err != nil {
		return status.Error(codes.Internal, msgCursorOffset)
	}

	avatarUrl, err := h.srv.UpdateAvatar(ctx, serviceDto.UpdatedAvatar{
		UserLink: parseUserLink,
		MimeType: metadata.ContentType,
		File:     file,
	})
	if err != nil {
		if errors.Is(err, common.ErrorNonexistentUser) {
			return status.Error(codes.NotFound, msgFailFoundUser)
		}
		logger.Error().Err(err).Msg("UpdateAvatar failed")

		sentryLogger.CaptureFromContext(ctx, err, "UpdateAvatar", map[string]interface{}{
			"user_link": metadata.UserLink,
			"mime_type": metadata.ContentType,
			"action":    "update_avatar",
		})

		return status.Error(codes.Internal, msgInternalError)
	}

	err = stream.SendAndClose(&pb.AvatarResponse{AvatarUrl: avatarUrl})
	if err != nil {
		return status.Error(codes.Internal, msgInternalError)
	}

	return nil
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

		sentryLogger.CaptureFromContext(ctx, err, "DeleteAvatar", map[string]interface{}{
			"user_link": req.UserLink,
			"action":    "delete_avatar",
		})

		return nil, status.Error(codes.Internal, msgInternalError)
	}

	return &pb.DeleteAvatarResponse{}, nil
}
