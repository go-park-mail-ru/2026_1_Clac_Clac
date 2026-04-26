package clients

import (
	"context"
	"fmt"

	"github.com/go-park-mail-ru/2026_1_Clac_Clac/facade/internal/domain"
	pb "github.com/go-park-mail-ru/2026_1_Clac_Clac/pkg/contracts/rate_limiter"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type RateLimiter struct {
	client pb.RateLimiterServiceClient
}

func NewRateLimiterClient(connection *grpc.ClientConn) *RateLimiter {
	return &RateLimiter{
		client: pb.NewRateLimiterServiceClient(connection),
	}
}

func convertRateLimiterGRPCError(err error) error {
	st, ok := status.FromError(err)
	if !ok {
		return err
	}
	if st.Code() == codes.InvalidArgument {
		return ErrInvalidInput
	}
	return err
}

func (r *RateLimiter) CheckRateLimit(ctx context.Context, check domain.RateLimitCheck) (bool, error) {
	req := &pb.CheckRateLimitRequest{
		UserIp:   check.UserIp,
		Action:   check.Action,
		WindowMs: check.WindowMs,
		Limit:    check.Limit,
	}

	resp, err := r.client.CheckRateLimit(ctx, req)
	if err != nil {
		return false, fmt.Errorf("client.CheckRateLimit: %w", convertRateLimiterGRPCError(err))
	}

	return resp.Exceeded, nil
}

func (r *RateLimiter) SetCooldown(ctx context.Context, cooldown domain.Cooldown) (domain.CooldownResult, error) {
	req := &pb.SetCooldownRequest{
		Name:         cooldown.Name,
		Email:        cooldown.Email,
		ExpirationMs: cooldown.ExpirationMs,
	}

	resp, err := r.client.SetCooldown(ctx, req)
	if err != nil {
		return domain.CooldownResult{}, fmt.Errorf("client.SetCooldown: %w", convertRateLimiterGRPCError(err))
	}

	return domain.CooldownResult{
		Allowed: resp.Allowed,
		WaitMs:  resp.WaitMs,
	}, nil
}
