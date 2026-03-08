package board

import (
	"context"
	"errors"

	"github.com/go-park-mail-ru/2026_1_Clac_Clac/internal/common"
	"github.com/go-park-mail-ru/2026_1_Clac_Clac/internal/models"
	dbConnection "github.com/go-park-mail-ru/2026_1_Clac_Clac/internal/repository/db_connection"
	"github.com/google/uuid"
)

var (
	ErrorNotExistingSession = errors.New("session not found or expired")
	ErrorSeesionExpired     = errors.New("time life session expired")
)

type BoardRepository struct {
	database *dbConnection.MapDatabases
}

func NewBoardRepository(db *dbConnection.MapDatabases) *BoardRepository {
	return &BoardRepository{
		database: db,
	}
}

func (br *BoardRepository) GetBoards(ctx context.Context, userID uuid.UUID) ([]models.Board, error) {
	br.database.MutexBoards.Lock()
	defer br.database.MutexBoards.Unlock()

	if user, exist := br.database.UsersDB[userID]; exist {
		return user.Boards, nil
	}

	return nil, common.ErrorNonexistentUser
}
