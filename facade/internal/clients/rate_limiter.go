package clients

import (
	"context"
	"fmt"

	"github.com/go-park-mail-ru/2026_1_Clac_Clac/facade/internal/domain"
	pb "github.com/go-park-mail-ru/2026_1_Clac_Clac/pkg/proto/rate_limiter/v1"
	"google.golang.org/grpc"
)

type RateLimiter struct {
	client pb.RateLimiterServiceClient
}

func NewRateLimiterClient(connection *grpc.ClientConn) *RateLimiter {
	return &RateLimiter{
		client: pb.NewRateLimiterServiceClient(connection),
	}
}

func (r *RateLimiter) UpdateCountRequests(ctx context.Context, check domain.RateLimitCheck) (bool, error) {
	req := &pb.CheckRateLimitRequest{
		UserIp:  check.UserIP,
		Action:  check.Action,
		WindowS: check.WindowS,
		Limit:   check.Limit,
	}

	resp, err := r.client.CheckRateLimit(ctx, req)
	if err != nil {
		return false, fmt.Errorf("client.CheckRateLimit: %w", convertGRPCError(err))
	}

	return resp.Exceeded, nil
}

func (r *RateLimiter) SetCooldown(ctx context.Context, cooldown domain.Cooldown) (domain.CooldownResult, error) {
	req := &pb.SetCooldownRequest{
		Name:        cooldown.Name,
		Email:       cooldown.Email,
		ExpirationS: cooldown.ExpirationS,
	}

	resp, err := r.client.SetCooldown(ctx, req)
	if err != nil {
		return domain.CooldownResult{}, fmt.Errorf("client.SetCooldown: %w", convertGRPCError(err))
	}

	return domain.CooldownResult{
		Allowed: resp.Allowed,
		WaitS:   resp.WaitS,
	}, nil
}
