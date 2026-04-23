package handler

import (
	"context"
	"errors"

	"github.com/go-park-mail-ru/2026_1_Clac_Clac/authorization/internal/common"
	pb "github.com/go-park-mail-ru/2026_1_Clac_Clac/pkg/contracts/auth"
	"github.com/google/uuid"
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
	convertedLink, err := uuid.Parse(req.UserLink)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, msgErrorParseUserLink)
	}

	sessionID, err := h.srv.CreateSession(ctx, convertedLink)
	if err != nil {
		return nil, status.Error(codes.Internal, msgInternalError)
	}

	return &pb.CreateSessionResponse{
		SessionId: sessionID,
	}, nil
}

func (h *Handler) GetUserLink(ctx context.Context, req *pb.GetUserLinkRequest) (*pb.GetUserLinkResponse, error) {
	userLink, err := h.srv.GetUserLink(ctx, req.SessionId)
	if err != nil {
		if errors.Is(err, common.ErrorNotExistingSession) {
			return nil, status.Error(codes.NotFound, msgDoesNotExistSession)
		}

		return nil, status.Error(codes.Internal, msgInternalError)
	}

	return &pb.GetUserLinkResponse{
		UserLink: userLink,
	}, nil
}

func (h *Handler) DeleteSession(ctx context.Context, req *pb.DeleteSessionRequest) (*pb.DeleteSessionResponse, error) {
	err := h.srv.DeleteSession(ctx, req.SessionId)
	if err != nil {
		return nil, status.Error(codes.Internal, msgInternalError)
	}

	return &pb.DeleteSessionResponse{}, nil
}

func (h *Handler) ExtendSession(ctx context.Context, req *pb.ExtendSessionRequest) (*pb.ExtendSessionResponse, error) {
	err := h.srv.ExtendSession(ctx, req.SessionId)
	if err != nil {
		if errors.Is(err, common.ErrorNotExistingSession) {
			return nil, status.Error(codes.NotFound, msgDoesNotExistSession)
		}

		return nil, status.Error(codes.Internal, msgInternalError)
	}

	return &pb.ExtendSessionResponse{}, nil
}
