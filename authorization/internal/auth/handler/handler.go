package handler

import (
	"context"
	"errors"
	"fmt"

	"github.com/go-park-mail-ru/2026_1_Clac_Clac/authorization/internal/common"
	sentryLogger "github.com/go-park-mail-ru/2026_1_Clac_Clac/pkg/logger"
	pb "github.com/go-park-mail-ru/2026_1_Clac_Clac/pkg/proto/auth/v1"
	"github.com/google/uuid"
	"github.com/rs/zerolog"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

const (
	msgInternalError = "something went wrong"

	msgErrorParseUserLink  = "can not parse to uuid user link"
	msgDoesNotExistSession = "session does not exist"
)

type AuthService interface {
	CreateSession(ctx context.Context, userLink uuid.UUID) (string, error)
	GetUserLink(ctx context.Context, sessionID string) (string, error)
	DeleteSession(ctx context.Context, sessionID string) error
	ExtendSession(ctx context.Context, sessionID string) error
}

type Handler struct {
	srv AuthService
	pb.UnimplementedAuthServiceServer
}

func NewHandler(srv AuthService) *Handler {
	return &Handler{
		srv: srv,
	}
}

func (h *Handler) CreateSession(ctx context.Context, req *pb.CreateSessionRequest) (*pb.CreateSessionResponse, error) {
	logger := zerolog.Ctx(ctx)

	convertedLink, err := uuid.Parse(req.UserLink)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, msgErrorParseUserLink)
	}

	sessionID, err := h.srv.CreateSession(ctx, convertedLink)
	if err != nil {
		errLog := fmt.Errorf("srv.CreateSession: %w", err)
		logger.Error().Err(errLog).Msg("srv.CreateSession failed")
		sentryLogger.CaptureFromContext(ctx, errLog, "CreateSession", map[string]interface{}{
			"user_link": req.UserLink,
			"action":    "create_session",
		})
		return nil, status.Error(codes.Internal, msgInternalError)
	}

	return &pb.CreateSessionResponse{
		SessionId: sessionID,
	}, nil
}

func (h *Handler) GetUserLink(ctx context.Context, req *pb.GetUserLinkRequest) (*pb.GetUserLinkResponse, error) {
	logger := zerolog.Ctx(ctx)

	userLink, err := h.srv.GetUserLink(ctx, req.SessionId)
	if err != nil {
		if errors.Is(err, common.ErrorNotExistingSession) {
			return nil, status.Error(codes.NotFound, msgDoesNotExistSession)
		}

		errLog := fmt.Errorf("srv.GetUserLink: %w", err)
		logger.Error().Err(errLog).Msg("srv.GetUserLink failed")
		sentryLogger.CaptureFromContext(ctx, errLog, "GetUserLink", map[string]interface{}{
			"action": "get_user_link",
		})
		return nil, status.Error(codes.Internal, msgInternalError)
	}

	return &pb.GetUserLinkResponse{
		UserLink: userLink,
	}, nil
}

func (h *Handler) DeleteSession(ctx context.Context, req *pb.DeleteSessionRequest) (*pb.DeleteSessionResponse, error) {
	logger := zerolog.Ctx(ctx)

	err := h.srv.DeleteSession(ctx, req.SessionId)
	if err != nil {
		errLog := fmt.Errorf("srv.DeleteSession: %w", err)
		logger.Error().Err(errLog).Msg("srv.DeleteSession failed")
		sentryLogger.CaptureFromContext(ctx, errLog, "DeleteSession", map[string]interface{}{
			"action": "delete_session",
		})
		return nil, status.Error(codes.Internal, msgInternalError)
	}

	return &pb.DeleteSessionResponse{}, nil
}

func (h *Handler) ExtendSession(ctx context.Context, req *pb.ExtendSessionRequest) (*pb.ExtendSessionResponse, error) {
	logger := zerolog.Ctx(ctx)

	err := h.srv.ExtendSession(ctx, req.SessionId)
	if err != nil {
		if errors.Is(err, common.ErrorNotExistingSession) {
			return nil, status.Error(codes.NotFound, msgDoesNotExistSession)
		}

		errLog := fmt.Errorf("srv.ExtendSession: %w", err)
		logger.Error().Err(errLog).Msg("srv.ExtendSession failed")
		sentryLogger.CaptureFromContext(ctx, errLog, "ExtendSession", map[string]interface{}{
			"action": "extend_session",
		})
		return nil, status.Error(codes.Internal, msgInternalError)
	}

	return &pb.ExtendSessionResponse{}, nil
}


