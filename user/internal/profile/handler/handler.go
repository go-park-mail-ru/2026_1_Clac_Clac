package handler

import (
	"bytes"
	"context"
	"errors"
	"net/http"
	"strings"

	pb "github.com/go-park-mail-ru/2026_1_Clac_Clac/pkg/profile"
	"github.com/go-park-mail-ru/2026_1_Clac_Clac/user/internal/common"
	serviceDto "github.com/go-park-mail-ru/2026_1_Clac_Clac/user/internal/profile/service/dto"
	"github.com/google/uuid"
	"github.com/microcosm-cc/bluemonday"
	"github.com/rs/zerolog"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

const (
	msgFailParseUserLink   = "user link can not convert to uuid"
	msgTooLargeAvatar      = "size avatar too large"
	msgEmptyFile           = "empty file provided"
	msgIncorrectTypeAvatar = "avatar can be only jpeg/png/jpg/webp"
	msgFailFoundUser       = "user not found"
	msgFailNullValue       = "get null, but wait not null"
	msgInvalidProfileData  = "invalid profile data"
	msgInternalError       = "something went wrong"
	msgInvalidInput        = "invalid input parameters"
)

type ProfileService interface {
	GetProfileUser(ctx context.Context, userLink uuid.UUID) (serviceDto.UserInfo, error)
	GetProfileByLink(ctx context.Context, userLink uuid.UUID) (serviceDto.UserInfo, error)
	UpdateProfile(ctx context.Context, updatedInfo serviceDto.UpdatedUserInfo) error
	UpdateAvatar(ctx context.Context, avatar serviceDto.UpdatedAvatar) (string, error)
	DeleteAvatar(ctx context.Context, userLink uuid.UUID) error
}

type Config struct {
	ValidExtensions       map[string]struct{}
	SiganatureTypeBytes   int
	MaxLenNameUser        int
	MaxLenDescriptionUser int
	MaxReadBytes          int64
}

type Handler struct {
	srv       ProfileService
	cfg       Config
	sanitizer *bluemonday.Policy
	pb.UnimplementedProfileServiceServer
}

func NewHandler(srv ProfileService, cfg Config) *Handler {
	return &Handler{
		srv:       srv,
		cfg:       cfg,
		sanitizer: bluemonday.StrictPolicy(),
	}
}

func (h *Handler) GetProfile(ctx context.Context, req *pb.UserLinkRequest) (*pb.ProfileResponse, error) {
	parseUserLink, err := uuid.Parse(req.UserLink)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, msgFailParseUserLink)
	}

	serviceUser, err := h.srv.GetProfileUser(ctx, parseUserLink)
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

func (h *Handler) GetProfileByLink(ctx context.Context, req *pb.UserLinkRequest) (*pb.ProfileResponse, error) {
	logger := zerolog.Ctx(ctx)

	parseUserLink, err := uuid.Parse(req.UserLink)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, msgFailParseUserLink)
	}

	serviceUser, err := h.srv.GetProfileByLink(ctx, parseUserLink)
	if err != nil {
		if errors.Is(err, common.ErrorNonexistentUser) {
			return nil, status.Error(codes.NotFound, msgFailFoundUser)
		}

		logger.Error().Err(err).Msg("ProfileService.GetProfileByLink failed")
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

	err = common.ValidateTextInfo(cleanDisplayName, h.cfg.MaxLenNameUser)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, msgInvalidInput)
	}

	err = common.ValidateTextInfo(cleanDescription, h.cfg.MaxLenDescriptionUser)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, msgInvalidInput)
	}

	userInfo := serviceDto.UpdatedUserInfo{
		Link:        parseUserLink,
		DisplayName: cleanDisplayName,
		Description: cleanDescription,
	}

	err = h.srv.UpdateProfile(ctx, userInfo)
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

	signatureSize := h.cfg.SiganatureTypeBytes
	if len(req.FileData) < signatureSize {
		signatureSize = len(req.FileData)
	}

	mimeType := http.DetectContentType(req.FileData[:signatureSize])

	if _, ok := h.cfg.ValidExtensions[mimeType]; !ok {
		logger.Error().Str("mime_type", mimeType).Msg("incorrect avatar type")
		return nil, status.Error(codes.InvalidArgument, msgIncorrectTypeAvatar)
	}

	fileReader := bytes.NewReader(req.FileData)

	avatarUrl, err := h.srv.UpdateAvatar(ctx, serviceDto.UpdatedAvatar{
		UserLink: parseUserLink,
		MimeType: mimeType,
		File:     fileReader,
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
