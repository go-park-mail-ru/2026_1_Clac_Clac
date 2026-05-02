package usecase

import (
	"context"
	"fmt"

	"github.com/go-park-mail-ru/2026_1_Clac_Clac/facade/internal/domain"
)

//go:generate mockery --name=SectionClient --output=mock_section_client
type SectionClient interface {
	GetSections(ctx context.Context, sectionReq domain.GetSectionsRequest) ([]domain.SectionInfo, error)
	GetSection(ctx context.Context, sectionReq domain.GetSectionRequest) (domain.SectionInfo, error)
	GetCards(ctx context.Context, cardReq domain.GetCardsRequest) ([]domain.CardInfo, error)
	CreateSection(ctx context.Context, sectionInfo domain.CreateSectionRequest) (domain.SectionInfo, error)
	DeleteSection(ctx context.Context, sectionInfo domain.DeleteSectionRequest) error
	ReorderSection(ctx context.Context, sectionInfo domain.ReorderSectionRequest) error
	UpdateSection(ctx context.Context, sectionInfo domain.UpdateSectionRequest) error
}

type Section struct {
	client SectionClient
}

func NewSection(client SectionClient) *Section {
	return &Section{
		client: client,
	}
}

func (s *Section) GetSections(ctx context.Context, sectionReq domain.GetSectionsRequest) ([]domain.SectionInfo, error) {
	sections, err := s.client.GetSections(ctx, sectionReq)
	if err != nil {
		return nil, fmt.Errorf("section.GetSections: %w", err)
	}

	return sections, nil
}

func (s *Section) GetSection(ctx context.Context, sectionReq domain.GetSectionRequest) (domain.SectionInfo, error) {
	section, err := s.client.GetSection(ctx, sectionReq)
	if err != nil {
		return domain.SectionInfo{}, fmt.Errorf("section.GetSection: %w", err)
	}

	return section, nil
}

func (s *Section) GetCards(ctx context.Context, cardReq domain.GetCardsRequest) ([]domain.CardInfo, error) {
	cards, err := s.client.GetCards(ctx, cardReq)
	if err != nil {
		return nil, fmt.Errorf("section.GetCards: %w", err)
	}

	return cards, nil
}

func (s *Section) CreateSection(ctx context.Context, sectionInfo domain.CreateSectionRequest) (domain.SectionInfo, error) {
	section, err := s.client.CreateSection(ctx, sectionInfo)
	if err != nil {
		return domain.SectionInfo{}, fmt.Errorf("section.CreateSection: %w", err)
	}

	return section, nil
}

func (s *Section) DeleteSection(ctx context.Context, sectionInfo domain.DeleteSectionRequest) error {
	err := s.client.DeleteSection(ctx, sectionInfo)
	if err != nil {
		return fmt.Errorf("section.DeleteSection: %w", err)
	}

	return nil
}

func (s *Section) ReorderSection(ctx context.Context, sectionInfo domain.ReorderSectionRequest) error {
	err := s.client.ReorderSection(ctx, sectionInfo)
	if err != nil {
		return fmt.Errorf("section.ReorderSection: %w", err)
	}

	return nil
}

func (s *Section) UpdateSection(ctx context.Context, sectionInfo domain.UpdateSectionRequest) error {
	err := s.client.UpdateSection(ctx, sectionInfo)
	if err != nil {
		return fmt.Errorf("section.UpdateSection: %w", err)
	}

	return nil
}
