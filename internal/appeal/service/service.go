package service

import (
	"context"
	"fmt"

	"github.com/go-park-mail-ru/2026_1_Clac_Clac/internal/appeal/common"
	repositoryDto "github.com/go-park-mail-ru/2026_1_Clac_Clac/internal/appeal/repository/dto"
	"github.com/go-park-mail-ru/2026_1_Clac_Clac/internal/appeal/service/dto"
	"github.com/google/uuid"
)

type AppealRepository interface {
	GetUserRole(ctx context.Context, userLink uuid.UUID) (common.Role, error)
	CreateAppeal(ctx context.Context, info repositoryDto.CreateAppealInfo) error
	GetUserAppeals(ctx context.Context, userLink uuid.UUID) ([]repositoryDto.AppealEntry, error)
	GetOpenAppeals(ctx context.Context) ([]repositoryDto.AppealEntry, error)
	DeleteAppeal(ctx context.Context, appealLink uuid.UUID) error
	ChangeAppealStatus(ctx context.Context, info repositoryDto.ChangeAppealStatusInfo) error
	GetStats(ctx context.Context) (repositoryDto.AppealStats, error)
}

type Service struct {
	rep AppealRepository
}

func NewService(rep AppealRepository) *Service {
	return &Service{rep: rep}
}

func (s *Service) CreateAppeal(ctx context.Context, appeal dto.EntityAppeal) error {
	err := s.rep.CreateAppeal(ctx, repositoryDto.CreateAppealInfo{
		UserLink:    &appeal.UserLink,
		Email:       appeal.Mail,
		Category:    appeal.Category,
		Description: appeal.Description,
		DisplayName: appeal.DisplayName,
	})

	if err != nil {
		return fmt.Errorf("rep.CreateAppeal: %w", err)
	}

	return nil
}

func (s *Service) DeleteAppeal(ctx context.Context, appealLink uuid.UUID) error {
	err := s.rep.DeleteAppeal(ctx, appealLink)
	if err != nil {
		return fmt.Errorf("rep.DeleteAppeal: %w", err)
	}

	return nil
}

func (s *Service) ChangeAppealStatus(ctx context.Context, info repositoryDto.ChangeAppealStatusInfo) error {
	err := s.rep.ChangeAppealStatus(ctx, info)
	if err != nil {
		return fmt.Errorf("rep.ChangeAppealStatus: %w", err)
	}

	return nil
}

func (s *Service) GetStats(ctx context.Context) (repositoryDto.AppealStats, error) {
	stats, err := s.rep.GetStats(ctx)
	if err != nil {
		return repositoryDto.AppealStats{}, fmt.Errorf("rep.GetStats: %w", err)
	}

	return stats, nil
}
