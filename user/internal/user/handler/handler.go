package handler

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
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

	msgFailParseUserLink  = "user link can not convert to uuid"
	msgEmptyFile          = "empty file provided"
	msgFailFoundUser      = "user not found"
	msgFailNullValue      = "get null, but wait not null"
	msgInvalidProfileData = "invalid profile data"
	msgInvalidMetadata    = "invalid metadata for avatar"
	msgCannotCreateFile   = "cannot create temp file for download avatar"
	msgWriteInFile        = "cannot write chunk to file"
	msgCursorOffset       = "cannot set cursor offset"
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
}

type Config struct {
	APIMethod string
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

	res, err := h.httpClient.Get(fmt.Sprintf(h.cfg.APIMethod, req.AccessToken))
	if err != nil {
		logger.Err(err).Msg("vk api request failed")

		sentryLogger.CaptureFromContext(ctx, err, "ProcessUserWithVK", map[string]interface{}{
			"email":  req.Email,
			"action": "vk_api_request",
		})

		return nil, status.Error(codes.Unavailable, msgOAuthCannotRequestUserData)
	}
	defer func() {
		if err := res.Body.Close(); err != nil {
			logger.Err(err).Msg("close vk api response body")
		}
	}()

	usersData := &api.VkAPIUsersData{}
	if err := json.NewDecoder(res.Body).Decode(usersData); err != nil {
		logger.Err(err).Msg("vk api decode response")

		sentryLogger.CaptureFromContext(ctx, err, "ProcessUserWithVK", map[string]interface{}{
			"email":  req.Email,
			"action": "vk_api_decode",
		})

		return nil, status.Error(codes.Internal, msgInternalError)
	}

	if len(usersData.Response) < 1 {
		logger.Error().Msg("vk api: empty user data")

		errEmptyData := errors.New("vk api returned empty user data")
		sentryLogger.CaptureFromContext(ctx, errEmptyData, "ProcessUserWithVK", map[string]interface{}{
			"email":  req.Email,
			"action": "vk_api_empty_data",
		})

		return nil, status.Error(codes.Internal, msgOAuthEmptyUserData)
	}

	userLink, err := h.srv.EnsureUserByEmail(ctx, serviceDto.EntityUser{
		DisplayName: usersData.Response[0].FirstName,
		Email:       req.Email,
	})
	if err != nil {
		logger.Err(err).Msg("EnsureUserByEmail failed")

		sentryLogger.CaptureFromContext(ctx, err, "ProcessUserWithVK", map[string]interface{}{
			"email":  req.Email,
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
