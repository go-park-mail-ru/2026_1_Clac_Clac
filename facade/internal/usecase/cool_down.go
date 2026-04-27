package usecase

import (
	"context"
	"fmt"

	"github.com/go-park-mail-ru/2026_1_Clac_Clac/facade/internal/domain"
)

type RateLimiterClient interface {
	SetCooldown(ctx context.Context, cooldown domain.Cooldown) (domain.CooldownResult, error)
}
type CoolDown struct {
	client RateLimiterClient
}

func NewCoolDown(client RateLimiterClient) *CoolDown {
	return &CoolDown{
		client: client,
	}
}

func (c *CoolDown) CheckCoolDown(ctx context.Context, cooldown domain.Cooldown) (domain.CooldownResult, error) {
	info, err := c.client.SetCooldown(ctx, cooldown)
	if err != nil {
		return domain.CooldownResult{}, fmt.Errorf("client.SetCooldown: %w", err)
	}

	return info, nil
}
