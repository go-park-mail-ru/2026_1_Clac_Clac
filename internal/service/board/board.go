package board

import (
	"context"

	"github.com/go-park-mail-ru/2026_1_Clac_Clac/internal/models"
	"github.com/google/uuid"
)

type BoardRepositpry interface {
	GetBoards(ctx context.Context, userID uuid.UUID) []models.Board
}

type BoardService struct {
	rep BoardRepositpry
}

func NewBoardService(rep BoardRepositpry) *BoardService {
	return &BoardService{
		rep: rep,
	}
}

func (bs *BoardService) GetBoards(ctx context.Context, userID uuid.UUID) []models.Board {
	return bs.rep.GetBoards(ctx, userID)
}
