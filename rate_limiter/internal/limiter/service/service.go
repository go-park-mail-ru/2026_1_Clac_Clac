package service

import (
	"context"
	"fmt"
	"time"

	repositoryDto "github.com/go-park-mail-ru/2026_1_Clac_Clac/rate_limiter/internal/limiter/repository/dto"
	"github.com/go-park-mail-ru/2026_1_Clac_Clac/rate_limiter/internal/limiter/service/dto"
)

type LimiterRepository interface {
	CheckLimit(ctx context.Context, config repositoryDto.RateLimiterConfig) (int64, error)
	SetCooldown(ctx context.Context, config repositoryDto.CooldownConfig) (bool, time.Duration, error)
}

type Service struct {
	rep LimiterRepository
}

func NewService(rep LimiterRepository) *Service {
	return &Service{
		rep: rep,
	}
}

func (s *Service) CheckRateLimit(ctx context.Context, config dto.RateLimiterConfig) (bool, error) {
	count, err := s.rep.CheckLimit(ctx, repositoryDto.RateLimiterConfig{
		UserIP: config.UserIP,
		Action: config.Action,
		Window: config.Window,
	})
	if err != nil {
		return false, fmt.Errorf("rep.CheckLimit: %w", err)
	}

	return count > config.Limit, nil
}

func (s *Service) SetCooldown(ctx context.Context, config dto.CooldownConfig) (bool, time.Duration, error) {
	key := fmt.Sprintf("cd:%s:%s", config.Name, config.Email)

	allowed, waitTime, err := s.rep.SetCooldown(ctx, repositoryDto.CooldownConfig{
		Key:        key,
		Expiration: config.Expiration,
	})
	if err != nil {
		return false, 0, fmt.Errorf("rep.SetCooldown: %w", err)
	}

	return allowed, waitTime, nil
}
