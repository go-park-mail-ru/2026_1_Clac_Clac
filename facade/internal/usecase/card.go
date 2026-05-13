package usecase

import (
	"context"
	"fmt"

	"github.com/go-park-mail-ru/2026_1_Clac_Clac/facade/internal/domain"
)

type CardClient interface {
	GetCard(ctx context.Context, infoCard domain.GetCardRequest) (domain.CardFullInfo, error)
	DeleteCard(ctx context.Context, infoCard domain.DeleteCardRequest) error
	UpdateCard(ctx context.Context, infoCard domain.UpdateCardRequest) error
	ReorderCards(ctx context.Context, infoCard domain.ReorderCardsRequest) error
	CreateCard(ctx context.Context, infoCard domain.CreateCardRequest) (domain.CreateCardResponse, error)
	GetComments(ctx context.Context, infoComments domain.GetCommentsRequest) (domain.GetCommentsResponse, error)
	CreateComment(ctx context.Context, infoComment domain.CreateCommentRequest) (domain.CreateCommentResponse, error)
	DeleteComment(ctx context.Context, infoComment domain.DeleteCommentRequest) error
	UpdateComment(ctx context.Context, infoComment domain.UpdateCommentRequest) error
	CreateSubtask(ctx context.Context, infoSubtask domain.CreateSubtaskRequest) (domain.SubtaskInfo, error)
	UpdateSubtask(ctx context.Context, infoSubtask domain.UpdateSubtaskRequest) error
	DeleteSubtask(ctx context.Context, infoSubtask domain.DeleteSubtask) error
}

type Card struct {
	card CardClient
}

func NewCard(card CardClient) *Card {
	return &Card{
		card: card,
	}
}

func (c *Card) GetCard(ctx context.Context, infoCard domain.GetCardRequest) (domain.CardFullInfo, error) {
	cardInfo, err := c.card.GetCard(ctx, infoCard)
	if err != nil {
		return domain.CardFullInfo{}, fmt.Errorf("card.GetCard: %w", err)
	}

	return cardInfo, nil
}

func (c *Card) DeleteCard(ctx context.Context, infoCard domain.DeleteCardRequest) error {
	err := c.card.DeleteCard(ctx, infoCard)
	if err != nil {
		return fmt.Errorf("card.DeleteCard: %w", err)
	}

	return nil
}

func (c *Card) UpdateCard(ctx context.Context, infoCard domain.UpdateCardRequest) error {
	err := c.card.UpdateCard(ctx, infoCard)
	if err != nil {
		return fmt.Errorf("card.UpdateCard: %w", err)
	}

	return nil
}

func (c *Card) ReorderCards(ctx context.Context, infoCard domain.ReorderCardsRequest) error {
	err := c.card.ReorderCards(ctx, infoCard)
	if err != nil {
		return fmt.Errorf("card.ReorderCards: %w", err)
	}

	return nil
}

func (c *Card) CreateCard(ctx context.Context, infoCard domain.CreateCardRequest) (domain.CreateCardResponse, error) {
	createdCard, err := c.card.CreateCard(ctx, infoCard)
	if err != nil {
		return domain.CreateCardResponse{}, fmt.Errorf("card.CreateCard: %w", err)
	}

	return createdCard, nil
}

func (c *Card) GetComments(ctx context.Context, infoComments domain.GetCommentsRequest) (domain.GetCommentsResponse, error) {
	comments, err := c.card.GetComments(ctx, infoComments)
	if err != nil {
		return domain.GetCommentsResponse{}, fmt.Errorf("card.GetComments: %w", err)
	}

	return comments, nil
}

func (c *Card) CreateComment(ctx context.Context, infoComment domain.CreateCommentRequest) (domain.CreateCommentResponse, error) {
	comment, err := c.card.CreateComment(ctx, infoComment)
	if err != nil {
		return domain.CreateCommentResponse{}, fmt.Errorf("card.CreateComment: %w", err)
	}

	return comment, nil
}

func (c *Card) DeleteComment(ctx context.Context, infoComment domain.DeleteCommentRequest) error {
	err := c.card.DeleteComment(ctx, infoComment)
	if err != nil {
		return fmt.Errorf("card.DeleteComment: %w", err)
	}

	return nil
}

func (c *Card) UpdateComment(ctx context.Context, infoComment domain.UpdateCommentRequest) error {
	err := c.card.UpdateComment(ctx, infoComment)
	if err != nil {
		return fmt.Errorf("card.UpdateComment: %w", err)
	}

	return nil
}

func (c *Card) CreateSubtask(ctx context.Context, infoSubtask domain.CreateSubtaskRequest) (domain.SubtaskInfo, error) {
	subtask, err := c.card.CreateSubtask(ctx, infoSubtask)
	if err != nil {
		return domain.SubtaskInfo{}, fmt.Errorf("card.CreateSubtask: %w", err)
	}

	return subtask, nil
}

func (c *Card) UpdateSubtask(ctx context.Context, infoSubtask domain.UpdateSubtaskRequest) error {
	err := c.card.UpdateSubtask(ctx, infoSubtask)
	if err != nil {
		return fmt.Errorf("card.UpdateSubtask: %w", err)
	}

	return nil
}

func (c *Card) DeleteSubtask(ctx context.Context, infoSubtask domain.DeleteSubtask) error {
	err := c.card.DeleteSubtask(ctx, infoSubtask)
	if err != nil {
		return fmt.Errorf("card.DeleteSubtask: %w", err)
	}

	return nil
}
