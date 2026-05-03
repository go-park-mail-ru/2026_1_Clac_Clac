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
	rbac "github.com/go-park-mail-ru/2026_1_Clac_Clac/pkg/appealRbac"
	sentryLogger "github.com/go-park-mail-ru/2026_1_Clac_Clac/pkg/logger"
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
	ErrUnexpectedCategory     = errors.New("unexpected category")
	ErrUnexpectedStatus       = errors.New("unexpected status")
	ErrInvalidActions         = errors.New("this role can not do it")
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

func parseProtoCategory(pbCategory pb.Category) (common.Category, error) {
	switch pbCategory {
	case pb.Category_CATEGORY_BUG:
		return common.Categories.Bug, nil
	case pb.Category_CATEGORY_PROPOSAL:
		return common.Categories.Proposal, nil
	case pb.Category_CATEGORY_COMPLAINT:
		return common.Categories.Complaint, nil
	}

	return "", ErrUnexpectedCategory
}

func toProtoCategory(category common.Category) pb.Category {
	switch category {
	case common.Categories.Bug:
		return pb.Category_CATEGORY_BUG
	case common.Categories.Proposal:
		return pb.Category_CATEGORY_PROPOSAL
	case common.Categories.Complaint:
		return pb.Category_CATEGORY_COMPLAINT
	}

	return pb.Category_CATEGORY_UNSPECIFIED
}

func parseProtoStatus(pbStatus pb.Status) (common.Status, error) {
	switch pbStatus {
	case pb.Status_STATUS_OPEN:
		return common.Statuses.Open, nil
	case pb.Status_STATUS_IN_WORK:
		return common.Statuses.InWork, nil
	case pb.Status_STATUS_CLOSE:
		return common.Statuses.Close, nil
	}

	return "", ErrUnexpectedStatus
}

func toProtoStatus(status common.Status) pb.Status {
	switch status {
	case common.Statuses.Open:
		return pb.Status_STATUS_OPEN
	case common.Statuses.InWork:
		return pb.Status_STATUS_IN_WORK
	case common.Statuses.Close:
		return pb.Status_STATUS_CLOSE
	}

	return pb.Status_STATUS_UNSPECIFIED
}

func toProtoRole(role rbac.Role) pb.Role {
	switch role {
	case rbac.Roles.User:
		return pb.Role_ROLE_USER
	case rbac.Roles.Support:
		return pb.Role_ROLE_SUPPORT
	case rbac.Roles.Admin:
		return pb.Role_ROLE_ADMIN
	}

	return pb.Role_ROLE_UNSPECIFIED
}

func (h *Handler) CreateAppeal(ctx context.Context, req *pb.CreateAppealRequest) (*pb.CreateAppealResponse, error) {
	logger := zerolog.Ctx(ctx)

	rawUserLink := req.GetUserLink()
	userLink, err := uuid.Parse(rawUserLink)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, ErrInvalidUserLink.Error())
	}

	appealCategory, err := parseProtoCategory(req.GetCategory())
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, ErrUnexpectedCategory.Error())
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

		errLog := fmt.Errorf("srv.CreateAppeal: %w", err)
		logger.Error().Err(errLog).Msg("failed to create appeal")
		sentryLogger.CaptureFromContext(ctx, errLog, "CreateAppeal", map[string]interface{}{
			"user_link": rawUserLink,
			"category":  req.GetCategory().String(),
			"action":    "create_appeal",
		})
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
		errLog := fmt.Errorf("srv.GetUserAppeals: %w", err)
		logger.Error().Err(errLog).Msg("failed to get user appeals")
		sentryLogger.CaptureFromContext(ctx, errLog, "GetAppeals", map[string]interface{}{
			"user_link": rawUserLink,
			"action":    "get_appeals",
		})
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
			Category:      toProtoCategory(a.Category),
			Status:        toProtoStatus(a.Status),
			Description:   a.Description,
			AttachmentUrl: attachmentKey,
			CreatedAt:     timestamppb.New(a.CreatedAt),
		})
	}

	return &pb.GetAppealsResponse{
		Role:        toProtoRole(appeals.Role),
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

		errLog := fmt.Errorf("srv.UploadAttachment: %w", err)
		logger.Error().Err(errLog).Msg("failed to upload attachment")
		sentryLogger.CaptureFromContext(ctx, errLog, "UploadAttachment", map[string]interface{}{
			"user_link":   rawUserLink,
			"appeal_link": rawAppealLink,
			"action":      "upload_attachment",
		})
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
		errLog := fmt.Errorf("srv.DeleteAppeal: %w", err)
		logger.Error().Err(errLog).Msg("failed to delete appeal")
		sentryLogger.CaptureFromContext(ctx, errLog, "DeleteAppeal", map[string]interface{}{
			"user_link":   rawUserLink,
			"appeal_link": rawAppealLink,
			"action":      "delete_appeal",
		})
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

		errLog := fmt.Errorf("srv.GetStats: %w", err)
		logger.Error().Err(errLog).Msg("get stats")
		sentryLogger.CaptureFromContext(ctx, errLog, "GetStats", map[string]interface{}{
			"user_link": rawUserLink,
			"action":    "get_stats",
		})
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

	appealNewStatus, err := parseProtoStatus(req.GetNewStatus())
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, ErrUnexpectedStatus.Error())
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

		errLog := fmt.Errorf("srv.ChangeAppealStatus: %w", err)
		logger.Error().Err(errLog).Msg("change status")
		sentryLogger.CaptureFromContext(ctx, errLog, "ChangeAppealStatus", map[string]interface{}{
			"user_link":   rawUserLink,
			"appeal_link": rawAppealLink,
			"action":      "change_appeal_status",
		})
		return nil, status.Error(codes.Internal, ErrCannotChangeStatus.Error())
	}

	return &pb.ChangeAppealStatusResponse{}, nil
}
