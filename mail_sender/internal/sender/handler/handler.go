package handler

import (
	"context"
	"errors"

	"github.com/go-park-mail-ru/2026_1_Clac_Clac/mail_sender/internal/common"
	pb "github.com/go-park-mail-ru/2026_1_Clac_Clac/pkg/proto/mail_sender/v1"
	"github.com/google/uuid"
	"github.com/rs/zerolog"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

const (
	msgInternalError = "something went wrong"
	msgInvalidInput  = "invalid input parameters"

	msgDoesNotExistResetToken = "reset token does not exist"
)

type ServiceSender interface {
	SendRecoveryCode(ctx context.Context, userLink uuid.UUID, email string) error
	CheckRecoveryCode(ctx context.Context, tokenID string) error
	GetUserLink(ctx context.Context, tokenID string) (string, error)
}

type Handler struct {
	srv ServiceSender
	pb.UnimplementedMailSenderServiceServer
}

func NewHandler(srv ServiceSender) *Handler {
	return &Handler{
		srv: srv,
	}
}

func (h *Handler) SendRecoveryCode(ctx context.Context, req *pb.SendRecoveryCodeRequest) (*pb.SendRecoveryCodeResponse, error) {
	logger := zerolog.Ctx(ctx)

	convertedUserLink, err := uuid.Parse(req.UserLink)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, msgInvalidInput)
	}

	err = h.srv.SendRecoveryCode(ctx, convertedUserLink, req.Email)
	if err != nil {
		logger.Error().Err(err).Msg("srv.SendRecoveryCode failed")
		return nil, status.Error(codes.Internal, msgInternalError)
	}

	return &pb.SendRecoveryCodeResponse{}, nil
}

func (h *Handler) CheckRecoveryCode(ctx context.Context, req *pb.CheckRecoveryCodeRequest) (*pb.CheckRecoveryCodeResponse, error) {
	logger := zerolog.Ctx(ctx)

	err := h.srv.CheckRecoveryCode(ctx, req.Code)
	if err != nil {
		if errors.Is(err, common.ErrorNotExistingResetToken) {
			return nil, status.Error(codes.NotFound, msgDoesNotExistResetToken)
		}

		logger.Error().Err(err).Msg("srv.CheckRecoveryCode failed")
		return nil, status.Error(codes.Internal, msgInternalError)
	}

	return &pb.CheckRecoveryCodeResponse{}, nil
}

func (h *Handler) ExchangeTokenForUser(ctx context.Context, req *pb.ExchangeTokenRequest) (*pb.ExchangeTokenResponse, error) {
	logger := zerolog.Ctx(ctx)

	userLink, err := h.srv.GetUserLink(ctx, req.ResetToken)
	if err != nil {
		if errors.Is(err, common.ErrorNotExistingResetToken) {
			return nil, status.Error(codes.NotFound, msgDoesNotExistResetToken)
		}

		logger.Error().Err(err).Msg("srv.GetUserLink failed")
		return nil, status.Error(codes.Internal, msgInternalError)
	}

	return &pb.ExchangeTokenResponse{
		UserLink: userLink,
	}, nil
}
