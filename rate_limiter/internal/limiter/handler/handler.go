package handler

import (
	"context"
	"time"

	pb "github.com/go-park-mail-ru/2026_1_Clac_Clac/pkg/contracts/rate_limiter"
	"github.com/go-park-mail-ru/2026_1_Clac_Clac/rate_limiter/internal/limiter/service"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

const (
	msgInternalError = "something went wrong"
	msgInvalidInput  = "invalid input parameters"
)

type ServiceLimiter interface {
	CheckRateLimit(ctx context.Context, config service.RateLimiterConfig) (bool, error)
	SetCooldown(ctx context.Context, config service.CooldownConfig) (bool, time.Duration, error)
}

type Handler struct {
	srv ServiceLimiter
	pb.UnimplementedRateLimiterServiceServer
}

func NewHandler(srv ServiceLimiter) *Handler {
	return &Handler{
		srv: srv,
	}
}

func (h *Handler) CheckRateLimit(ctx context.Context, req *pb.CheckRateLimitRequest) (*pb.CheckRateLimitResponse, error) {
	if req.UserIp == "" || req.Action == "" || req.WindowMs <= 0 || req.Limit <= 0 {
		return nil, status.Error(codes.InvalidArgument, msgInvalidInput)
	}

	exceeded, err := h.srv.CheckRateLimit(ctx, service.RateLimiterConfig{
		UserIP: req.UserIp,
		Action: req.Action,
		Window: time.Duration(req.WindowMs) * time.Millisecond,
		Limit:  req.Limit,
	})
	if err != nil {
		return nil, status.Error(codes.Internal, msgInternalError)
	}

	return &pb.CheckRateLimitResponse{
		Exceeded: exceeded,
	}, nil
}

func (h *Handler) SetCooldown(ctx context.Context, req *pb.SetCooldownRequest) (*pb.SetCooldownResponse, error) {
	if req.Name == "" || req.Email == "" || req.ExpirationMs <= 0 {
		return nil, status.Error(codes.InvalidArgument, msgInvalidInput)
	}

	allowed, waitTime, err := h.srv.SetCooldown(ctx, service.CooldownConfig{
		Name:       req.Name,
		Email:      req.Email,
		Expiration: time.Duration(req.ExpirationMs) * time.Millisecond,
	})
	if err != nil {
		return nil, status.Error(codes.Internal, msgInternalError)
	}

	return &pb.SetCooldownResponse{
		Allowed: allowed,
		WaitMs:  waitTime.Milliseconds(),
	}, nil
}
