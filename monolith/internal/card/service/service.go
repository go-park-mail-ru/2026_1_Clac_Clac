package service

import (
	"context"
	"fmt"

	repositoryDto "github.com/go-park-mail-ru/2026_1_Clac_Clac/monolith/internal/card/repository/dto"
	"github.com/go-park-mail-ru/2026_1_Clac_Clac/monolith/internal/card/service/dto"
	"github.com/google/uuid"
)

type CardRep interface {
	GetCard(ctx context.Context, linkCard uuid.UUID) (repositoryDto.InfoCard, error)
	DeleteCard(ctx context.Context, linkCard uuid.UUID) error
	UpdateCardDetails(ctx context.Context, updatedCard repositoryDto.UpdatingCardDetails) error
	ReorderCard(ctx context.Context, updatingPlaceCard repositoryDto.PlaceCard) error
	CreateCard(ctx context.Context, newCard repositoryDto.NewCard) (int, error)
}

type Service struct {
	rep CardRep
}

func NewService(rep CardRep) *Service {
	return &Service{
		rep: rep,
	}
}

func (s *Service) GetCard(ctx context.Context, linkCard uuid.UUID) (dto.InfoCard, error) {
	card, err := s.rep.GetCard(ctx, linkCard)
	if err != nil {
		return dto.InfoCard{}, fmt.Errorf("rep.GetCard: %w", err)
	}

	return dto.InfoCard{
		Description:  card.Description,
		Title:        card.Title,
		NameExecuter: card.NameExecuter,
		DataDeadLine: card.DataDeadLine,
	}, nil
}

func (s *Service) DeleteCard(ctx context.Context, linkCard uuid.UUID) error {
	err := s.rep.DeleteCard(ctx, linkCard)
	if err != nil {
		return fmt.Errorf("rep.DeleteCard: %w", err)
	}

	return nil
}

func (s *Service) UpdateCardDetails(ctx context.Context, updatingCard dto.UpdatingCardDetails) error {
	err := s.rep.UpdateCardDetails(ctx, repositoryDto.UpdatingCardDetails{
		LinkCard:     updatingCard.LinkCard,
		Description:  updatingCard.Description,
		Title:        updatingCard.Title,
		LinkExecuter: updatingCard.LinkExecuter,
		DataDeadLine: updatingCard.DataDeadLine,
	})
	if err != nil {
		return fmt.Errorf("rep.UpdateCardDetails: %w", err)
	}

	return nil
}

func (s *Service) ReorderCard(ctx context.Context, updatedCard dto.PlaceCard) error {
	err := s.rep.ReorderCard(ctx, repositoryDto.PlaceCard{
		LinkCard:    updatedCard.LinkCard,
		LinkSection: updatedCard.LinkSection,
		Position:    updatedCard.Position,
	})
	if err != nil {
		return fmt.Errorf("rep.ReordredCard: %w", err)
	}

	return nil
}

func (s *Service) CreateCard(ctx context.Context, newCard dto.NewCard) (dto.PlaceCard, error) {
	linkCard := uuid.New()

	position, err := s.rep.CreateCard(ctx, repositoryDto.NewCard{
		LinkAuthor:   newCard.LinkAuthor,
		LinkCard:     linkCard,
		LinkSection:  newCard.LinkSection,
		Description:  newCard.Description,
		Title:        newCard.Title,
		LinkExecuter: newCard.LinkExecuter,
		DataDeadLine: newCard.DataDeadLine,
	})
	if err != nil {
		return dto.PlaceCard{}, fmt.Errorf("rep.CreateCard: %w", err)
	}

	return dto.PlaceCard{
		LinkCard:    linkCard,
		LinkSection: newCard.LinkSection,
		Position:    position,
	}, nil
}
