package service

import (
	"context"
	"errors"
	"fmt"

	"github.com/go-park-mail-ru/2026_1_Clac_Clac/board/internal/section/common"
	repositoryDto "github.com/go-park-mail-ru/2026_1_Clac_Clac/board/internal/section/repository/dto"
	"github.com/go-park-mail-ru/2026_1_Clac_Clac/board/internal/section/service/dto"
	rbac "github.com/go-park-mail-ru/2026_1_Clac_Clac/pkg/boardRbac"
	"github.com/google/uuid"
)

//go:generate mockery --name=SectionRepository --output=mock_section_rep --outpkg=mockSectionRep
type SectionRepository interface {
	GetSection(ctx context.Context, link uuid.UUID) (repositoryDto.FullSectionInfo, error)
	GetSections(ctx context.Context, boardLink uuid.UUID) ([]repositoryDto.FullSectionInfo, error)
	GetCards(ctx context.Context, linkSection uuid.UUID) ([]repositoryDto.Card, error)
	CreateSection(ctx context.Context, newSection repositoryDto.CreatingSection) (repositoryDto.FullSectionInfo, error)
	DeleteSection(ctx context.Context, linkSection uuid.UUID) error
	ReorderSection(ctx context.Context, linkBoard uuid.UUID, sectionLinks []uuid.UUID) error
	UpdateSection(ctx context.Context, updatingSection repositoryDto.FullSectionInfo) error
}

type Service struct {
	rep               SectionRepository
	permissionChecker rbac.Service
}

func NewService(rep SectionRepository, permissionChecker rbac.Service) *Service {
	return &Service{
		rep:               rep,
		permissionChecker: permissionChecker,
	}
}

func (s *Service) GetSection(ctx context.Context, sectionLink uuid.UUID, userLink uuid.UUID) (dto.FullSectionInfo, error) {
	err := s.permissionChecker.CheckPermissionOnSection(ctx, sectionLink, userLink, rbac.Actions.View)
	if err != nil {
		if errors.Is(err, rbac.ErrActionDenied) {
			return dto.FullSectionInfo{}, rbac.ErrActionDenied
		}

		return dto.FullSectionInfo{}, fmt.Errorf("SectionService.CheckPermissionOnSection: %w", err)
	}

	result, err := s.rep.GetSection(ctx, sectionLink)
	if err != nil {
		return dto.FullSectionInfo{}, fmt.Errorf("SectionRepository.GetSectionInfo: %w", err)
	}

	sectionInfo := dto.FullSectionInfo{
		SectionLink: result.SectionLink,
		SectionName: result.SectionName,
		Position:    result.Position,
		IsMandatory: result.IsMandatory,
		Color:       result.Color,
		MaxTasks:    result.MaxTasks,
	}

	return sectionInfo, nil
}

func (s *Service) GetSections(ctx context.Context, boardLink uuid.UUID, userLink uuid.UUID) ([]dto.FullSectionInfo, error) {
	err := s.permissionChecker.CheckPermissionOnBoard(ctx, boardLink, userLink, rbac.Actions.View)
	if err != nil {
		if errors.Is(err, rbac.ErrActionDenied) {
			return []dto.FullSectionInfo{}, rbac.ErrActionDenied
		}

		return []dto.FullSectionInfo{}, fmt.Errorf("SectionService.CheckPermissionOnBoard: %w", err)
	}

	sections, err := s.rep.GetSections(ctx, boardLink)
	if err != nil {
		return []dto.FullSectionInfo{}, fmt.Errorf("SectionRepository.GetAllSections: %w", err)
	}

	var convertSections []dto.FullSectionInfo
	for _, section := range sections {
		convertSections = append(convertSections, dto.FullSectionInfo{
			SectionLink: section.SectionLink,
			SectionName: section.SectionName,
			Position:    section.Position,
			IsMandatory: section.IsMandatory,
			Color:       section.Color,
			MaxTasks:    section.MaxTasks,
		})
	}

	return convertSections, nil
}

func (s *Service) GetCards(ctx context.Context, sectionLink uuid.UUID, userLink uuid.UUID) ([]dto.Card, error) {
	err := s.permissionChecker.CheckPermissionOnSection(ctx, sectionLink, userLink, rbac.Actions.View)
	if err != nil {
		if errors.Is(err, rbac.ErrActionDenied) {
			return []dto.Card{}, rbac.ErrActionDenied
		}

		return []dto.Card{}, fmt.Errorf("SectionService.CheckPermissionOnSection: %w", err)
	}

	cards, err := s.rep.GetCards(ctx, sectionLink)
	if err != nil {
		return []dto.Card{}, fmt.Errorf("SectionRepository.GetCards: %w", err)
	}

	convertCards := make([]dto.Card, 0, len(cards))
	for _, card := range cards {
		convertCards = append(convertCards, dto.Card{
			CardLink:     card.CardLink,
			ExecutorLink: card.ExecutorLink,
			Title:        card.Title,
			DeadLine:     card.DeadLine,
			Subtasks:     card.Subtasks,
			Position:    card.Position,
		})
	}

	return convertCards, nil
}

func (s *Service) CreateSection(ctx context.Context, newSection dto.CreatingSection, userLink uuid.UUID) (dto.EntitySection, error) {
	err := s.permissionChecker.CheckPermissionOnBoard(ctx, newSection.BoardLink, userLink, rbac.Actions.Edit)
	if err != nil {
		if errors.Is(err, rbac.ErrActionDenied) {
			return dto.EntitySection{}, rbac.ErrActionDenied
		}

		return dto.EntitySection{}, fmt.Errorf("SectionService.CheckPermissionOnBoard: %w", err)
	}

	sectionLink := uuid.New()

	result, err := s.rep.CreateSection(ctx, repositoryDto.CreatingSection{
		SectionLink: sectionLink,
		BoardLink:   newSection.BoardLink,
		SectionName: newSection.SectionName,
		IsMandatory: newSection.IsMandatory,
		Color:       newSection.Color,
		MaxTasks:    newSection.MaxTasks,
	})
	if err != nil {
		return dto.EntitySection{}, fmt.Errorf("SectionRepository.CreateSection: %w", err)
	}

	sectionInfo := dto.EntitySection{
		SectionLink: sectionLink,
		SectionName: result.SectionName,
		Position:    result.Position,
		IsMandatory: result.IsMandatory,
		Color:       result.Color,
		MaxTasks:    result.MaxTasks,
	}

	return sectionInfo, nil
}

func (s *Service) DeleteSection(ctx context.Context, sectionLink uuid.UUID, userLink uuid.UUID) error {
	err := s.permissionChecker.CheckPermissionOnSection(ctx, sectionLink, userLink, rbac.Actions.Edit)
	if err != nil {
		if errors.Is(err, rbac.ErrActionDenied) {
			return rbac.ErrActionDenied
		}

		return fmt.Errorf("SectionService.CheckPermissionOnSection: %w", err)
	}

	info, err := s.rep.GetSection(ctx, sectionLink)
	if err != nil {
		return fmt.Errorf("SectionRepository.GetSectionInfo: %w", err)
	}

	if info.Position == 1 {
		return common.ErrCannotDeleteBacklog
	}

	err = s.rep.DeleteSection(ctx, sectionLink)
	if err != nil {
		return fmt.Errorf("SectionRepository.DeleteSection: %w", err)
	}

	return nil
}

func (s *Service) ReorderSection(ctx context.Context, boardLink uuid.UUID, sectionLinks []uuid.UUID, userLink uuid.UUID) error {
	err := s.permissionChecker.CheckPermissionOnBoard(ctx, boardLink, userLink, rbac.Actions.Edit)
	if err != nil {
		if errors.Is(err, rbac.ErrActionDenied) {
			return rbac.ErrActionDenied
		}

		return fmt.Errorf("SectionService.CheckPermissionOnBoard: %w", err)
	}

	if len(sectionLinks) == 0 {
		return nil
	}

	err = s.rep.ReorderSection(ctx, boardLink, sectionLinks)
	if err != nil {
		return fmt.Errorf("SectionRepository.ReorderSection: %w", err)
	}

	return nil
}

func (s *Service) UpdateSection(ctx context.Context, updatingSection dto.FullSectionInfo, userLink uuid.UUID) error {
	err := s.permissionChecker.CheckPermissionOnSection(ctx, updatingSection.SectionLink, userLink, rbac.Actions.Edit)
	if err != nil {
		if errors.Is(err, rbac.ErrActionDenied) {
			return rbac.ErrActionDenied
		}

		return fmt.Errorf("SectionService.CheckPermissionOnSection: %w", err)
	}

	info, err := s.rep.GetSection(ctx, updatingSection.SectionLink)
	if err != nil {
		return fmt.Errorf("SectionRepository.GetSectionInfo: %w", err)
	}

	if info.Position == 1 {
		return common.ErrCannotUpdateBacklog
	}

	err = s.rep.UpdateSection(ctx, repositoryDto.FullSectionInfo{
		SectionLink: updatingSection.SectionLink,
		SectionName: updatingSection.SectionName,
		Position:    updatingSection.Position,
		IsMandatory: updatingSection.IsMandatory,
		Color:       updatingSection.Color,
		MaxTasks:    updatingSection.MaxTasks,
	})
	if err != nil {
		return fmt.Errorf("SectionRepository.UpdateSection: %w", err)
	}

	return nil
}
