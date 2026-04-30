package delivery

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"path/filepath"
	"strings"

	serviceDto "github.com/go-park-mail-ru/2026_1_Clac_Clac/appeal/internal/service/dto"
	rbac "github.com/go-park-mail-ru/2026_1_Clac_Clac/pkg/boardRbac"
	pb "github.com/go-park-mail-ru/2026_1_Clac_Clac/pkg/proto/appeal/v1"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"

	"github.com/go-park-mail-ru/2026_1_Clac_Clac/appeal/internal/common"
	"github.com/go-park-mail-ru/2026_1_Clac_Clac/appeal/internal/delivery/dto"
	"github.com/google/uuid"
	"github.com/rs/zerolog"
)

var (
	ErrInvalidActions = errors.New("this role can not do it")

	msgInternalError          = "server error internal"
	ErrInvalidRequestSchema   = errors.New("invalid schema")
	ErrInvalidEmailOrName     = errors.New("incorrect email or name")
	ErrParseMultipartForm     = errors.New("file too large or invalid form")
	ErrCannotFindAttachment   = errors.New("cannot find 'attachment' key")
	ErrCannotReadFile         = errors.New("cannot read file")
	ErrInvalidContentType     = errors.New("invalid content type")
	ErrCannotOperateWithFile  = errors.New("cannot operate with file")
	ErrCannotUploadFile       = errors.New("cannot upload attachment")
	ErrInvalidUserLink        = errors.New("invalid user link")
	ErrCannotCreateAppeal     = errors.New("cannot create appeal")
	ErrCannotGetAppeals       = errors.New("cannot get appeals")
	ErrInvalidAppealLink      = errors.New("invalid appeal link")
	ErrCannotUploadAttachment = errors.New("cannot upload attachment")
	ErrCannotDeleteAppeal     = errors.New("cannot delete appeal")
	ErrCannotGetStats         = errors.New("cannot get stats")
	ErrCannotChangeStatus     = errors.New("cannot change status")
)

//go:generate mockery --name=AppealService --output mock_appeal_srv
type AppealService interface {
	CreateAppeal(ctx context.Context, appeal serviceDto.EntityAppeal) (uuid.UUID, error)
	GetAppeals(ctx context.Context, userLink uuid.UUID) (serviceDto.Appeals, error)
	DeleteAppeal(ctx context.Context, appealLink uuid.UUID, userLink uuid.UUID) error
	GetStats(ctx context.Context, userLink uuid.UUID) (serviceDto.AppealStats, error)
	ChangeAppealStatus(ctx context.Context, info serviceDto.ChangeAppealStatusInfo) error
	UploadAttachment(ctx context.Context, file io.Reader, contentType, extension string, appealLink uuid.UUID, userLink uuid.UUID) (string, error)
}

type Config struct {
	AttachmentBaseURL string
}

type Handler struct {
	pb.UnimplementedAppealServiceServer

	srv  AppealService
	conf Config
}

func NewHandler(srv AppealService, conf Config) *Handler {
	return &Handler{
		srv:  srv,
		conf: conf,
	}
}

func (h *Handler) CreateAppeal(ctx context.Context, req *pb.CreateAppealRequest) (*pb.CreateAppealResponse, error) {
	logger := zerolog.Ctx(ctx)

	rawUserLink := req.GetUserLink()
	userLink, err := uuid.Parse(rawUserLink)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, ErrInvalidUserLink.Error())
	}

	appealCategory, err := common.ParseProtoCategory(req.GetCategory())
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, common.ErrUnexpectedCategory.Error())
	}

	request := dto.EntityAppealRequest{
		Mail:        req.GetEmail(),
		Category:    appealCategory,
		Description: req.GetDescription(),
		DisplayName: req.GetDisplayName(),
	}
	request.Sanitize()

	if err := ValidatorRequestAppeal(request.Mail, request.DisplayName); err != nil {
		return nil, status.Error(codes.InvalidArgument, ErrInvalidEmailOrName.Error())
	}

	appealLink, err := h.srv.CreateAppeal(ctx, serviceDto.EntityAppeal{
		UserLink:    userLink,
		DisplayName: request.DisplayName,
		Mail:        request.Mail,
		Description: request.Description,
		Category:    request.Category,
	})
	if err != nil {
		switch {
		case errors.Is(err, common.ErrorExistingUser):
			return nil, status.Error(codes.InvalidArgument, common.ErrorExistingUser.Error())
		case errors.Is(err, common.ErrorNotNullValue):
			return nil, status.Error(codes.InvalidArgument, common.ErrorNotNullValue.Error())
		case errors.Is(err, common.ErrInvalidCategory):
			return nil, status.Error(codes.InvalidArgument, common.ErrInvalidCategory.Error())
		}

		logger.Error().Err(fmt.Errorf("srv.CreateAppeal: %w", err)).Msg("failed to create appeal")
		return nil, status.Error(codes.Internal, ErrCannotCreateAppeal.Error())
	}

	return &pb.CreateAppealResponse{
		AppealLink: appealLink.String(),
	}, nil
}

func (h *Handler) GetAppeals(ctx context.Context, req *pb.GetAppealsRequest) (*pb.GetAppealsResponse, error) {
	logger := zerolog.Ctx(ctx)

	rawUserLink := req.GetUserLink()
	userLink, err := uuid.Parse(rawUserLink)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, ErrInvalidUserLink.Error())
	}

	appeals, err := h.srv.GetAppeals(ctx, userLink)
	if err != nil {
		logger.Error().Err(fmt.Errorf("srv.GetUserAppeals: %w", err)).Msg("failed to get user appeals")
		return nil, status.Error(codes.Internal, ErrCannotGetAppeals.Error())
	}

	responseAppeals := make([]*pb.AppealInfo, 0, len(appeals.Appeals))
	for _, a := range appeals.Appeals {
		attachmentKey := a.AttachmentKey
		if attachmentKey != "" {
			attachmentKey = fmt.Sprintf("%s/%s", h.conf.AttachmentBaseURL, attachmentKey)
		}

		responseAppeals = append(responseAppeals, &pb.AppealInfo{
			AppealId:      int64(a.AppealID),
			AppealLink:    a.AppealLink.String(),
			Email:         a.Email,
			DisplayName:   a.DisplayName,
			Category:      common.ToProtoCategory(a.Category),
			Status:        common.ToProtoStatus(a.Status),
			Description:   a.Description,
			AttachmentUrl: attachmentKey,
			CreatedAt:     timestamppb.New(a.CreatedAt),
		})
	}

	return &pb.GetAppealsResponse{
		Role:        common.ToProtoRole(appeals.Role),
		AppealsInfo: responseAppeals,
	}, nil
}

func (h *Handler) UploadAttachment(ctx context.Context, req *pb.UploadAttachmentRequest) (*pb.UploadAttachmentResponse, error) {
	logger := zerolog.Ctx(ctx)

	rawUserLink := req.GetUserLink()
	userLink, err := uuid.Parse(rawUserLink)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, ErrInvalidUserLink.Error())
	}

	rawAppealLink := req.GetAppealLink()
	appealLink, err := uuid.Parse(rawAppealLink)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, ErrInvalidAppealLink.Error())
	}

	image := req.GetImage()

	contentType := http.DetectContentType(image)
	if !strings.HasPrefix(contentType, "image/") {
		return nil, status.Error(codes.InvalidArgument, ErrInvalidContentType.Error())
	}

	filename := req.GetFilename()
	extension := filepath.Ext(filename)

	attachmentKey, err := h.srv.UploadAttachment(ctx, bytes.NewReader(image), contentType, extension, appealLink, userLink)
	if err != nil {
		if errors.Is(err, common.ErrorAppealNotFound) {
			return nil, status.Error(codes.NotFound, common.ErrorAppealNotFound.Error())
		}

		logger.Error().Err(fmt.Errorf("srv.UploadAttachment: %w", err)).Msg("failed to upload attachment")
		return nil, status.Error(codes.Internal, ErrCannotUploadAttachment.Error())
	}

	return &pb.UploadAttachmentResponse{
		AttachmentUrl: fmt.Sprintf("%s/%s", h.conf.AttachmentBaseURL, attachmentKey),
	}, nil
}

func (h *Handler) DeleteAppeal(ctx context.Context, req *pb.DeleteAppealRequest) (*pb.DeleteAppealResponse, error) {
	logger := zerolog.Ctx(ctx)

	rawUserLink := req.GetUserLink()
	userLink, err := uuid.Parse(rawUserLink)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, ErrInvalidUserLink.Error())
	}

	rawAppealLink := req.GetAppealLink()
	appealLink, err := uuid.Parse(rawAppealLink)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, ErrInvalidAppealLink.Error())
	}

	err = h.srv.DeleteAppeal(ctx, appealLink, userLink)
	if err != nil {
		logger.Error().Err(fmt.Errorf("srv.DeleteAppeal: %w", err)).Msg("failed to delete appeal")
		return nil, status.Error(codes.Internal, ErrCannotDeleteAppeal.Error())
	}

	return &pb.DeleteAppealResponse{}, nil
}

func (h *Handler) GetStats(ctx context.Context, req *pb.GetStatsRequest) (*pb.GetStatsResponse, error) {
	logger := zerolog.Ctx(ctx)

	rawUserLink := req.GetUserLink()
	userLink, err := uuid.Parse(rawUserLink)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, ErrInvalidUserLink.Error())
	}

	stats, err := h.srv.GetStats(ctx, userLink)
	if err != nil {
		if errors.Is(err, common.ErrorPermissionDenied) {
			return nil, status.Error(codes.PermissionDenied, rbac.ErrActionDenied.Error())
		}

		logger.Error().Err(fmt.Errorf("srv.GetStats: %w", err)).Msg("get stats")
		return nil, status.Error(codes.Internal, ErrCannotGetStats.Error())
	}

	return &pb.GetStatsResponse{
		OpenAppeals:   int64(stats.Open),
		InWorkAppeals: int64(stats.InWork),
		CloseAppeals:  int64(stats.Close),
	}, nil
}

func (h *Handler) ChangeAppealStatus(ctx context.Context, req *pb.ChangeAppealStatusRequest) (*pb.ChangeAppealStatusResponse, error) {
	logger := zerolog.Ctx(ctx)

	rawUserLink := req.GetUserLink()
	userLink, err := uuid.Parse(rawUserLink)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, ErrInvalidUserLink.Error())
	}

	rawAppealLink := req.GetAppealLink()
	appealLink, err := uuid.Parse(rawAppealLink)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, ErrInvalidAppealLink.Error())
	}

	appealNewStatus, err := common.ParseProtoStatus(req.GetNewStatus())
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, common.ErrUnexpectedStatus.Error())
	}

	err = h.srv.ChangeAppealStatus(ctx, serviceDto.ChangeAppealStatusInfo{
		SupporterLink: userLink,
		AppealLink:    appealLink,
		Status:        appealNewStatus,
	})
	if err != nil {
		if errors.Is(err, common.ErrorPermissionDenied) {
			return nil, status.Error(codes.PermissionDenied, rbac.ErrActionDenied.Error())
		}

		logger.Error().Err(fmt.Errorf("srv.ChangeAppealStatus: %w", err)).Msg("change status")
		return nil, status.Error(codes.Internal, ErrCannotChangeStatus.Error())
	}

	return &pb.ChangeAppealStatusResponse{}, nil
}
