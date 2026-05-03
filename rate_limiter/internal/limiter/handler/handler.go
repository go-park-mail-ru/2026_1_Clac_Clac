package handler

import (
	"context"
	"fmt"
	"time"

	sentryLogger "github.com/go-park-mail-ru/2026_1_Clac_Clac/pkg/logger"
	pb "github.com/go-park-mail-ru/2026_1_Clac_Clac/pkg/proto/rate_limiter/v1"
	serviceDto "github.com/go-park-mail-ru/2026_1_Clac_Clac/rate_limiter/internal/limiter/service/dto"
	"github.com/rs/zerolog"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

const (
	msgInternalError = "something went wrong"
	msgInvalidInput  = "invalid input parameters"
)

type ServiceLimiter interface {
	CheckRateLimit(ctx context.Context, config serviceDto.RateLimiterConfig) (bool, error)
	SetCooldown(ctx context.Context, config serviceDto.CooldownConfig) (bool, time.Duration, error)
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
	logger := zerolog.Ctx(ctx)

	if req.UserIp == "" || req.Action == "" || req.WindowS <= 0 || req.Limit <= 0 {
		return nil, status.Error(codes.InvalidArgument, msgInvalidInput)
	}

	exceeded, err := h.srv.CheckRateLimit(ctx, serviceDto.RateLimiterConfig{
		UserIP: req.UserIp,
		Action: req.Action,
		Window: time.Duration(req.WindowS) * time.Second,
		Limit:  req.Limit,
	})
	if err != nil {
		errLog := fmt.Errorf("srv.CheckRateLimit: %w", err)
		logger.Error().Err(errLog).Msg("srv.CheckRateLimit failed")
		sentryLogger.CaptureFromContext(ctx, errLog, "CheckRateLimit", map[string]interface{}{
			"user_ip": req.UserIp,
			"action":  req.Action,
		})
		return nil, status.Error(codes.Internal, msgInternalError)
	}

	return &pb.CheckRateLimitResponse{
		Exceeded: exceeded,
	}, nil
}

func (h *Handler) SetCooldown(ctx context.Context, req *pb.SetCooldownRequest) (*pb.SetCooldownResponse, error) {
	logger := zerolog.Ctx(ctx)

	if req.Name == "" || req.Email == "" || req.ExpirationS <= 0 {
		return nil, status.Error(codes.InvalidArgument, msgInvalidInput)
	}

	allowed, waitTime, err := h.srv.SetCooldown(ctx, serviceDto.CooldownConfig{
		Name:       req.Name,
		Email:      req.Email,
		Expiration: time.Duration(req.ExpirationS) * time.Second,
	})
	if err != nil {
		errLog := fmt.Errorf("srv.SetCooldown: %w", err)
		logger.Error().Err(errLog).Msg("srv.SetCooldown failed")
		sentryLogger.CaptureFromContext(ctx, errLog, "SetCooldown", map[string]interface{}{
			"name":   req.Name,
			"email":  req.Email,
			"action": "set_cooldown",
		})
		return nil, status.Error(codes.Internal, msgInternalError)
	}

	return &pb.SetCooldownResponse{
		Allowed: allowed,
		WaitS:   int64(waitTime.Seconds()),
	}, nil
}
