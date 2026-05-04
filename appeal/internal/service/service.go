package service

import (
	"context"
	"errors"
	"fmt"
	"io"

	repositoryDto "github.com/go-park-mail-ru/2026_1_Clac_Clac/appeal/internal/repository/dto"
	"github.com/go-park-mail-ru/2026_1_Clac_Clac/appeal/internal/service/dto"
	rbac "github.com/go-park-mail-ru/2026_1_Clac_Clac/pkg/appealRbac"
	"github.com/google/uuid"
	"github.com/rs/zerolog"
)

//go:generate mockery --name=AppealRepository --output mock_appeal_rep
type AppealRepository interface {
	CreateAppeal(ctx context.Context, info repositoryDto.CreateAppealInfo) (uuid.UUID, error)
	GetUserAppeals(ctx context.Context, userLink uuid.UUID) ([]repositoryDto.AppealEntry, error)
	GetOpenAppeals(ctx context.Context, supportLink uuid.UUID) ([]repositoryDto.AppealEntry, error)
	DeleteAppeal(ctx context.Context, appealLink uuid.UUID) error
	ChangeAppealStatus(ctx context.Context, info repositoryDto.ChangeAppealStatusInfo) error
	GetStats(ctx context.Context) (repositoryDto.AppealStats, error)
	UploadAttachment(ctx context.Context, source io.Reader, filename, contentType string) (string, error)
	UpdateAttachmentKey(ctx context.Context, key string, appealLink uuid.UUID) error
}

type Service struct {
	rep               AppealRepository
	permissionChecker rbac.Service
}

func NewService(rep AppealRepository, permissionChecker rbac.Service) *Service {
	return &Service{rep: rep, permissionChecker: permissionChecker}
}

func (s *Service) CreateAppeal(ctx context.Context, appeal dto.EntityAppeal) (uuid.UUID, error) {
	err := s.permissionChecker.CheckPermission(ctx, appeal.UserLink, rbac.Actions.Create)
	if err != nil {
		if errors.Is(err, rbac.ErrActionDenied) {
			return uuid.Nil, rbac.ErrActionDenied
		}

		return uuid.Nil, fmt.Errorf("s.CheckPermission: %w", err)
	}

	appealLink, err := s.rep.CreateAppeal(ctx, repositoryDto.CreateAppealInfo{
		UserLink:    &appeal.UserLink,
		Email:       appeal.Mail,
		Category:    appeal.Category,
		Description: appeal.Description,
		DisplayName: appeal.DisplayName,
	})

	if err != nil {
		return uuid.UUID{}, fmt.Errorf("rep.CreateAppeal: %w", err)
	}

	return appealLink, nil
}

func (s *Service) DeleteAppeal(ctx context.Context, appealLink uuid.UUID, userLink uuid.UUID) error {
	err := s.permissionChecker.CheckPermission(ctx, userLink, rbac.Actions.Delete)
	if err != nil {
		if errors.Is(err, rbac.ErrActionDenied) {
			return rbac.ErrActionDenied
		}

		return fmt.Errorf("s.CheckPermission: %w", err)
	}

	err = s.rep.DeleteAppeal(ctx, appealLink)
	if err != nil {
		return fmt.Errorf("rep.DeleteAppeal: %w", err)
	}

	return nil
}

func (s *Service) ChangeAppealStatus(ctx context.Context, info dto.ChangeAppealStatusInfo) error {
	err := s.permissionChecker.CheckPermission(ctx, info.SupporterLink, rbac.Actions.ChangeStatus)
	if err != nil {
		logger := zerolog.Ctx(ctx)
		logger.Error().Err(err).Msg("permission check failed for ChangeAppealStatus, returning ErrActionDenied")
		return rbac.ErrActionDenied
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
	err := s.permissionChecker.CheckPermission(ctx, userLink, rbac.Actions.ViewStats)
	if err != nil {
		logger := zerolog.Ctx(ctx)
		logger.Error().Err(err).Msg("permission check failed for GetStats, returning ErrActionDenied")
		return dto.AppealStats{}, rbac.ErrActionDenied
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

func (s *Service) UploadAttachment(ctx context.Context, file io.Reader, contentType, extension string, appealLink uuid.UUID, userLink uuid.UUID) (string, error) {
	err := s.permissionChecker.CheckPermission(ctx, userLink, rbac.Actions.Edit)
	if err != nil {
		if errors.Is(err, rbac.ErrActionDenied) {
			return "", rbac.ErrActionDenied
		}

		return "", fmt.Errorf("s.CheckPermission: %w", err)
	}

	filename := uuid.New().String() + extension

	key, err := s.rep.UploadAttachment(ctx, file, filename, contentType)
	if err != nil {
		return "", fmt.Errorf("rep.UploadAttachment: %w", err)
	}

	if err := s.rep.UpdateAttachmentKey(ctx, key, appealLink); err != nil {
		return "", fmt.Errorf("rep.UpdateAttachmentKey: %w", err)
	}

	return key, nil
}

func (s *Service) GetAppeals(ctx context.Context, userLink uuid.UUID) (dto.Appeals, error) {
	// Считаю, что у всех пользователей есть возможность читать обращения
	// Поэтому получаю роль без проверки
	userRole, err := s.permissionChecker.GetUserRole(ctx, userLink)
	if err != nil {
		return dto.Appeals{}, fmt.Errorf("s.GetUserRole: %w", err)
	}

	var rawAppeals = make([]repositoryDto.AppealEntry, 0)
	switch userRole {
	case rbac.Roles.Support, rbac.Roles.Admin:
		rawAppeals, err = s.rep.GetOpenAppeals(ctx, userLink)
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
			AppealID:      rawAppeal.AppealID,
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
