package service

import (
	"context"
	"fmt"

	"github.com/go-park-mail-ru/2026_1_Clac_Clac/internal/appeal/common"
	repositoryDto "github.com/go-park-mail-ru/2026_1_Clac_Clac/internal/appeal/repository/dto"
	"github.com/go-park-mail-ru/2026_1_Clac_Clac/internal/appeal/service/dto"
	"github.com/google/uuid"
)

//go:generate mockery --name=AppealRepository --output mock_appeal_rep
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

func (s *Service) ChangeAppealStatus(ctx context.Context, info dto.ChangeAppealStatusInfo) error {
	userRole, err := s.rep.GetUserRole(ctx, info.SupporterLink)
	if err != nil {
		return fmt.Errorf("rep.GetUserRole: %w", err)
	}

	if userRole == common.Roles.None {
		return common.ErrorPermissionDenied
	}

	err = s.rep.ChangeAppealStatus(ctx, repositoryDto.ChangeAppealStatusInfo{
		SupporterLink: info.SupporterLink,
		AppealLink:    info.AppealLink,
		Status:        info.Status,
	})
	if err != nil {
		return fmt.Errorf("rep.ChangeAppealStatus: %w", err)
	}

	return nil
}

func (s *Service) GetStats(ctx context.Context, userLink uuid.UUID) (dto.AppealStats, error) {
	userRole, err := s.rep.GetUserRole(ctx, userLink)
	if err != nil {
		return dto.AppealStats{}, fmt.Errorf("rep.GetUserRole: %w", err)
	}

	if userRole != common.Roles.Admin {
		return dto.AppealStats{}, common.ErrorPermissionDenied
	}

	stats, err := s.rep.GetStats(ctx)
	if err != nil {
		return dto.AppealStats{}, fmt.Errorf("rep.GetStats: %w", err)
	}

	return dto.AppealStats{
		Open:   stats.Open,
		InWork: stats.InWork,
		Close:  stats.Close,
	}, nil
}

func (s *Service) GetAppeals(ctx context.Context, userLink uuid.UUID) (dto.Appeals, error) {
	userRole, err := s.rep.GetUserRole(ctx, userLink)
	if err != nil {
		return dto.Appeals{}, fmt.Errorf("ServiceAppeal.GetUserRole: %w", err)
	}

	var rawAppeals []repositoryDto.AppealEntry
	switch userRole {
	case common.Roles.Support, common.Roles.Admin:
		rawAppeals, err = s.rep.GetOpenAppeals(ctx)
		if err != nil {
			return dto.Appeals{}, fmt.Errorf("ServiceAppeal.GetOpenAppeals: %w", err)
		}
	default:
		rawAppeals, err = s.rep.GetUserAppeals(ctx, userLink)
		if err != nil {
			return dto.Appeals{}, fmt.Errorf("SerivceAppeal.GetUserAppeals: %w", err)
		}
	}

	appeals := dto.Appeals{
		Role:    userRole,
		Appeals: make([]dto.Appeal, 0, len(rawAppeals)),
	}

	for _, rawAppeal := range rawAppeals {
		appeals.Appeals = append(appeals.Appeals, dto.Appeal{
			AppelID:       rawAppeal.AppealID,
			AppealLink:    rawAppeal.AppealLink,
			Email:         rawAppeal.Email,
			DisplayName:   rawAppeal.DisplayName,
			Status:        rawAppeal.Status,
			Category:      rawAppeal.Category,
			Description:   rawAppeal.Description,
			AttachmentKey: rawAppeal.AttachmentKey,
			CreatedAt:     rawAppeal.CreatedAt,
		})
	}

	return appeals, nil
}
