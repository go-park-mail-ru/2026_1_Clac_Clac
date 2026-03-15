package service

import (
	"context"
	"fmt"

	"github.com/go-park-mail-ru/2026_1_Clac_Clac/internal/board/models"
	"github.com/go-park-mail-ru/2026_1_Clac_Clac/internal/db"

	"github.com/google/uuid"
)

type BoardRepository interface {
	GetBoards(ctx context.Context, userID uuid.UUID) ([]models.Board, error)
	AddEmptyBoard(ctx context.Context, baord db.Board, userID uuid.UUID) error
}

type Service struct {
	rep BoardRepository
}

func NewBoardService(rep BoardRepository) *Service {
	return &Service{
		rep: rep,
	}
}

func (bs *Service) GetBoards(ctx context.Context, userID uuid.UUID) ([]models.Board, error) {
	boards, err := bs.rep.GetBoards(ctx, userID)
	if err != nil {
		return []models.Board{}, fmt.Errorf("rep.GetBoards: %w", err)
	}

	return boards, nil
}

func (bs *Service) CreateEmptyBoard(ctx context.Context, userID uuid.UUID) error {
	emptyBoard := db.Board{
		ID: uuid.New(),
	}

	err := bs.rep.AddEmptyBoard(ctx, emptyBoard, userID)
	if err != nil {
		return fmt.Errorf("rep.AddEmptyBoard: %w", err)
	}

	return nil
}
