package service

import (
	"context"
	"fmt"

	"github.com/go-park-mail-ru/2026_1_Clac_Clac/internal/board/models"

	"github.com/google/uuid"
)

type BoardRepository interface {
	GetBoards(ctx context.Context, userID uuid.UUID) ([]models.Board, error)
	AddEmptyBoard(ctx context.Context, board models.Board, userID uuid.UUID) error
}

type Service struct {
	rep BoardRepository
}

func NewService(rep BoardRepository) *Service {
	return &Service{
		rep: rep,
	}
}

func (s *Service) GetBoards(ctx context.Context, userID uuid.UUID) ([]models.Board, error) {
	boards, err := s.rep.GetBoards(ctx, userID)
	if err != nil {
		return []models.Board{}, fmt.Errorf("rep.GetBoards: %w", err)
	}

	return boards, nil
}

func (s *Service) CreateEmptyBoard(ctx context.Context, link uuid.UUID) error {
	emptyBoard := models.Board{
		Link: uuid.New(),
	}

	err := s.rep.AddEmptyBoard(ctx, emptyBoard, link)
	if err != nil {
		return fmt.Errorf("rep.AddEmptyBoard: %w", err)
	}

	return nil
}
