package service

import (
	"context"
	"fmt"

	"github.com/go-park-mail-ru/2026_1_Clac_Clac/board/internal/card/common"
	repositoryDto "github.com/go-park-mail-ru/2026_1_Clac_Clac/board/internal/card/repository/dto"
	"github.com/go-park-mail-ru/2026_1_Clac_Clac/board/internal/card/service/dto"
	"github.com/google/uuid"
)

//go:generate mockery --name=CardRepository --output=mock_card_rep
type CardRepository interface {
	GetCard(ctx context.Context, linkCard uuid.UUID) (repositoryDto.InfoCard, error)
	DeleteCard(ctx context.Context, linkCard uuid.UUID) error
	UpdateCardDetails(ctx context.Context, updatedCard repositoryDto.UpdatingCardDetails) error
	ReorderCard(ctx context.Context, updatingPlaceCard repositoryDto.PlaceCard) error
	CreateCard(ctx context.Context, newCard repositoryDto.NewCard) (int, error)
	GetComments(ctx context.Context, cardLink uuid.UUID) ([]repositoryDto.CommentInfo, error)
	CreateComment(ctx context.Context, createCardInfo repositoryDto.CreateCommentInfo) (repositoryDto.CommentInfo, error)
	IsCommentAuthor(ctx context.Context, commentLink uuid.UUID, userLink uuid.UUID) bool
	DeleteComment(ctx context.Context, commentLink uuid.UUID) error
	UpdateComment(ctx context.Context, updateCommentInfo repositoryDto.UpdateCommentInfo) error
}

type Service struct {
	rep CardRepository
}

func NewService(rep CardRepository) *Service {
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

func (s *Service) GetComments(ctx context.Context, cardLink uuid.UUID) ([]dto.CommentInfo, error) {
	comments, err := s.rep.GetComments(ctx, cardLink)
	if err != nil {
		return []dto.CommentInfo{}, fmt.Errorf("CardRepository.GetComments: %w", err)
	}

	commentsInfo := make([]dto.CommentInfo, 0)
	for _, comment := range comments {
		commentsInfo = append(commentsInfo, dto.CommentInfo{
			Link:       comment.Link,
			ParentLink: comment.ParentLink,
			AuthorLink: comment.AuthorLink,
			Text:       comment.Text,
		})
	}

	return commentsInfo, nil
}

func (s *Service) CreateComment(ctx context.Context, createCardInfo dto.CreateCommentInfo) (dto.CommentInfo, error) {
	comment, err := s.rep.CreateComment(ctx, repositoryDto.CreateCommentInfo{
		CardLink:   createCardInfo.CardLink,
		ParentLink: createCardInfo.ParentLink,
		AuthorLink: createCardInfo.AuthorLink,
		Text:       createCardInfo.Text,
	})
	if err != nil {
		return dto.CommentInfo{}, fmt.Errorf("CardRepository.CreateComment: %w", err)
	}

	return dto.CommentInfo{
		Link:       comment.Link,
		ParentLink: comment.ParentLink,
		AuthorLink: comment.AuthorLink,
		Text:       comment.Text,
	}, nil
}

func (s *Service) DeleteComment(ctx context.Context, commentLink uuid.UUID, userLink uuid.UUID) error {
	isCommentAuthor := s.rep.IsCommentAuthor(ctx, commentLink, userLink)
	if !isCommentAuthor {
		return common.ErrPermissionDenied
	}

	err := s.rep.DeleteComment(ctx, commentLink)
	if err != nil {
		return fmt.Errorf("CardService.DeleteCommend: %w", err)
	}

	return nil
}

func (s *Service) UpdateComment(ctx context.Context, updateCommentInfo dto.UpdateCommentInfo) error {
	isCommentAuthor := s.rep.IsCommentAuthor(ctx, updateCommentInfo.CommentLink, updateCommentInfo.UserLink)
	if !isCommentAuthor {
		return common.ErrPermissionDenied
	}

	err := s.rep.UpdateComment(ctx, repositoryDto.UpdateCommentInfo{
		CommentLink: updateCommentInfo.CommentLink,
		Text:        updateCommentInfo.Text,
	})
	if err != nil {
		return fmt.Errorf("CardRepository.UpdateComment: %w", err)
	}

	return nil
}
