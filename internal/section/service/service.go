package service

import (
	"context"
	"fmt"

	"github.com/go-park-mail-ru/2026_1_Clac_Clac/internal/common"
	repositoryDto "github.com/go-park-mail-ru/2026_1_Clac_Clac/internal/section/repository/dto"
	"github.com/go-park-mail-ru/2026_1_Clac_Clac/internal/section/service/dto"
	"github.com/google/uuid"
)

// mockery --name=SectionRepository --output=mock_section_rep --outpkg=mockSectionRep
type SectionRepository interface {
	GetSectionInfo(ctx context.Context, link uuid.UUID) (repositoryDto.FullSectionInfo, error)
	GetAllSections(ctx context.Context, boarderLink uuid.UUID) ([]repositoryDto.FullSectionInfo, error)
	CreateSection(ctx context.Context, newSection repositoryDto.CreatingSection) (repositoryDto.FullSectionInfo, error)
	DeleteSection(ctx context.Context, linkSection uuid.UUID) error
	ReorderSection(ctx context.Context, linkBoard uuid.UUID, sectionLinks []uuid.UUID) error
	UpdateSection(ctx context.Context, updatingSection repositoryDto.FullSectionInfo) error
}

type Deps struct {
	Rep SectionRepository
}

type Service struct {
	deps Deps
}

func NewService(deps Deps) *Service {
	return &Service{
		deps: deps,
	}
}

func (s *Service) GetSectionInfo(ctx context.Context, link uuid.UUID) (dto.FullSectionInfo, error) {
	result, err := s.deps.Rep.GetSectionInfo(ctx, link)
	if err != nil {
		return dto.FullSectionInfo{}, fmt.Errorf("rep.GetSectionInfo: %w", err)
	}

	sectionInfo := dto.FullSectionInfo{
		SectionName: result.SectionName,
		Position:    result.Position,
		IsMandatory: result.IsMandatory,
		Color:       result.Color,
		MaxTasks:    result.MaxTasks,
	}

	return sectionInfo, nil
}

func (s *Service) CreateSection(ctx context.Context, newSection dto.CreatingSection) (dto.EntitySection, error) {
	sectionLink := uuid.New()

	result, err := s.deps.Rep.CreateSection(ctx, repositoryDto.CreatingSection{
		SectionLink: sectionLink,
		BoardLink:   newSection.BoardLink,
		SectionName: newSection.SectionName,
		IsMandatory: newSection.IsMandatory,
		Color:       newSection.Color,
		MaxTasks:    newSection.MaxTasks,
	})
	if err != nil {
		return dto.EntitySection{}, fmt.Errorf("rep.CreateSection: %w", err)
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

func (s *Service) DeleteSection(ctx context.Context, linksSection uuid.UUID) error {
	info, err := s.deps.Rep.GetSectionInfo(ctx, linksSection)
	if err != nil {
		return fmt.Errorf("rep.GetSectionInfo: %w", err)
	}

	if info.Position == 1 {
		return common.ErrorDeleteBacklog
	}

	err = s.deps.Rep.DeleteSection(ctx, linksSection)
	if err != nil {
		return fmt.Errorf("rep.DeleteSection: %w", err)
	}

	return nil
}

func (s *Service) ReorderSection(ctx context.Context, linkBoard uuid.UUID, sectionLinks []uuid.UUID) error {
	if len(sectionLinks) == 0 {
		return nil
	}

	err := s.deps.Rep.ReorderSection(ctx, linkBoard, sectionLinks)
	if err != nil {
		return fmt.Errorf("rep.ReorderSection: %w", err)
	}

	return nil
}

func (s *Service) UpdateSection(ctx context.Context, updatingSection dto.FullSectionInfo) error {
	info, err := s.deps.Rep.GetSectionInfo(ctx, updatingSection.SectionLink)
	if err != nil {
		return fmt.Errorf("rep.GetSectionInfo: %w", err)
	}

	if info.Position == 1 {
		return common.ErrorUpdateBacklog
	}

	err = s.deps.Rep.UpdateSection(ctx, repositoryDto.FullSectionInfo{
		SectionLink: updatingSection.SectionLink,
		SectionName: updatingSection.SectionName,
		Position:    updatingSection.Position,
		IsMandatory: updatingSection.IsMandatory,
		Color:       updatingSection.Color,
		MaxTasks:    updatingSection.MaxTasks,
	})
	if err != nil {
		return fmt.Errorf("rep.UpdateSection: %w", err)
	}

	return nil
}

func (s *Service) GetAllSections(ctx context.Context, boarderLink uuid.UUID) ([]dto.FullSectionInfo, error) {
	sections, err := s.deps.Rep.GetAllSections(ctx, boarderLink)
	if err != nil {
		return []dto.FullSectionInfo{}, fmt.Errorf("rep.GetAllSections: %w", err)
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
