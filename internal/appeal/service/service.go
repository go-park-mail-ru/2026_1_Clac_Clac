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
	CreateAppeal(ctx context.Context, info repositoryDto.CreateAppealInfo) (repositoryDto.AppealEntry, error)
}

type Service struct {
	rep AppealRepository
}

func (s *Service) CreateAppeal(ctx context.Context, appeal dto.EntityAppeal) error {
	_, err := s.rep.CreateAppeal(ctx, repositoryDto.CreateAppealInfo{
		UserLink:    appeal.UserLink,
		Mail:        appeal.Mail,
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
	return nil
}
