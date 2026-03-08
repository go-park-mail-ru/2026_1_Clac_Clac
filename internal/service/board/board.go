package board

import (
	"context"
	"fmt"

	"github.com/go-park-mail-ru/2026_1_Clac_Clac/internal/models"
	"github.com/google/uuid"
)

type BoardRepository interface {
	GetBoards(ctx context.Context, userID uuid.UUID) ([]models.Board, error)
}

type BoardService struct {
	rep BoardRepository
}

func NewBoardService(rep BoardRepository) *BoardService {
	return &BoardService{
		rep: rep,
	}
}

func (bs *BoardService) GetBoards(ctx context.Context, userID uuid.UUID) ([]models.Board, error) {
	boards, err := bs.rep.GetBoards(ctx, userID)
	if err != nil {
		return []models.Board{}, fmt.Errorf("rep.GetBoards: %w", err)
	}

	return boards, nil
}
